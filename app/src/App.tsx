import { Routes, Route } from 'react-router-dom'
import { ThemeProvider } from './context/ThemeContext'
import Navbar from './components/Navbar'
import Landing from './pages/Landing'
import Login from './pages/Login'
import Signup from './pages/Signup'
import { DashboardPage } from './pages/Dashboard'

function App() {
  return (
    <ThemeProvider>
      <div className="app-shell">
        <Navbar />
        <Routes>
          <Route path="/" element={<Landing />} />
          <Route path="/login" element={<Login />} />
          <Route path="/signup" element={<Signup />} />
          <Route path="/dashboard" element={<DashboardPage />} />
        </Routes>
      </div>
    </ThemeProvider>
  )
}

export default App
