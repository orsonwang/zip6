import { useMemo } from 'preact/hooks'

export default function AddressResultTable({ results, keyword }) {
  const isEmpty = useMemo(() => !results || results.length === 0, [results])

  if (isEmpty) {
    return (
      <div class="p-8 text-center text-base-content/50">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          class="h-16 w-16 mx-auto mb-4 opacity-50"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"
          />
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"
          />
        </svg>
        <p>輸入完整地址查詢郵遞區號</p>
        <p class="text-sm mt-2">例如：台北市信義區基隆路一段172巷1號3樓</p>
      </div>
    )
  }

  // Count matched results
  const matchedCount = results.filter(r => r.matched).length

  return (
    <div class="table-container">
      {matchedCount > 0 && (
        <div class="p-2 bg-success/10 text-success text-sm">
          找到 {matchedCount} 筆符合的郵遞區號
        </div>
      )}
      <table class="table table-zebra table-fixed-header w-full">
        <thead class="bg-base-200">
          <tr>
            <th class="w-12"></th>
            <th class="w-28">郵遞區號</th>
            <th class="w-24">縣市</th>
            <th class="w-24">區域</th>
            <th>路名</th>
            <th>投遞範圍</th>
          </tr>
        </thead>
        <tbody>
          {results.map((item, idx) => (
            <tr
              key={idx}
              class={`hover ${item.matched ? 'bg-success/20' : ''}`}
            >
              <td>
                {item.matched && (
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    class="h-5 w-5 text-success"
                    viewBox="0 0 20 20"
                    fill="currentColor"
                  >
                    <path
                      fill-rule="evenodd"
                      d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                      clip-rule="evenodd"
                    />
                  </svg>
                )}
              </td>
              <td class={`font-mono font-semibold ${item.matched ? 'text-success' : 'text-primary'}`}>
                {item.zipcode}
              </td>
              <td>{item.city}</td>
              <td>{item.district}</td>
              <td>{item.street}</td>
              <td class="text-sm text-base-content/70">{item.scope}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
