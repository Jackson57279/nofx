package api

import "testing"

func TestNormalizeAlchemySymbols(t *testing.T) {
	symbols, err := normalizeAlchemySymbols([]string{" eth ", "BTC", "eth", "sol"})
	if err != nil {
		t.Fatalf("expected valid symbols: %v", err)
	}
	expected := []string{"ETH", "BTC", "SOL"}
	if len(symbols) != len(expected) {
		t.Fatalf("expected %d symbols, got %d", len(expected), len(symbols))
	}
	for i := range expected {
		if symbols[i] != expected[i] {
			t.Fatalf("expected %v, got %v", expected, symbols)
		}
	}
}

func TestNormalizeAlchemySymbolsRejectsInvalid(t *testing.T) {
	if _, err := normalizeAlchemySymbols([]string{"ETH", "BAD/SYMBOL"}); err == nil {
		t.Fatal("expected invalid symbol error")
	}
}

func TestBuildAlchemyTokenPricesURL(t *testing.T) {
	got, err := buildAlchemyTokenPricesURL("test key", []string{"ETH", "BTC"})
	if err != nil {
		t.Fatalf("expected url: %v", err)
	}
	if got != "https://api.g.alchemy.com/prices/v1/test%20key/tokens/by-symbol?symbols=%5BETH%2CBTC%5D" {
		t.Fatalf("unexpected url: %s", got)
	}
}

func TestFlattenAlchemyPrices(t *testing.T) {
	body := []byte(`{"data":[{"symbol":"ETH","prices":[{"currency":"usd","value":"3200.12","lastUpdatedAt":"2026-01-01T00:00:00Z"}]}]}`)
	prices := flattenAlchemyPrices(body)
	if len(prices) != 1 {
		t.Fatalf("expected one price, got %d", len(prices))
	}
	if prices[0].Symbol != "ETH" || prices[0].Currency != "USD" || prices[0].Value != "3200.12" {
		t.Fatalf("unexpected price: %+v", prices[0])
	}
}
