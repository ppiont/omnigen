import { useState } from "react";
import { useNavigate } from "react-router-dom";
import ToggleSwitch from "../components/ToggleSwitch.jsx";
import BrandPresetSelector from "../components/create/BrandPresetSelector.jsx";
import MediaUploadBar from "../components/create/MediaUploadBar.jsx";
import ValidationMessage from "../components/create/ValidationMessage.jsx";
import BatchGenerationToggle from "../components/create/BatchGenerationToggle.jsx";
import GenerationState from "../components/create/GenerationState.jsx";
import ScenePreviewGrid from "../components/create/ScenePreviewGrid.jsx";
import ProgressModal from "../components/create/ProgressModal.jsx";
import { generate, jobs } from "../utils/api.js";
import "../styles/dashboard.css";
import "../styles/create.css";

const categories = [
  "Ad Creative",
  "Product Demo",
  "Explainer",
  "Social Media",
  "Tutorial",
];
const styles = ["Clinical", "Professional", "Documentary", "Informative", "Trustworthy"];
const durations = ["10s", "20s", "30s", "40s", "50s", "60s"]; // Must be multiple of 10
const aspects = ["16:9", "9:16", "1:1"]; // Backend only supports these

// Removed unused options: visualStyles, tones, tempos, platforms, goals

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
  const [selectedStyle, setSelectedStyle] = useState("Clinical");
  const [selectedDuration, setSelectedDuration] = useState("30s");
  const [selectedAspect, setSelectedAspect] = useState("16:9");
  const [autoEnhance, setAutoEnhance] = useState(true);
  const [loopVideo, setLoopVideo] = useState(false);
  const [isGenerating, setIsGenerating] = useState(false);
  const [progress, setProgress] = useState(0);

  // Phase 1 additions
  const [selectedBrandPreset, setSelectedBrandPreset] = useState("default");
  const [productImage, setProductImage] = useState(null);
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

  // Progress modal state
  const [generatingJobId, setGeneratingJobId] = useState(null);

  // Voice selection
  const [voice, setVoice] = useState("Ash"); // Default to Ash (male)
  // Side Effects - Required field for pharmaceutical ads
  const [sideEffects, setSideEffects] = useState("");

  // Map pharmaceutical styles to backend-compatible styles
  const styleMap = {
    'Clinical': 'documentary',      // Clinical → documentary (medical/realistic)
    'Professional': 'cinematic',    // Professional → cinematic (polished)
    'Documentary': 'documentary',   // Documentary → documentary (matches)
    'Informative': 'documentary',   // Informative → documentary (educational)
    'Trustworthy': 'cinematic',     // Trustworthy → cinematic (polished, credible)
  };

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
    console.log("=".repeat(80));
    console.log("🎬 [CREATE] VIDEO GENERATION PIPELINE STARTED");
    console.log("=".repeat(80));
    console.log("[CREATE] 📝 User Input:", {
      prompt: prompt.trim(),
      category: selectedCategory,
      style: selectedStyle,
      duration: selectedDuration,
      aspectRatio: selectedAspect,
      brandPreset: selectedBrandPreset,
    });

    // Validation
    if (!prompt.trim()) {
      console.warn("[CREATE] ⚠️ Validation failed: Empty prompt");
      setValidationError("Please describe your video to get started");
      return;
    }

    // Validate side effects (required for pharmaceutical ads)
    if (!sideEffects.trim()) {
      console.warn("[CREATE] ⚠️ Validation failed: Empty side effects");
      setValidationError("Side Effects is required. Please enter the side effects information.");
      return;
    }

    if (isGenerating) {
      console.warn("[CREATE] ⚠️ Already generating, ignoring request");
      return;
    }

    // Validate duration is multiple of 10
    const durationNum = parseInt(selectedDuration);
    if (durationNum % 10 !== 0) {
      console.warn(
        "[CREATE] ⚠️ Validation failed: Duration must be multiple of 10"
      );
      setValidationError("Duration must be 10, 20, 30, 40, 50, or 60 seconds");
      return;
    }

    // Clear validation error and reset state
    setValidationError("");
    setIsGenerating(true);
    setGenerationError(null);

    try {
      console.log("\n" + "=".repeat(80));
      console.log("🎥 [GENERATE] Starting video generation");
      console.log("=".repeat(80));

      setGenerationState("rendering");
      setGenerationProgress(0);

      // Prepare generate request with required fields
      const generateParams = {
        prompt: prompt.trim(),
        duration: durationNum,
        aspect_ratio: selectedAspect,
      };
      

      // Add style (pharmaceutical ad styles) - MAP to backend-compatible style
      if (selectedStyle) {
        const backendStyle = styleMap[selectedStyle] || 'documentary'; // Default to documentary
        generateParams.style = backendStyle;
        console.log(`[CREATE] Mapped style "${selectedStyle}" to backend style "${backendStyle}"`);
      }

      // Voice and side_effects are kept in UI for future use but not sent to API yet
      // TODO: Add voice and side_effects support when backend is ready
      // if (voice) {
      //   generateParams.voice = voice;
      // }
      // if (sideEffects.trim()) {
      //   generateParams.side_effects = sideEffects.trim();
      // }

      // Add start_image (Product Image - used ONLY for first scene)
      if (productImage) {
        // Use preview (data URI) if available, otherwise skip
        if (
          productImage.preview &&
          productImage.preview.startsWith("data:image/")
        ) {
          generateParams.start_image = productImage.preview;
          console.log("[CREATE] 📸 Using product image for first scene (data URI)");
        } else if (
          productImage.url &&
          (productImage.url.startsWith("http://") ||
            productImage.url.startsWith("https://"))
        ) {
          generateParams.start_image = productImage.url;
          console.log("[CREATE] 📸 Using product image for first scene (URL)");
        } else {
          console.log(
            "[CREATE] ⚠️ Product image provided but not a valid URL, skipping"
          );
        }
      }

      console.log("[CREATE] 📡 API Call: POST /api/v1/generate");
      console.log("[CREATE] 📦 Request payload:", generateParams);

      const generateResponse = await generate.create(generateParams);

      const jobId = generateResponse.job_id;
      console.log("[CREATE] ✅ Video generation job created");
      console.log("[CREATE] 🆔 Job ID:", jobId);
      console.log("[CREATE] 📊 Status:", generateResponse.status);
      console.log(
        "[CREATE] ⏱️ Estimated completion:",
        generateResponse.estimated_completion_seconds || "N/A",
        "seconds"
      );

      // Store job ID for legacy compatibility
      setGeneratedJobId(jobId);

      // Open progress modal with SSE tracking
      setGeneratingJobId(jobId);
    } catch (error) {
      console.error("\n" + "=".repeat(80));
      console.error("❌ [ERROR] VIDEO GENERATION PIPELINE ERROR");
      console.error("=".repeat(80));
      console.error("[ERROR] Generation failed:", error);
      setGenerationState("error");
      setGenerationError(error.message || "Generation failed");
      setIsGenerating(false);
    }

    console.log("=".repeat(80));
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
    if (!generatedJobId) {
      console.warn("[CREATE] ⚠️ Cannot navigate to workspace: No job ID");
      return;
    }

    console.log("[CREATE] 🚀 Navigating to workspace:", generatedJobId);
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
          productImage: productImage?.name || null,
        },
        generatedAt: Date.now(),
      },
    });
  };

  // Handle progress modal completion
  const handleProgressComplete = async (jobData) => {
    console.log("[CREATE] ✅ Progress modal completion:", jobData);

    try {
      // Fetch final job details to get video URL
      const finalJob = await jobs.get(jobData.job_id);
      console.log("[CREATE] 📊 Final job data:", finalJob);

      // Update state for completion
      setGenerationState("ready");
      setGenerationProgress(100);
      setIsGenerating(false);
      setGeneratingJobId(null);

      // Navigate to workspace with complete data
      navigate(`/workspace/${jobData.job_id}`, {
        state: {
          prompt,
          scenes,
          videoPreview: finalJob.video_url,
          config: {
            category: selectedCategory,
            style: selectedStyle,
            duration: selectedDuration,
            aspectRatio: selectedAspect,
            brandPreset: selectedBrandPreset,
            autoEnhance,
            loopVideo,
            productImage: productImage?.name || null,
          },
          generatedAt: Date.now(),
        },
      });
    } catch (error) {
      console.error("[CREATE] ❌ Error fetching final job data:", error);
      // Fallback: navigate anyway with basic info
      setGenerationState("ready");
      setGenerationProgress(100);
      setIsGenerating(false);
      setGeneratingJobId(null);

      navigate(`/workspace/${jobData.job_id}`, {
        state: {
          prompt,
          config: {
            category: selectedCategory,
            style: selectedStyle,
            duration: selectedDuration,
            aspectRatio: selectedAspect,
            brandPreset: selectedBrandPreset,
          },
          generatedAt: Date.now(),
        },
      });
    }
  };

  // Handle progress modal cancellation
  const handleProgressCancel = () => {
    console.log("[CREATE] ❌ Progress modal cancelled");
    setGeneratingJobId(null);
    setIsGenerating(false);
    setGenerationState("idle");
    setGenerationProgress(0);
  };

  // Handle retry/generate another
  const handleRetry = () => {
    console.log("[CREATE] 🔄 Resetting generation state");

    // Clear any polling timeouts
    if (window._createPollTimeout) {
      clearTimeout(window._createPollTimeout);
      window._createPollTimeout = null;
    }

    setGenerationState("idle");
    setGenerationProgress(0);
    setScenes([]);
    setSceneCount(0);
    setCurrentScene(0);
    setGenerationError(null);
    setVideoPreview(null);
    setGeneratedJobId(null);
    setIsGenerating(false);
    setGeneratingJobId(null); // Also clear progress modal
  };

  return (
    <>
      <ProgressModal
        jobId={generatingJobId}
        onComplete={handleProgressComplete}
        onCancel={handleProgressCancel}
        isOpen={!!generatingJobId}
      />

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
              disabled={isGenerating || generationState !== "idle"}
            />

            {/* Media Upload Bar - Below prompt */}
            <MediaUploadBar
              productImage={productImage}
              onProductImageSelect={setProductImage}
              durations={durations}
              selectedDuration={selectedDuration}
              onDurationChange={setSelectedDuration}
              disabled={isGenerating || generationState !== "idle"}
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
              isVisible={[
                "planning",
                "rendering",
                "stitching",
                "ready",
              ].includes(generationState)}
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
              {/* Side Effects - Required field for pharmaceutical ads */}
              <div className="option-group">
                <label className="option-label">Side Effects <span style={{ color: 'var(--error)' }}>*</span></label>
                <textarea
                  className="dropdown-field"
                  placeholder="Enter side effects information (e.g., Common side effects include headache, nausea, dizziness...)"
                  value={sideEffects}
                  onChange={(e) => setSideEffects(e.target.value)}
                  rows={4}
                  required
                  disabled={isGenerating || generationState !== "idle"}
                />
                <p className="option-helper">
                  Required: Enter the side effects information that will be included in your pharmaceutical ad video
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
                      disabled={isGenerating || generationState !== "idle"}
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
                      disabled={isGenerating || generationState !== "idle"}
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
                <label className="option-label">Voice</label>
                <div className="button-group">
                  <button
                    type="button"
                    className={`voice-btn ${
                      voice === "Ash" ? "is-active" : ""
                    }`}
                    onClick={() => setVoice("Ash")}
                    disabled={isGenerating || generationState !== "idle"}
                  >
                    Ash
                  </button>
                  <button
                    type="button"
                    className={`voice-btn ${
                      voice === "Rebecca" ? "is-active" : ""
                    }`}
                    onClick={() => setVoice("Rebecca")}
                    disabled={isGenerating || generationState !== "idle"}
                  >
                    Rebecca
                  </button>
                </div>
                <p className="option-helper">
                  Choose the voice for your video narration
                </p>
              </div>

              {/* Batch Generation - Full width at bottom */}
              <div className="option-group batch-generation-group">
                <BatchGenerationToggle />
              </div>
            </div>
          </div>
        )}
      </section>

      <button
        type="button"
        className="generate-button"
        disabled={!prompt.trim() || !sideEffects.trim() || isGenerating || generationState !== "idle" || !!generatingJobId}
        onClick={handleGenerate}
      >
        {isGenerating ? "Generating..." : "Generate Video"}
      </button>
    </>
  );
}

export default Create;
