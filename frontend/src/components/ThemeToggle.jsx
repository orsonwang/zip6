import { useState, useEffect } from 'preact/hooks'

const THEMES = ['light', 'dark', 'nord', 'cupcake']
const STORAGE_KEY = 'zip6-theme'

export default function ThemeToggle() {
  const [theme, setTheme] = useState('light')

  // Load saved theme on mount
  useEffect(() => {
    const savedTheme = localStorage.getItem(STORAGE_KEY)
    if (savedTheme && THEMES.includes(savedTheme)) {
      setTheme(savedTheme)
      document.documentElement.setAttribute('data-theme', savedTheme)
    } else {
      // Check system preference
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
      const defaultTheme = prefersDark ? 'dark' : 'light'
      setTheme(defaultTheme)
      document.documentElement.setAttribute('data-theme', defaultTheme)
    }
  }, [])

  const handleThemeChange = (newTheme) => {
    setTheme(newTheme)
    document.documentElement.setAttribute('data-theme', newTheme)
    localStorage.setItem(STORAGE_KEY, newTheme)
  }

  const isDark = theme === 'dark'

  return (
    <div class="dropdown dropdown-end">
      <label tabIndex={0} class="btn btn-ghost btn-circle">
        {isDark ? (
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
              d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"
            />
          </svg>
        ) : (
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
              d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"
            />
          </svg>
        )}
      </label>
      <ul
        tabIndex={0}
        class="dropdown-content z-[1] menu p-2 shadow-lg bg-base-100 rounded-box w-36"
      >
        {THEMES.map((t) => (
          <li key={t}>
            <button
              class={theme === t ? 'active' : ''}
              onClick={() => handleThemeChange(t)}
            >
              {t === 'light' && 'Light'}
              {t === 'dark' && 'Dark'}
              {t === 'nord' && 'Nord'}
              {t === 'cupcake' && 'Cupcake'}
            </button>
          </li>
        ))}
      </ul>
    </div>
  )
}
