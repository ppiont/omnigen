import Logo from "./Logo.jsx";
import { useTheme } from "../context/ThemeContext.jsx";

const IconChevronLeft = () => (
  <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
    <path
      d="M15 18l-6-6 6-6"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      fill="none"
    />
  </svg>
);

const IconChevronRight = () => (
  <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
    <path
      d="M9 18l6-6-6-6"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      fill="none"
    />
  </svg>
);

const IconSun = () => (
  <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false" width="18" height="18">
    <circle
      cx="12"
      cy="12"
      r="4"
      stroke="currentColor"
      strokeWidth="2"
      fill="none"
    />
    <path
      d="M12 2v2M12 20v2M2 12h2M20 12h2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
    />
  </svg>
);

const IconMoon = () => (
  <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false" width="18" height="18">
    <path
      d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"
      stroke="currentColor"
      strokeWidth="2"
      fill="none"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
);

const IconLogOut = () => (
  <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false" width="18" height="18">
    <path
      d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4M16 17l5-5-5-5M21 12H9"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      fill="none"
    />
  </svg>
);

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
            <IconChevronRight />
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
              <IconChevronLeft />
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
                {theme === "light" ? <IconMoon /> : <IconSun />}
              </button>
              <button
                type="button"
                className="sign-out-btn"
                onClick={handleSignOut}
                aria-label="Sign out"
                title="Sign out"
              >
                <IconLogOut />
              </button>
            </div>
          </>
        )}
      </div>
    </nav>
  );
}

export default Sidebar;
