import { ChevronLeft, ChevronRight, Sun, Moon, LogOut } from "lucide-react";
import Logo from "./Logo.jsx";
import { useTheme } from "../context/ThemeContext.jsx";

function Sidebar({
  tabs = [],
  activeTab,
  onSelect,
  isDrawerOpen = false,
  onClose,
  isCollapsed = false,
  onToggleCollapse,
}) {
  const { theme, toggleTheme } = useTheme();

  const handleSelect = (tabId) => {
    if (onSelect) {
      onSelect(tabId);
    }
    if (onClose) {
      onClose();
    }
  };

  const handleSignOut = () => {
    // TODO: Implement sign out logic
    console.log("Sign out clicked");
  };

  const sidebarClasses = [
    "dashboard-sidebar",
    isDrawerOpen ? "is-open" : "",
    isCollapsed ? "is-collapsed" : "",
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <nav
      className={sidebarClasses}
      aria-label="Primary navigation"
    >
      <div className="sidebar-brand">
        {isCollapsed ? (
          <button
            type="button"
            onClick={onToggleCollapse}
            className="collapse-toggle"
            aria-label="Expand sidebar"
          >
            <ChevronRight size={20} />
          </button>
        ) : (
          <>
            <div className="brand-content">
              <Logo size="small" />
              <span className="brand-mark">OMNIGEN</span>
            </div>
            <button
              type="button"
              onClick={onToggleCollapse}
              className="collapse-toggle"
              aria-label="Collapse sidebar"
            >
              <ChevronLeft size={20} />
            </button>
          </>
        )}
      </div>

      <ul className="sidebar-nav">
        {tabs.map((tab) => {
          const isActive = tab.id === activeTab;
          const classes = [
            "sidebar-tab",
            tab.id === "create" ? "sidebar-tab-create" : "",
            isActive ? "is-active" : "",
            tab.disabled ? "is-disabled" : "",
          ]
            .filter(Boolean)
            .join(" ");

          return (
            <li key={tab.id}>
              <button
                type="button"
                className={classes}
                onClick={() => !tab.disabled && handleSelect(tab.id)}
                aria-current={isActive ? "page" : undefined}
                disabled={tab.disabled}
                aria-label={isCollapsed ? tab.label : undefined}
                title={isCollapsed ? `${tab.label}${tab.description ? ` - ${tab.description}` : ""}` : undefined}
              >
                <span className="tab-icon" aria-hidden="true">
                  {tab.icon}
                </span>
                {!isCollapsed && (
                  <>
                    <span className="tab-content">
                      <span className="tab-label">{tab.label}</span>
                      {tab.description && (
                        <span className="tab-description">{tab.description}</span>
                      )}
                    </span>
                    {tab.badge && (
                      <span className="tab-badge">{tab.badge}</span>
                    )}
                  </>
                )}
                {isActive && !isCollapsed && (
                  <span className="tab-indicator" aria-hidden="true" />
                )}
              </button>
            </li>
          );
        })}
      </ul>

      <div className="sidebar-profile">
        {isCollapsed ? (
          <div className="profile-avatar">AP</div>
        ) : (
          <>
            <div className="profile-card">
              <div className="profile-avatar">AP</div>
              <div className="profile-info">
                <p className="profile-name">Akhil Patel</p>
                <span className="profile-plan">Max Plan</span>
              </div>
            </div>
            <div className="profile-actions">
              <button
                type="button"
                className="theme-toggle"
                onClick={toggleTheme}
                aria-label={`Switch to ${theme === "light" ? "dark" : "light"} theme`}
                title={`Switch to ${theme === "light" ? "dark" : "light"} theme`}
              >
                {theme === "light" ? <Moon size={18} /> : <Sun size={18} />}
              </button>
              <button
                type="button"
                className="sign-out-btn"
                onClick={handleSignOut}
                aria-label="Sign out"
                title="Sign out"
              >
                <LogOut size={18} />
              </button>
            </div>
          </>
        )}
      </div>
    </nav>
  );
}

export default Sidebar;
