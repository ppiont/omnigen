import { useState, useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { LayoutGrid, Sparkles, Video, Settings, Menu } from "lucide-react";
import Sidebar from "./Sidebar.jsx";
import "../styles/dashboard.css";

const sidebarTabs = [
  {
    id: "dashboard",
    label: "Dashboard",
    description: "Overview",
    icon: <LayoutGrid size={20} />,
  },
  {
    id: "create",
    label: "Create",
    description: "Generate video ads",
    icon: <Sparkles size={20} />,
  },
  {
    id: "videos",
    label: "Videos",
    description: "Browse library",
    icon: <Video size={20} />,
  },
  {
    id: "settings",
    label: "Settings",
    description: "Configure options",
    icon: <Settings size={20} />,
  },
];


function AppLayout({ children }) {
  const location = useLocation();
  const navigate = useNavigate();
  const [isDrawerOpen, setIsDrawerOpen] = useState(false);
  const [isCollapsed, setIsCollapsed] = useState(false);

  const closeDrawer = () => setIsDrawerOpen(false);

  useEffect(() => {
    if (!isDrawerOpen) {
      return undefined;
    }

    const handleEscape = (e) => {
      if (e.key === "Escape") {
        closeDrawer();
      }
    };

    document.addEventListener("keydown", handleEscape);
    return () => document.removeEventListener("keydown", handleEscape);
  }, [isDrawerOpen]);

  const getActiveTab = () => {
    const path = location.pathname;
    if (path === "/dashboard") return "dashboard";
    if (path === "/create") return "create";
    if (path === "/videos") return "videos";
    if (path === "/settings") return "settings";
    return "dashboard";
  };

  const handleTabSelect = (tabId) => {
    const routes = {
      dashboard: "/dashboard",
      create: "/create",
      videos: "/videos",
      settings: "/settings",
    };
    navigate(routes[tabId]);
    closeDrawer();
  };

  return (
    <main className="dashboard-page">
      <div className="dashboard-shell">
        <Sidebar
          tabs={sidebarTabs}
          activeTab={getActiveTab()}
          onSelect={handleTabSelect}
          isDrawerOpen={isDrawerOpen}
          onClose={closeDrawer}
          isCollapsed={isCollapsed}
          onToggleCollapse={() => setIsCollapsed(!isCollapsed)}
        />

        <section className="dashboard-main">
          <div className="mobile-header">
            <button
              className="drawer-toggle"
              aria-label="Open navigation"
              onClick={() => setIsDrawerOpen(true)}
            >
              <Menu size={24} />
            </button>
            <h1>OMNIGEN</h1>
          </div>

          <div className="dashboard-content">{children}</div>
        </section>
      </div>

      <div
        className={`sidebar-overlay ${isDrawerOpen ? "is-visible" : ""}`}
        role="presentation"
        aria-hidden={!isDrawerOpen}
        onClick={closeDrawer}
      />
    </main>
  );
}

export default AppLayout;
