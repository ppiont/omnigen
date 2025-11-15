import { useEffect, useState } from "react";
import Sidebar from "../components/Sidebar.jsx";
import VideoCard from "../components/VideoCard.jsx";
import ToggleSwitch from "../components/ToggleSwitch.jsx";
import "../styles/dashboard.css";

const sidebarTabs = [
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
        <path
          d="M9 8l5 4-5 4V8z"
          fill="currentColor"
        />
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

const recentVideos = [
  {
    id: "1",
    title: "Product Showcase - Tech Headphones",
    format: "16:9",
    duration: "30s",
    cost: "$1.20",
    timestamp: "2h ago",
    thumbnail: "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=600&h=400&fit=crop",
  },
  {
    id: "2",
    title: "Social Ad - Fashion Brand",
    format: "9:16",
    duration: "15s",
    cost: "$0.80",
    timestamp: "4h ago",
    thumbnail: "https://images.unsplash.com/photo-1441986300917-64674bd600d8?w=600&h=400&fit=crop",
  },
  {
    id: "3",
    title: "Product Demo - Smart Watch",
    format: "1:1",
    duration: "30s",
    cost: "$1.20",
    timestamp: "1d ago",
    thumbnail: "https://images.unsplash.com/photo-1523275335684-37898b6baf30?w=600&h=400&fit=crop",
  },
  {
    id: "4",
    title: "Brand Video - Minimal Aesthetic",
    format: "16:9",
    duration: "60s",
    cost: "$2.10",
    timestamp: "2d ago",
    thumbnail: "https://images.unsplash.com/photo-1561070791-2526d30994b5?w=600&h=400&fit=crop",
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

const IconChevronDown = () => (
  <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
    <path
      d="M6 9l6 6 6-6"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      fill="none"
    />
  </svg>
);

function Dashboard() {
  const [activeTab, setActiveTab] = useState("create");
  const [isDrawerOpen, setIsDrawerOpen] = useState(false);
  const [isCollapsed, setIsCollapsed] = useState(false);
  const [prompt, setPrompt] = useState("");
  const [isAdvancedOpen, setIsAdvancedOpen] = useState(false);
  const [selectedCategory, setSelectedCategory] = useState("Ad Creative");
  const [selectedStyle, setSelectedStyle] = useState("Cinematic");
  const [selectedDuration, setSelectedDuration] = useState("30s");
  const [selectedAspect, setSelectedAspect] = useState("16:9");
  const [autoEnhance, setAutoEnhance] = useState(true);
  const [loopVideo, setLoopVideo] = useState(false);
  const [isGenerating, setIsGenerating] = useState(false);
  const [progress, setProgress] = useState(0);
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [showCurrentPassword, setShowCurrentPassword] = useState(false);
  const [showNewPassword, setShowNewPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

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

  const characterLimit = 2000;
  const characterCount = prompt.length;

  const getEstimatedTime = () => {
    const durationNum = parseInt(selectedDuration);
    if (durationNum <= 15) return "~30s";
    if (durationNum <= 30) return "~45s";
    if (durationNum <= 60) return "1 min";
    return "1-2 min";
  };

  const getEstimatedCost = () => {
    const durationNum = parseInt(selectedDuration);
    const cost = (durationNum / 30) * 1.5;
    return `$${cost.toFixed(2)}`;
  };

  const handleGenerate = () => {
    if (!prompt.trim() || isGenerating) return;

    setIsGenerating(true);
    setProgress(0);

    const interval = setInterval(() => {
      setProgress((prev) => {
        if (prev >= 100) {
          clearInterval(interval);
          setIsGenerating(false);
          setProgress(0);
          return 0;
        }
        return prev + 2;
      });
    }, 50);
  };

  const toggleAdvanced = () => {
    setIsAdvancedOpen(!isAdvancedOpen);
  };

  const categories = ["Ad Creative", "Product Showcase", "Social Media"];
  const styles = ["Cinematic", "Anime", "Realistic", "Abstract"];
  const durations = ["15s", "30s", "60s", "90s"];
  const aspects = ["16:9", "9:16", "1:1"];

  return (
    <main className="dashboard-page">
      <div className="dashboard-shell">
        <Sidebar
          tabs={sidebarTabs}
          activeTab={activeTab}
          onSelect={setActiveTab}
          isDrawerOpen={isDrawerOpen}
          onClose={closeDrawer}
          isCollapsed={isCollapsed}
          onToggleCollapse={() => setIsCollapsed(!isCollapsed)}
        />

        <section className="dashboard-main">
          <div className="mobile-header">
            <button
              type="button"
              className="drawer-toggle"
              aria-label="Open sidebar"
              onClick={() => setIsDrawerOpen(true)}
            >
              <IconMenu />
            </button>
            <h1>Ad Creative Studio</h1>
          </div>

          <div className="dashboard-content">
            {activeTab === "create" && (
              <>
                <div className="generation-grid">
                  <section className="prompt-card">
                    <h2 className="card-title">Generate a video</h2>
                    <div className="prompt-section">
                      <label className="prompt-label">
                        Video Prompt
                        <span className="char-counter">
                          {characterCount} / {characterLimit}
                        </span>
                      </label>
                      <textarea
                        className="prompt-textarea"
                        placeholder="Describe your video ad... (e.g., 'Product showcase video for wireless headphones with modern aesthetic')"
                        value={prompt}
                        onChange={(e) => setPrompt(e.target.value)}
                        maxLength={characterLimit}
                        rows={6}
                      />
                    </div>
                  </section>

                  <section className="preview-card">
                    <h3 className="card-subtitle">Preview</h3>
                    <div className="preview-container">
                      {isGenerating ? (
                        <div className="preview-generating">
                          <div className="preview-progress-bar">
                            <div
                              className="preview-progress-fill"
                              style={{ width: `${progress}%` }}
                            />
                          </div>
                          <p className="preview-status">Generating video... {progress}%</p>
                        </div>
                      ) : (
                        <div className="preview-placeholder">
                          <div className="preview-aspect-ratio">
                            <span className="preview-aspect-text">{selectedAspect}</span>
                          </div>
                          <p className="preview-placeholder-text">Your video will appear here</p>
                        </div>
                      )}
                    </div>
                    <div className="estimation-grid">
                      <div className="estimation-item">
                        <span className="estimation-label">Estimated time</span>
                        <span className="estimation-value">{getEstimatedTime()}</span>
                      </div>
                      <div className="estimation-item">
                        <span className="estimation-label">Estimated cost</span>
                        <span className="estimation-value">{getEstimatedCost()}</span>
                      </div>
                    </div>
                  </section>
                </div>

                <section className="advanced-panel">
                  <button
                    type="button"
                    className="advanced-toggle"
                    onClick={toggleAdvanced}
                    aria-expanded={isAdvancedOpen}
                  >
                    <span>Advanced options</span>
                    <span className={`advanced-chevron ${isAdvancedOpen ? "is-open" : ""}`}>
                      <IconChevronDown />
                    </span>
                  </button>
                  {isAdvancedOpen && (
                    <div className="advanced-content">
                      <div className="options-grid">
                        <div className="option-group">
                          <label className="option-label">Category</label>
                          <select
                            className="dropdown-field"
                            value={selectedCategory}
                            onChange={(e) => setSelectedCategory(e.target.value)}
                          >
                            {categories.map((cat) => (
                              <option key={cat} value={cat}>
                                {cat}
                              </option>
                            ))}
                          </select>
                        </div>

                        <div className="option-group">
                          <label className="option-label">Style</label>
                          <div className="button-group">
                            {styles.map((style) => (
                              <button
                                key={style}
                                type="button"
                                className={`style-btn ${selectedStyle === style ? "is-active" : ""}`}
                                onClick={() => setSelectedStyle(style)}
                              >
                                {style}
                              </button>
                            ))}
                          </div>
                        </div>

                        <div className="option-group">
                          <label className="option-label">Duration</label>
                          <div className="button-group">
                            {durations.map((dur) => (
                              <button
                                key={dur}
                                type="button"
                                className={`duration-btn ${selectedDuration === dur ? "is-active" : ""}`}
                                onClick={() => setSelectedDuration(dur)}
                              >
                                {dur}
                              </button>
                            ))}
                          </div>
                        </div>

                        <div className="option-group">
                          <label className="option-label">Aspect Ratio</label>
                          <div className="button-group">
                            {aspects.map((aspect) => (
                              <button
                                key={aspect}
                                type="button"
                                className={`aspect-btn ${selectedAspect === aspect ? "is-active" : ""}`}
                                onClick={() => setSelectedAspect(aspect)}
                              >
                                {aspect}
                              </button>
                            ))}
                          </div>
                        </div>

                        <div className="option-group">
                          <label className="option-label">Options</label>
                          <div className="toggle-group">
                            <div className="toggle-item">
                              <ToggleSwitch
                                checked={autoEnhance}
                                onChange={() => setAutoEnhance(!autoEnhance)}
                                label="Auto-enhance"
                              />
                              <span className="toggle-label">Auto-enhance</span>
                            </div>
                            <div className="toggle-item">
                              <ToggleSwitch
                                checked={loopVideo}
                                onChange={() => setLoopVideo(!loopVideo)}
                                label="Loop video"
                              />
                              <span className="toggle-label">Loop video</span>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  )}
                </section>

                <button
                  type="button"
                  className="generate-button"
                  disabled={!prompt.trim() || isGenerating}
                  onClick={handleGenerate}
                >
                  {isGenerating ? "Generating..." : "Generate Video"}
                </button>
              </>
            )}

            {activeTab === "videos" && (
              <section className="videos-section">
                <h2 className="section-title">Your videos</h2>
                {recentVideos.length === 0 ? (
                  <div className="empty-state">
                    <p className="empty-state-text">Your generated videos will appear here</p>
                  </div>
                ) : (
                  <div className="videos-grid">
                    {recentVideos.map((video) => (
                      <VideoCard
                        key={video.id}
                        thumbnail={video.thumbnail}
                        title={video.title}
                        format={video.format}
                        duration={video.duration}
                        cost={video.cost}
                        timestamp={video.timestamp}
                      />
                    ))}
                  </div>
                )}
              </section>
            )}

            {activeTab === "settings" && (
              <section className="settings-section">
                <h2 className="section-title">Settings</h2>
                <div className="settings-card">
                  <h3 className="settings-subtitle">Change Password</h3>
                  <form
                    className="password-form"
                    onSubmit={(e) => {
                      e.preventDefault();
                      setIsSaving(true);
                      // Simulate API call
                      setTimeout(() => {
                        setIsSaving(false);
                        setCurrentPassword("");
                        setNewPassword("");
                        setConfirmPassword("");
                        alert("Password updated successfully!");
                      }, 1000);
                    }}
                  >
                    <div className="form-group">
                      <label className="form-label" htmlFor="current-password">
                        Current Password
                      </label>
                      <div className="password-input-wrapper">
                        <input
                          id="current-password"
                          type={showCurrentPassword ? "text" : "password"}
                          className="form-input"
                          value={currentPassword}
                          onChange={(e) => setCurrentPassword(e.target.value)}
                          placeholder="Enter your current password"
                          required
                        />
                        <button
                          type="button"
                          className="password-toggle"
                          onClick={() => setShowCurrentPassword(!showCurrentPassword)}
                          aria-label={showCurrentPassword ? "Hide password" : "Show password"}
                        >
                          {showCurrentPassword ? (
                            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2">
                              <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24" />
                              <line x1="1" y1="1" x2="23" y2="23" />
                            </svg>
                          ) : (
                            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2">
                              <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                              <circle cx="12" cy="12" r="3" />
                            </svg>
                          )}
                        </button>
                      </div>
                    </div>

                    <div className="form-group">
                      <label className="form-label" htmlFor="new-password">
                        New Password
                      </label>
                      <div className="password-input-wrapper">
                        <input
                          id="new-password"
                          type={showNewPassword ? "text" : "password"}
                          className="form-input"
                          value={newPassword}
                          onChange={(e) => setNewPassword(e.target.value)}
                          placeholder="Enter your new password"
                          required
                          minLength={8}
                        />
                        <button
                          type="button"
                          className="password-toggle"
                          onClick={() => setShowNewPassword(!showNewPassword)}
                          aria-label={showNewPassword ? "Hide password" : "Show password"}
                        >
                          {showNewPassword ? (
                            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2">
                              <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24" />
                              <line x1="1" y1="1" x2="23" y2="23" />
                            </svg>
                          ) : (
                            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2">
                              <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                              <circle cx="12" cy="12" r="3" />
                            </svg>
                          )}
                        </button>
                      </div>
                    </div>

                    <div className="form-group">
                      <label className="form-label" htmlFor="confirm-password">
                        Confirm New Password
                      </label>
                      <div className="password-input-wrapper">
                        <input
                          id="confirm-password"
                          type={showConfirmPassword ? "text" : "password"}
                          className="form-input"
                          value={confirmPassword}
                          onChange={(e) => setConfirmPassword(e.target.value)}
                          placeholder="Confirm your new password"
                          required
                        />
                        <button
                          type="button"
                          className="password-toggle"
                          onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                          aria-label={showConfirmPassword ? "Hide password" : "Show password"}
                        >
                          {showConfirmPassword ? (
                            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2">
                              <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24" />
                              <line x1="1" y1="1" x2="23" y2="23" />
                            </svg>
                          ) : (
                            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2">
                              <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                              <circle cx="12" cy="12" r="3" />
                            </svg>
                          )}
                        </button>
                      </div>
                      {confirmPassword && newPassword !== confirmPassword && (
                        <p className="form-error">Passwords do not match</p>
                      )}
                    </div>

                    <button
                      type="submit"
                      className="save-button"
                      disabled={isSaving || !currentPassword || !newPassword || newPassword !== confirmPassword}
                    >
                      {isSaving ? "Saving..." : "Save Changes"}
                    </button>
                  </form>
                </div>
              </section>
            )}
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
