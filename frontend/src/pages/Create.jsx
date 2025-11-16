import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { v4 as uuidv4 } from "uuid";
import ToggleSwitch from "../components/ToggleSwitch.jsx";
import BrandPresetSelector from "../components/create/BrandPresetSelector.jsx";
import MediaUploadBar from "../components/create/MediaUploadBar.jsx";
import ValidationMessage from "../components/create/ValidationMessage.jsx";
import BatchGenerationToggle from "../components/create/BatchGenerationToggle.jsx";
import GenerationState from "../components/create/GenerationState.jsx";
import ScenePreviewGrid from "../components/create/ScenePreviewGrid.jsx";
import { simulateGeneration, resetGeneration } from "../utils/mockGenerationData.js";
import "../styles/dashboard.css";
import "../styles/create.css";

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

  // Phase 1 additions
  const [selectedBrandPreset, setSelectedBrandPreset] = useState("default");
  const [referenceImage, setReferenceImage] = useState(null);
  const [templateImage, setTemplateImage] = useState(null);
  const [validationError, setValidationError] = useState("");

  // Phase 2 additions - State Machine
  const [generationState, setGenerationState] = useState("idle");
  const [generationProgress, setGenerationProgress] = useState(0);
  const [scenes, setScenes] = useState([]);
  const [sceneCount, setSceneCount] = useState(0);
  const [currentScene, setCurrentScene] = useState(0);
  const [videoPreview, setVideoPreview] = useState(null);
  const [generationError, setGenerationError] = useState(null);
  const [generatedJobId, setGeneratedJobId] = useState(null);

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

  const handleGenerate = async () => {
    // Validation
    if (!prompt.trim()) {
      setValidationError("Please describe your video to get started");
      return;
    }

    if (isGenerating) return;

    // Clear validation error and reset state
    setValidationError("");
    setIsGenerating(true);
    setGenerationError(null);

    // Generate job ID
    const jobId = uuidv4();
    setGeneratedJobId(jobId);

    // Build config
    const config = {
      category: selectedCategory,
      style: selectedStyle,
      duration: selectedDuration,
      aspectRatio: selectedAspect,
      brandPreset: selectedBrandPreset,
      autoEnhance,
      loopVideo,
      referenceImage: referenceImage?.name || null,
      templateImage: templateImage?.name || null,
    };

    // State change handler
    const handleStateChange = (stateUpdate) => {
      setGenerationState(stateUpdate.state);
      if (stateUpdate.progress !== undefined)
        setGenerationProgress(stateUpdate.progress);
      if (stateUpdate.sceneCount !== undefined)
        setSceneCount(stateUpdate.sceneCount);
      if (stateUpdate.currentScene !== undefined)
        setCurrentScene(stateUpdate.currentScene);
      if (stateUpdate.videoPreview !== undefined)
        setVideoPreview(stateUpdate.videoPreview);
      if (stateUpdate.error !== undefined)
        setGenerationError(stateUpdate.error);
    };

    // Scene update handler
    const handleSceneUpdate = (updatedScenes) => {
      setScenes(updatedScenes);
    };

    // Run simulation
    const result = await simulateGeneration(
      prompt,
      config,
      handleStateChange,
      handleSceneUpdate
    );

    setIsGenerating(false);

    if (!result.success) {
      console.error("Generation failed:", result.error);
    }
  };

  // Get character counter class based on count
  const getCharCounterClass = () => {
    if (characterCount >= characterLimit) return "char-counter danger";
    if (characterCount >= characterLimit * 0.9) return "char-counter warning";
    return "char-counter normal";
  };

  const toggleAdvanced = () => setIsAdvancedOpen(!isAdvancedOpen);

  // Handle viewing in workspace
  const handleViewWorkspace = () => {
    if (!generatedJobId) return;

    navigate(`/workspace/${generatedJobId}`, {
      state: {
        prompt,
        scenes,
        videoPreview,
        config: {
          category: selectedCategory,
          style: selectedStyle,
          duration: selectedDuration,
          aspectRatio: selectedAspect,
          brandPreset: selectedBrandPreset,
          autoEnhance,
          loopVideo,
          referenceImage: referenceImage?.name || null,
          templateImage: templateImage?.name || null,
        },
        generatedAt: Date.now(),
      },
    });
  };

  // Handle retry/generate another
  const handleRetry = () => {
    const reset = resetGeneration();
    setGenerationState(reset.state);
    setGenerationProgress(reset.progress);
    setScenes(reset.scenes);
    setSceneCount(reset.sceneCount);
    setCurrentScene(reset.currentScene);
    setGenerationError(reset.error);
    setVideoPreview(reset.videoPreview);
    setGeneratedJobId(null);
    setIsGenerating(false);
  };

  return (
    <>
      <div className="generation-grid">
        <section className="prompt-card">
          <h2 className="card-title">Generate a video</h2>
          <div className="prompt-section">
            <label className="prompt-label">
              Video Prompt
              <span className={getCharCounterClass()}>
                {characterCount} / {characterLimit}
              </span>
            </label>
            <textarea
              className="prompt-textarea"
              placeholder="Describe your video ad... (e.g., 'Product showcase video for wireless headphones with modern aesthetic')"
              value={prompt}
              onChange={(e) => {
                setPrompt(e.target.value);
                if (validationError) setValidationError("");
              }}
              maxLength={characterLimit}
              rows={6}
            />

            {/* Media Upload Bar - Below prompt */}
            <MediaUploadBar
              referenceImage={referenceImage}
              onReferenceImageSelect={setReferenceImage}
              templateImage={templateImage}
              onTemplateImageSelect={setTemplateImage}
              durations={durations}
              selectedDuration={selectedDuration}
              onDurationChange={setSelectedDuration}
            />

            {validationError && (
              <ValidationMessage message={validationError} type="error" />
            )}
          </div>
        </section>

        <section className="preview-card">
          <h3 className="card-subtitle">Preview</h3>
          <div className="preview-container">
            <GenerationState
              state={generationState}
              progress={generationProgress}
              error={generationError}
              sceneCount={sceneCount}
              currentScene={currentScene}
              videoPreview={videoPreview}
              onRetry={handleRetry}
              onViewWorkspace={handleViewWorkspace}
              aspectRatio={selectedAspect}
            />

            {/* Scene Preview Grid - show during PLANNING, RENDERING, STITCHING, READY */}
            <ScenePreviewGrid
              scenes={scenes}
              isVisible={["planning", "rendering", "stitching", "ready"].includes(
                generationState
              )}
            />
          </div>

          {/* Estimation - only show when idle */}
          {generationState === "idle" && (
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
          )}
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
              {/* Brand Preset Selector */}
              <BrandPresetSelector
                selectedPreset={selectedBrandPreset}
                onChange={setSelectedBrandPreset}
              />

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
                <p className="option-helper">
                  Choose the type of content you're creating
                </p>
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
                <p className="option-helper">
                  Choose the visual aesthetic for your video
                </p>
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
                <p className="option-helper">
                  16:9 for YouTube, 9:16 for TikTok/Stories, 1:1 for Instagram
                  feed
                </p>
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
                <p className="option-helper">
                  Additional video enhancements and playback options
                </p>
              </div>

              {/* Batch Generation Toggle */}
              <BatchGenerationToggle />
            </div>
          </div>
        )}
      </section>

      <button
        type="button"
        className="generate-button"
        disabled={!prompt.trim() || isGenerating || generationState !== "idle"}
        onClick={handleGenerate}
      >
        {isGenerating ? "Generating..." : "Generate Video"}
      </button>
    </>
  );
}

export default Create;
