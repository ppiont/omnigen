import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { v4 as uuidv4 } from "uuid";
import ToggleSwitch from "../components/ToggleSwitch.jsx";
import "../styles/dashboard.css";

const categories = [
  "Ad Creative",
  "Product Demo",
  "Explainer",
  "Social Media",
  "Tutorial",
];
const styles = ["Cinematic", "Modern", "Minimalist", "Bold", "Playful"];
const durations = ["15s", "30s", "60s", "90s"];
const aspects = ["16:9", "9:16", "1:1", "4:5"];

function IconChevronDown() {
  return (
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
}

function Create() {
  const navigate = useNavigate();
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

          // Generate UUID for new video and redirect to workspace
          const videoId = uuidv4();
          navigate(`/workspace/${videoId}`);

          return 100;
        }
        return prev + 10;
      });
    }, 500);
  };

  const toggleAdvanced = () => setIsAdvancedOpen(!isAdvancedOpen);

  return (
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
                <p className="preview-status">
                  Generating video... {progress}%
                </p>
              </div>
            ) : (
              <div className="preview-placeholder">
                <div className="preview-aspect-ratio">
                  <span className="preview-aspect-text">{selectedAspect}</span>
                </div>
                <p className="preview-placeholder-text">
                  Your video will appear here
                </p>
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
          <span
            className={`advanced-chevron ${isAdvancedOpen ? "is-open" : ""}`}
          >
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
                      className={`style-btn ${
                        selectedStyle === style ? "is-active" : ""
                      }`}
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
                      className={`duration-btn ${
                        selectedDuration === dur ? "is-active" : ""
                      }`}
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
                      className={`aspect-btn ${
                        selectedAspect === aspect ? "is-active" : ""
                      }`}
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
  );
}

export default Create;
