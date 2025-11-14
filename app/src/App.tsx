import { AnimatePresence, motion } from 'framer-motion'
import { Navigate, Route, Routes, useLocation } from 'react-router-dom'
import { ThemeProvider, useTheme } from './context/ThemeContext'
import { DashboardPage } from './pages/Dashboard'

const AppRoutes = () => {
  const location = useLocation()
  const { theme } = useTheme()

  return (
    <div
      className={`min-h-screen transition-colors ${
        theme === 'light' ? 'bg-light-bg text-light-text' : 'bg-background text-text-primary'
      }`}
    >
      <AnimatePresence mode="wait">
        <Routes location={location} key={location.pathname}>
          <Route
            path="/dashboard"
            element={
              <motion.div initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, y: -16 }}>
                <DashboardPage />
              </motion.div>
            }
          />
          <Route path="*" element={<Navigate to="/dashboard" replace />} />
        </Routes>
      </AnimatePresence>
    </div>
  )
}

function App() {
  return (
    <ThemeProvider>
      <AppRoutes />
    </ThemeProvider>
  )
}

export default App
