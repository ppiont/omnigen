import { ChevronLeft, ChevronRight, Sun, Moon, LogOut } from "lucide-react";
import Logo from "./Logo.jsx";
import { useTheme } from "../context/useTheme.js";
import { useAuth } from "../contexts/useAuth.js";

/**
 * Format user name to "First L" (first name + last initial)
 */
function formatUserName(fullName) {
  if (!fullName) return "";
  const parts = fullName.trim().split(/\s+/);
  if (parts.length === 1) return parts[0];
  const firstName = parts[0];
  const lastInitial = parts[parts.length - 1][0];
  return `${firstName} ${lastInitial}`;
}

/**
 * Get user initials for avatar (first letter of first and last name)
 */
function getUserInitials(fullName) {
  if (!fullName) return "U";
  const parts = fullName.trim().split(/\s+/);
  if (parts.length === 1) return parts[0][0].toUpperCase();
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

/**
 * Format subscription tier for display
 */
function formatPlanName(tier) {
  if (!tier || tier === "free") return "Free Plan";
  if (tier === "max") return "Max Plan";
  if (tier === "pro") return "Pro Plan";
  // Capitalize first letter
  return tier.charAt(0).toUpperCase() + tier.slice(1) + " Plan";
}

function Sidebar({
  tabs = [],
  activeTab,
  onSelect,
  isDrawerOpen = false,
  onClose,
  isCollapsed = false,
  onToggleCollapse,
}) {
  const { user, logout } = useAuth();
  const { theme, toggleTheme } = useTheme();

  const displayName = formatUserName(user?.name || "User");
  const initials = getUserInitials(user?.name || "User");
  const planName = formatPlanName(user?.subscription_tier);

  const handleSignOut = async () => {
    if (logout) {
      await logout();
    }
  };

  const handleSelect = (tabId) => {
    if (onSelect) {
      onSelect(tabId);
    }
    if (onClose) {
      onClose();
    }
  };

  const sidebarClasses = [
    "dashboard-sidebar",
    isDrawerOpen ? "is-open" : "",
    isCollapsed ? "is-collapsed" : "",
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <nav className={sidebarClasses} aria-label="Primary navigation">
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
<<<<<<< Updated upstream
                title={
                  isCollapsed
                    ? `${tab.label}${
                        tab.description ? ` - ${tab.description}` : ""
                      }`
                    : undefined
                }
=======
                title={isCollapsed ? tab.label : undefined}
>>>>>>> Stashed changes
              >
                <span className="tab-icon" aria-hidden="true">
                  {tab.icon}
                </span>
                {!isCollapsed && (
                  <>
<<<<<<< Updated upstream
                    <span className="tab-content">
                      <span className="tab-label">{tab.label}</span>
                      {tab.description && (
                        <span className="tab-description">
                          {tab.description}
                        </span>
                      )}
                    </span>
=======
                    <span className="tab-label">{tab.label}</span>
>>>>>>> Stashed changes
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
          <div className="profile-avatar">{initials}</div>
        ) : (
          <>
            <div className="profile-card">
              <div className="profile-avatar">{initials}</div>
              <div className="profile-info">
                <p className="profile-name">{displayName}</p>
                <span className="profile-plan">{planName}</span>
              </div>
            </div>
            <div className="profile-actions">
              <button
                type="button"
                className="profile-action-btn"
                onClick={toggleTheme}
                aria-label={`Switch to ${
                  theme === "light" ? "dark" : "light"
                } theme`}
                title={`Switch to ${
                  theme === "light" ? "dark" : "light"
                } theme`}
              >
                {theme === "light" ? <Moon size={18} /> : <Sun size={18} />}
              </button>
              <button
                type="button"
                className="profile-action-btn"
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
