import { Routes, Route, useLocation } from "react-router-dom";
import { ThemeProvider } from "./context/ThemeContext.jsx";
import Navbar from "./components/Navbar.jsx";
import Landing from "./pages/Landing.jsx";
import Login from "./pages/Login.jsx";
import Signup from "./pages/Signup.jsx";
import DashboardOverview from "./pages/DashboardOverview.jsx";
import Create from "./pages/Create.jsx";
import Videos from "./pages/Videos.jsx";
import Settings from "./pages/Settings.jsx";

function App() {
  const location = useLocation();
  const appRoutes = ["/dashboard", "/create", "/videos", "/settings"];
  const showNavbar = !appRoutes.includes(location.pathname);

  return (
    <ThemeProvider>
      <div className="app-shell">
        {showNavbar && <Navbar />}
        <Routes>
          <Route path="/" element={<Landing />} />
          <Route path="/login" element={<Login />} />
          <Route path="/signup" element={<Signup />} />
          <Route path="/dashboard" element={<DashboardOverview />} />
          <Route path="/create" element={<Create />} />
          <Route path="/videos" element={<Videos />} />
          <Route path="/settings" element={<Settings />} />
        </Routes>
      </div>
    </ThemeProvider>
  );
}

export default App;
