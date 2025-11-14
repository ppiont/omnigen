import { useEffect, useState } from "react";
import Sidebar from "../components/Sidebar.jsx";
import StatCard from "../components/StatCard.jsx";
import "../styles/dashboard.css";

const sidebarTabs = [
  {
    id: "create",
    label: "Create",
    description: "Launch new flows",
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
    id: "dashboard",
    label: "Dashboard",
    description: "Monitor signals",
    icon: (
      <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
        <rect
          x="4"
          y="4"
          width="7"
          height="7"
          rx="1.5"
          stroke="currentColor"
          strokeWidth="1.5"
          fill="none"
        />
        <rect
          x="13"
          y="4"
          width="7"
          height="5"
          rx="1.5"
          stroke="currentColor"
          strokeWidth="1.5"
          fill="none"
        />
        <rect
          x="4"
          y="13"
          width="7"
          height="7"
          rx="1.5"
          stroke="currentColor"
          strokeWidth="1.5"
          fill="none"
        />
        <rect
          x="13"
          y="11"
          width="7"
          height="9"
          rx="1.5"
          stroke="currentColor"
          strokeWidth="1.5"
          fill="none"
        />
      </svg>
    ),
  },
  {
    id: "projects",
    label: "Projects",
    description: "Organize assets",
    icon: (
      <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
        <path
          d="M4 8.5h16v8.25A2.25 2.25 0 0117.75 19H6.25A2.25 2.25 0 014 16.75z"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.5"
          strokeLinejoin="round"
        />
        <path
          d="M4 8.5V6.75A2.25 2.25 0 016.25 4.5H10l2 2.5h5.75A2.25 2.25 0 0120 9.25V8.5"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.5"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    ),
  },
  {
    id: "settings",
    label: "Settings",
    description: "Fine-tune system",
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

const stats = [
  {
    label: "Total Projects",
    value: "24",
    helper: "Updated 2 hours ago",
    trend: { direction: "up", value: "+12%", caption: "this week" },
  },
  {
    label: "Active Generations",
    value: "8",
    helper: "Across 3 agents",
    trend: { direction: "up", value: "+3", caption: "vs yesterday" },
  },
  {
    label: "API Calls (30d)",
    value: "1,247",
    helper: "Through orchestrator",
    trend: { direction: "up", value: "+18%", caption: "month over month" },
  },
  {
    label: "Success Rate",
    value: "98.4%",
    helper: "Autonomous runs",
    trend: { direction: "down", value: "-0.6%", caption: "variance" },
  },
];

const activityFeed = [
  {
    id: "a1",
    title: "Vision synthesis pipeline deployed",
    description: "Create tab • Multimodal agents linked",
    time: "2m ago",
  },
  {
    id: "a2",
    title: "Workspace “Neon Pulse” archived",
    description: "Projects • Auto clean executed",
    time: "18m ago",
  },
  {
    id: "a3",
    title: "Policy tuning completed",
    description: "Settings • Risk threshold tightened",
    time: "42m ago",
  },
  {
    id: "a4",
    title: "Prompt pack shared with Delta team",
    description: "Dashboard • Access synced",
    time: "1h ago",
  },
];

const tips = [
  {
    id: "t1",
    title: "Highlight Create tab",
    body: "Pin high-priority flows so invites drop users directly into action.",
  },
  {
    id: "t2",
    title: "Balance agent load",
    body: "Use the Projects tab to cap concurrent generations per workspace.",
  },
  {
    id: "t3",
    title: "Watch drift windows",
    body: "Settings → Safeguards lets you automate aurora alerts for anomalies.",
  },
];

const IconSearch = () => (
  <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
    <circle
      cx="11"
      cy="11"
      r="6.5"
      stroke="currentColor"
      strokeWidth="1.5"
      fill="none"
    />
    <path
      d="M16.5 16.5L21 21"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
    />
  </svg>
);

const IconMenu = () => (
  <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
    <path
      d="M4 7h16M4 12h16M4 17h10"
      stroke="currentColor"
      strokeWidth="1.7"
      strokeLinecap="round"
    />
  </svg>
);

function Dashboard() {
  const [activeTab, setActiveTab] = useState("create");
  const [isDrawerOpen, setIsDrawerOpen] = useState(false);

  const closeDrawer = () => setIsDrawerOpen(false);

  useEffect(() => {
    if (!isDrawerOpen) {
      return undefined;
    }

    const handleKeyDown = (event) => {
      if (event.key === "Escape") {
        closeDrawer();
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isDrawerOpen]);

  return (
    <main className="dashboard-page">
      <div className="dashboard-shell">
        <Sidebar
          tabs={sidebarTabs}
          activeTab={activeTab}
          onSelect={setActiveTab}
          isDrawerOpen={isDrawerOpen}
          onClose={closeDrawer}
        />

        <section className="dashboard-main">
          <header className="dashboard-topbar">
            <div className="topbar-left">
              <button
                type="button"
                className="drawer-toggle"
                aria-label="Open sidebar"
                onClick={() => setIsDrawerOpen(true)}
              >
                <IconMenu />
              </button>
              <div className="topbar-text">
                <p className="page-kicker">Workspace</p>
                <h1>Create Command Center</h1>
              </div>
            </div>

            <div className="topbar-right">
              <label className="search-field">
                <span className="search-icon">
                  <IconSearch />
                </span>
                <input
                  type="search"
                  placeholder="Search agents, projects, prompts..."
                  aria-label="Search workspace"
                />
              </label>
              <button
                type="button"
                className="avatar-pill"
                aria-label="Open profile menu"
              >
                <span>YC</span>
              </button>
            </div>
          </header>

          <section className="stats-grid" aria-label="Workspace stats">
            {stats.map((stat, index) => (
              <StatCard
                key={stat.label}
                {...stat}
                motionDelay={`${index * 0.08}s`}
              />
            ))}
          </section>

          <div className="panels-grid">
            <section
              className="dashboard-panel primary"
              aria-label="Recent activity"
            >
              <header className="panel-header">
                <div>
                  <p className="panel-kicker">Live feed</p>
                  <h2>Recent Activity</h2>
                </div>
                <button type="button" className="text-link">
                  View all
                </button>
              </header>

              <ul className="activity-list">
                {activityFeed.map((item) => (
                  <li key={item.id} className="activity-item">
                    <div className="activity-copy">
                      <p className="activity-title">{item.title}</p>
                      <p className="activity-meta">{item.description}</p>
                    </div>
                    <span className="activity-time">{item.time}</span>
                  </li>
                ))}
              </ul>
            </section>

            <section
              className="dashboard-panel secondary"
              aria-label="Quick tips"
            >
              <header className="panel-header">
                <div>
                  <p className="panel-kicker">Guidance</p>
                  <h2>Quick Tips</h2>
                </div>
              </header>

              <ul className="tips-list">
                {tips.map((tip, index) => (
                  <li
                    key={tip.id}
                    className="tip-card"
                    data-motion="rise"
                    style={{ "--motion-delay": `${0.15 + index * 0.05}s` }}
                  >
                    <p className="tip-title">{tip.title}</p>
                    <p className="tip-body">{tip.body}</p>
                  </li>
                ))}
              </ul>
            </section>
          </div>
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

export default Dashboard;
