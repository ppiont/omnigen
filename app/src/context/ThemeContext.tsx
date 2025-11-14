/* eslint-disable react-refresh/only-export-components */
import type { PropsWithChildren } from 'react'
import { createContext, useContext, useEffect, useMemo, useState } from 'react'

type Theme = 'dark' | 'light'

type ThemeContextValue = {
  theme: Theme
  toggleTheme: () => void
}

const ThemeContext = createContext<ThemeContextValue | undefined>(undefined)

const THEME_STORAGE_KEY = 'omnigen-theme'

const getInitialTheme = (): Theme => {
  if (typeof window === 'undefined') {
    return 'dark'
  }
  const stored = window.localStorage.getItem(THEME_STORAGE_KEY) as Theme | null
  if (stored && ['dark', 'light'].includes(stored)) {
    return stored
  }
  return window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark'
}

export const ThemeProvider = ({ children }: PropsWithChildren) => {
  const [theme, setTheme] = useState<Theme>(getInitialTheme)

  useEffect(() => {
    if (typeof window === 'undefined') return
    document.documentElement.classList.toggle('dark', theme === 'dark')
    window.localStorage.setItem(THEME_STORAGE_KEY, theme)
  }, [theme])

  const value = useMemo<ThemeContextValue>(
    () => ({
      theme,
      toggleTheme: () => setTheme((prev) => (prev === 'dark' ? 'light' : 'dark')),
    }),
    [theme],
  )

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
}

export const useTheme = () => {
  const ctx = useContext(ThemeContext)
  if (!ctx) {
    throw new Error('useTheme must be used within ThemeProvider')
  }
  return ctx
}

