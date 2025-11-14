function Sidebar({
  tabs = [],
  activeTab,
  onSelect,
  isDrawerOpen = false,
  onClose,
}) {
  const handleSelect = (tabId) => {
    if (onSelect) {
      onSelect(tabId);
    }
    if (onClose) {
      onClose();
    }
  };

  return (
    <nav
      className={`dashboard-sidebar ${isDrawerOpen ? "is-open" : ""}`}
      aria-label="Primary navigation"
    >
      <div className="sidebar-brand">
        <span className="brand-mark">Omnigen</span>
        <p className="brand-subtitle">Neon workspace</p>
      </div>

      <ul className="sidebar-nav">
        {tabs.map((tab) => {
          const isActive = tab.id === activeTab;
          const classes = [
            "sidebar-tab",
            tab.id === "create" ? "sidebar-tab-create" : "",
            isActive ? "is-active" : "",
          ]
            .filter(Boolean)
            .join(" ");

          return (
            <li key={tab.id}>
              <button
                type="button"
                className={classes}
                onClick={() => handleSelect(tab.id)}
                aria-current={isActive ? "page" : undefined}
              >
                <span className="tab-icon" aria-hidden="true">
                  {tab.icon}
                </span>
                <span className="tab-content">
                  <span className="tab-label">{tab.label}</span>
                  {tab.description && (
                    <span className="tab-description">{tab.description}</span>
                  )}
                </span>
                {isActive && (
                  <span className="tab-indicator" aria-hidden="true" />
                )}
              </button>
            </li>
          );
        })}
      </ul>
    </nav>
  );
}

export default Sidebar;
