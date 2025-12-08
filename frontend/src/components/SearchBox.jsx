import { useState, useEffect, useRef } from 'preact/hooks'

export default function SearchBox({ onSearch, placeholder }) {
  const [value, setValue] = useState('')
  const inputRef = useRef(null)
  const debounceRef = useRef(null)

  // Debounced search
  useEffect(() => {
    if (debounceRef.current) {
      clearTimeout(debounceRef.current)
    }

    debounceRef.current = setTimeout(() => {
      onSearch(value)
    }, 300)

    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current)
      }
    }
  }, [value, onSearch])

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e) => {
      // Ctrl+F or Cmd+F to focus search
      if ((e.ctrlKey || e.metaKey) && e.key === 'f') {
        e.preventDefault()
        inputRef.current?.focus()
      }
      // Escape to clear
      if (e.key === 'Escape') {
        setValue('')
        inputRef.current?.blur()
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [])

  const handleClear = () => {
    setValue('')
    inputRef.current?.focus()
  }

  return (
    <div class="form-control">
      <div class="relative">
        {/* Search icon on the left */}
        <span class="absolute left-3 top-1/2 -translate-y-1/2 text-base-content/50">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="h-5 w-5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
            />
          </svg>
        </span>
        <input
          ref={inputRef}
          type="text"
          placeholder={placeholder || "輸入關鍵字搜尋（縣市、區域、郵遞區號、路名...）"}
          class="input input-bordered w-full pl-10 pr-10"
          value={value}
          onInput={(e) => setValue(e.target.value)}
        />
        {/* Clear button on the right */}
        {value && (
          <button
            class="absolute right-2 top-1/2 -translate-y-1/2 btn btn-ghost btn-circle btn-sm"
            onClick={handleClear}
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              class="h-4 w-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        )}
      </div>
      <label class="label">
        <span class="label-text-alt text-base-content/50">
          提示：Ctrl+F 聚焦搜尋框、Esc 清除
        </span>
      </label>
    </div>
  )
}
