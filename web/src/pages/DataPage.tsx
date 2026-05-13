import { useEffect, useMemo, useState } from 'react'
import { toast } from 'sonner'
import { Activity, Database, ExternalLink, KeyRound, RefreshCw, Save, TrendingUp } from 'lucide-react'
import { api } from '../lib/api'
import type { AlchemyTokenPrice } from '../types'

const ALCHEMY_KEY_STORAGE = 'nofx_alchemy_api_key'
const ALCHEMY_SYMBOLS_STORAGE = 'nofx_alchemy_symbols'
const DEFAULT_SYMBOLS = 'BTC,ETH,SOL,USDC,LINK'

function parseSymbols(input: string): string[] {
  return input
    .split(',')
    .map((symbol) => symbol.trim().toUpperCase())
    .filter(Boolean)
    .slice(0, 25)
}

function formatPrice(price: AlchemyTokenPrice): string {
  if (!price.value) return '—'
  const numeric = Number(price.value)
  if (!Number.isFinite(numeric)) return price.value
  if (numeric >= 1000) return `$${numeric.toLocaleString(undefined, { maximumFractionDigits: 2 })}`
  if (numeric >= 1) return `$${numeric.toLocaleString(undefined, { maximumFractionDigits: 4 })}`
  return `$${numeric.toLocaleString(undefined, { maximumFractionDigits: 8 })}`
}

export function DataPage() {
  const [apiKey, setApiKey] = useState('')
  const [symbolsInput, setSymbolsInput] = useState(DEFAULT_SYMBOLS)
  const [prices, setPrices] = useState<AlchemyTokenPrice[]>([])
  const [loading, setLoading] = useState(false)
  const [lastUpdated, setLastUpdated] = useState<string | null>(null)

  const symbols = useMemo(() => parseSymbols(symbolsInput), [symbolsInput])
  const hasConfig = apiKey.trim().length > 0 && symbols.length > 0

  useEffect(() => {
    setApiKey(localStorage.getItem(ALCHEMY_KEY_STORAGE) || '')
    setSymbolsInput(localStorage.getItem(ALCHEMY_SYMBOLS_STORAGE) || DEFAULT_SYMBOLS)
  }, [])

  const saveConfig = () => {
    localStorage.setItem(ALCHEMY_KEY_STORAGE, apiKey.trim())
    localStorage.setItem(ALCHEMY_SYMBOLS_STORAGE, symbolsInput.trim() || DEFAULT_SYMBOLS)
    toast.success('Alchemy market data config saved locally')
  }

  const fetchPrices = async () => {
    if (!apiKey.trim()) {
      toast.error('Add your Alchemy API key first')
      return
    }
    if (symbols.length === 0) {
      toast.error('Add at least one token symbol')
      return
    }
    setLoading(true)
    try {
      const response = await api.getAlchemyTokenPrices(apiKey.trim(), symbols)
      setPrices(response.prices || [])
      setLastUpdated(new Date().toLocaleString())
      localStorage.setItem(ALCHEMY_KEY_STORAGE, apiKey.trim())
      localStorage.setItem(ALCHEMY_SYMBOLS_STORAGE, symbolsInput.trim() || DEFAULT_SYMBOLS)
      toast.success('Alchemy prices updated')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to fetch Alchemy prices')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-[calc(100vh-64px)] bg-[#0B0E11] text-white px-4 py-8">
      <div className="mx-auto max-w-6xl space-y-6">
        <section className="rounded-3xl border border-zinc-800 bg-gradient-to-br from-zinc-950 via-zinc-900 to-black p-6 shadow-2xl">
          <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
            <div className="space-y-3">
              <div className="inline-flex items-center gap-2 rounded-full border border-blue-400/20 bg-blue-400/10 px-3 py-1 text-xs font-medium text-blue-300">
                <Database size={14} /> Alchemy Market Data
              </div>
              <div>
                <h1 className="text-3xl font-bold tracking-tight">On-chain prices for AI trading context</h1>
                <p className="mt-2 max-w-2xl text-sm text-zinc-400">
                  Connect an Alchemy API key and pull live token prices through the NOFX backend proxy. This is read-only market data — no wallet signing or trades happen here.
                </p>
              </div>
            </div>
            <a
              href="https://dashboard.alchemy.com/signup"
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-2 rounded-xl border border-zinc-700 bg-zinc-900 px-4 py-2 text-sm text-zinc-200 transition hover:border-blue-400/50 hover:text-blue-300"
            >
              Get Alchemy key <ExternalLink size={14} />
            </a>
          </div>

          <div className="mt-6 grid gap-4 lg:grid-cols-[1.3fr_0.7fr]">
            <div className="rounded-2xl border border-zinc-800 bg-black/30 p-4">
              <label className="mb-2 flex items-center gap-2 text-sm font-semibold text-zinc-200">
                <KeyRound size={16} className="text-blue-300" /> Alchemy API Key
              </label>
              <input
                type="password"
                value={apiKey}
                onChange={(event) => setApiKey(event.target.value)}
                placeholder="Paste your Alchemy API key"
                className="w-full rounded-xl border border-zinc-800 bg-[#0B0E11] px-4 py-3 text-sm text-white outline-none transition placeholder:text-zinc-600 focus:border-blue-400/60"
              />
              <p className="mt-2 text-xs text-zinc-500">
                Stored in this browser only. The backend uses it per request and does not persist it.
              </p>
            </div>

            <div className="rounded-2xl border border-zinc-800 bg-black/30 p-4">
              <label className="mb-2 flex items-center gap-2 text-sm font-semibold text-zinc-200">
                <Activity size={16} className="text-emerald-300" /> Symbols
              </label>
              <input
                type="text"
                value={symbolsInput}
                onChange={(event) => setSymbolsInput(event.target.value)}
                placeholder="BTC,ETH,SOL"
                className="w-full rounded-xl border border-zinc-800 bg-[#0B0E11] px-4 py-3 text-sm text-white outline-none transition placeholder:text-zinc-600 focus:border-emerald-400/60"
              />
              <p className="mt-2 text-xs text-zinc-500">Comma-separated, max 25 symbols.</p>
            </div>
          </div>

          <div className="mt-5 flex flex-wrap items-center gap-3">
            <button
              type="button"
              onClick={fetchPrices}
              disabled={loading || !hasConfig}
              className="inline-flex items-center gap-2 rounded-xl bg-blue-500 px-4 py-2 text-sm font-semibold text-white transition hover:bg-blue-400 disabled:cursor-not-allowed disabled:bg-zinc-700 disabled:text-zinc-400"
            >
              <RefreshCw size={15} className={loading ? 'animate-spin' : ''} />
              {loading ? 'Fetching…' : 'Fetch prices'}
            </button>
            <button
              type="button"
              onClick={saveConfig}
              className="inline-flex items-center gap-2 rounded-xl border border-zinc-700 bg-zinc-900 px-4 py-2 text-sm font-semibold text-zinc-200 transition hover:border-zinc-500"
            >
              <Save size={15} /> Save locally
            </button>
            {lastUpdated && <span className="text-xs text-zinc-500">Last updated {lastUpdated}</span>}
          </div>
        </section>

        <section className="rounded-3xl border border-zinc-800 bg-zinc-950/70 p-6">
          <div className="mb-4 flex items-center justify-between">
            <div>
              <h2 className="text-lg font-semibold">Live token prices</h2>
              <p className="text-sm text-zinc-500">Powered by Alchemy Prices API</p>
            </div>
            <TrendingUp className="text-emerald-300" size={22} />
          </div>

          {prices.length === 0 ? (
            <div className="rounded-2xl border border-dashed border-zinc-800 bg-black/20 p-8 text-center text-sm text-zinc-500">
              Add your Alchemy key and fetch prices to see market data here.
            </div>
          ) : (
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {prices.map((price, index) => (
                <div key={`${price.symbol}-${price.currency || index}`} className="rounded-2xl border border-zinc-800 bg-black/30 p-4">
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <div className="text-sm text-zinc-500">{price.currency || 'USD'}</div>
                      <div className="text-2xl font-bold">{price.symbol}</div>
                    </div>
                    <div className="rounded-full bg-emerald-400/10 px-3 py-1 text-xs text-emerald-300">Alchemy</div>
                  </div>
                  {price.error ? (
                    <div className="mt-4 text-sm text-red-300">{price.error}</div>
                  ) : (
                    <>
                      <div className="mt-4 text-3xl font-semibold tracking-tight">{formatPrice(price)}</div>
                      {price.last_updated && <div className="mt-2 text-xs text-zinc-600">{price.last_updated}</div>}
                    </>
                  )}
                </div>
              ))}
            </div>
          )}
        </section>
      </div>
    </div>
  )
}
