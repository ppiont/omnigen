import { useContext } from "react";
import { ThemeContext } from "./themeContext.js";

/**
 * Hook for accessing theme context.
 *
 * @returns {{theme: string, toggleTheme: () => void}} Theme context value
 */
export function useTheme() {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error("useTheme must be used within ThemeProvider");
  }
  return context;
}
