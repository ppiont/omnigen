import { useState, useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import Sidebar from "./Sidebar.jsx";
import "../styles/dashboard.css";

const sidebarTabs = [
  {
    id: "dashboard",
    label: "Dashboard",
    description: "Overview",
    icon: (
      <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
        <rect
          x="3"
          y="3"
          width="7"
          height="7"
          rx="1"
          stroke="currentColor"
          strokeWidth="1.5"
          fill="none"
        />
        <rect
          x="14"
          y="3"
          width="7"
          height="7"
          rx="1"
          stroke="currentColor"
          strokeWidth="1.5"
          fill="none"
        />
        <rect
          x="3"
          y="14"
          width="7"
          height="7"
          rx="1"
          stroke="currentColor"
          strokeWidth="1.5"
          fill="none"
        />
        <rect
          x="14"
          y="14"
          width="7"
          height="7"
          rx="1"
          stroke="currentColor"
          strokeWidth="1.5"
          fill="none"
        />
      </svg>
    ),
  },
  {
    id: "create",
    label: "Create",
    description: "Generate video ads",
    icon: (
      <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
        <path
          d="M12 3.5l1.4 4.1 4.1 1.4-4.1 1.4L12 14.5l-1.4-4.1-4.1-1.4 4.1-1.4z"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.5"
          strokeLinejoin="round"
        />
        <circle
          cx="12"
          cy="12"
          r="2.2"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.5"
        />
      </svg>
    ),
  },
  {
    id: "videos",
    label: "Videos",
    description: "Browse library",
    icon: (
      <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
        <rect
          x="4"
          y="4"
          width="16"
          height="16"
          rx="2"
          stroke="currentColor"
          strokeWidth="1.5"
          fill="none"
        />
        <path d="M9 8l5 4-5 4V8z" fill="currentColor" />
      </svg>
    ),
  },
  {
    id: "settings",
    label: "Settings",
    description: "Configure options",
    icon: (
      <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
        <path
          d="M12 8.25A3.75 3.75 0 1112 15.75 3.75 3.75 0 0112 8.25z"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.5"
        />
        <path
          d="M4.5 12.75V11.25l2.1-.7a5.36 5.36 0 01.9-1.54l-.48-2.2 1.06-1.06 2.2.48a5.36 5.36 0 011.54-.9l.7-2.1h1.5l.7 2.1a5.36 5.36 0 011.54.9l2.2-.48 1.06 1.06-.48 2.2a5.36 5.36 0 01.9 1.54l2.1.7v1.5l-2.1.7a5.36 5.36 0 01-.9 1.54l.48 2.2-1.06 1.06-2.2-.48a5.36 5.36 0 01-1.54.9l-.7 2.1h-1.5l-.7-2.1a5.36 5.36 0 01-1.54-.9l-2.2.48-1.06-1.06.48-2.2a5.36 5.36 0 01-.9-1.54z"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.3"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    ),
  },
];

function IconMenu() {
  return (
    <svg
      viewBox="0 0 24 24"
      width="24"
      height="24"
      stroke="currentColor"
      strokeWidth="2"
      fill="none"
    >
      <line x1="3" y1="12" x2="21" y2="12" />
      <line x1="3" y1="6" x2="21" y2="6" />
      <line x1="3" y1="18" x2="21" y2="18" />
    </svg>
  );
}

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
              <IconMenu />
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
