import { useState, useCallback } from 'preact/hooks'
import { SearchByAddress } from '../wailsjs/go/main/App'
import SearchBox from './components/SearchBox'
import AddressResultTable from './components/AddressResultTable'
import ThemeToggle from './components/ThemeToggle'

export function App() {
  const [keyword, setKeyword] = useState('')
  const [addressResult, setAddressResult] = useState(null)

  const handleSearch = useCallback(async (kw) => {
    setKeyword(kw)
    if (!kw.trim()) {
      setAddressResult(null)
      return
    }

    try {
      const response = await SearchByAddress(kw)
      setAddressResult(response)
    } catch (err) {
      console.error('Search error:', err)
      setAddressResult(null)
    }
  }, [])

  return (
    <div class="min-h-screen bg-base-200 p-4">
      <div class="max-w-6xl mx-auto">
        {/* Header */}
        <div class="flex justify-between items-center mb-4">
          <h1 class="text-2xl font-bold">台灣郵遞區號查詢</h1>
          <ThemeToggle />
        </div>

        {/* Search area */}
        <div class="card bg-base-100 shadow-xl mb-4">
          <div class="card-body">
            <SearchBox
              onSearch={handleSearch}
              placeholder="輸入地址查詢郵遞區號（如：基隆路一段172巷1號）"
            />

            {/* Parsed address info */}
            {addressResult?.parsedAddress && (
              <div class="text-sm text-info mt-2">
                解析結果：{addressResult.parsedAddress}
              </div>
            )}

            {/* Best match result */}
            {addressResult?.bestMatch && (
              <div class="alert alert-success mt-2">
                <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <div>
                  <div class="font-bold text-lg">郵遞區號：{addressResult.bestMatch.zipcode}</div>
                  <div class="text-sm">
                    {addressResult.bestMatch.city} {addressResult.bestMatch.district} {addressResult.bestMatch.street}
                    {addressResult.bestMatch.scope && ` (${addressResult.bestMatch.scope})`}
                  </div>
                </div>
              </div>
            )}

            {/* No match warning */}
            {addressResult && !addressResult.bestMatch && addressResult.results?.length > 0 && (
              <div class="alert alert-warning mt-2">
                <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
                <span>找到相關路名但無法精確匹配門牌範圍，請參考以下結果</span>
              </div>
            )}
          </div>
        </div>

        {/* Results table */}
        <div class="card bg-base-100 shadow-xl">
          <div class="card-body p-0">
            <AddressResultTable
              results={addressResult?.results || []}
              keyword={keyword}
            />
          </div>
        </div>
      </div>
    </div>
  )
}
