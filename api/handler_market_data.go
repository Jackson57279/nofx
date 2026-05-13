package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"nofx/security"

	"github.com/gin-gonic/gin"
)

const alchemyTokenPricesBaseURL = "https://api.g.alchemy.com/prices/v1"

var alchemySymbolPattern = regexp.MustCompile(`^[A-Za-z0-9.$_-]{1,24}$`)

type alchemyTokenPricesRequest struct {
	APIKey  string   `json:"api_key"`
	Symbols []string `json:"symbols"`
}

type alchemyTokenPrice struct {
	Symbol      string `json:"symbol"`
	Currency    string `json:"currency,omitempty"`
	Value       string `json:"value,omitempty"`
	LastUpdated string `json:"last_updated,omitempty"`
	Error       string `json:"error,omitempty"`
}

type alchemyTokenPricesResponse struct {
	Provider string              `json:"provider"`
	Prices   []alchemyTokenPrice `json:"prices"`
	Raw      json.RawMessage     `json:"raw,omitempty"`
}

type alchemyPricesAPIResponse struct {
	Data []struct {
		Symbol string `json:"symbol"`
		Prices []struct {
			Currency      string `json:"currency"`
			Value         string `json:"value"`
			LastUpdatedAt string `json:"lastUpdatedAt"`
		} `json:"prices"`
		Error any `json:"error"`
	} `json:"data"`
}

// handleAlchemyTokenPrices proxies Alchemy's Token Prices API so the frontend
// can use Alchemy market data without browser CORS issues. The API key is only
// used for this request and is never logged or persisted by the server.
func (s *Server) handleAlchemyTokenPrices(c *gin.Context) {
	var req alchemyTokenPricesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	apiKey := strings.TrimSpace(req.APIKey)
	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Alchemy API key is required"})
		return
	}

	symbols, err := normalizeAlchemySymbols(req.Symbols)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	endpoint, err := buildAlchemyTokenPricesURL(apiKey, symbols)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Alchemy request"})
		return
	}

	httpClient := security.SafeHTTPClient(20 * time.Second)
	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, endpoint, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build Alchemy request"})
		return
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", "nofx-alchemy-market-data/1.0")

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("Alchemy request failed: %v", err)})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to read Alchemy response"})
		return
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("Alchemy API error (%d): %s", resp.StatusCode, string(body))})
		return
	}

	prices := flattenAlchemyPrices(body)
	c.JSON(http.StatusOK, alchemyTokenPricesResponse{
		Provider: "alchemy",
		Prices:   prices,
		Raw:      json.RawMessage(body),
	})
}

func normalizeAlchemySymbols(input []string) ([]string, error) {
	seen := make(map[string]bool)
	symbols := make([]string, 0, len(input))
	for _, raw := range input {
		symbol := strings.ToUpper(strings.TrimSpace(raw))
		if symbol == "" {
			continue
		}
		if !alchemySymbolPattern.MatchString(symbol) {
			return nil, fmt.Errorf("invalid token symbol: %s", raw)
		}
		if !seen[symbol] {
			seen[symbol] = true
			symbols = append(symbols, symbol)
		}
	}
	if len(symbols) == 0 {
		return nil, fmt.Errorf("at least one token symbol is required")
	}
	if len(symbols) > 25 {
		return nil, fmt.Errorf("Alchemy supports up to 25 symbols per request")
	}
	return symbols, nil
}

func buildAlchemyTokenPricesURL(apiKey string, symbols []string) (string, error) {
	base, err := url.Parse(alchemyTokenPricesBaseURL + "/" + url.PathEscape(apiKey) + "/tokens/by-symbol")
	if err != nil {
		return "", err
	}
	query := base.Query()
	query.Set("symbols", "["+strings.Join(symbols, ",")+"]")
	base.RawQuery = query.Encode()
	return base.String(), nil
}

func flattenAlchemyPrices(body []byte) []alchemyTokenPrice {
	var parsed alchemyPricesAPIResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil
	}
	prices := make([]alchemyTokenPrice, 0)
	for _, token := range parsed.Data {
		if token.Error != nil {
			prices = append(prices, alchemyTokenPrice{Symbol: token.Symbol, Error: fmt.Sprint(token.Error)})
			continue
		}
		if len(token.Prices) == 0 {
			prices = append(prices, alchemyTokenPrice{Symbol: token.Symbol, Error: "no price returned"})
			continue
		}
		for _, price := range token.Prices {
			prices = append(prices, alchemyTokenPrice{
				Symbol:      token.Symbol,
				Currency:    strings.ToUpper(price.Currency),
				Value:       price.Value,
				LastUpdated: price.LastUpdatedAt,
			})
		}
	}
	return prices
}
