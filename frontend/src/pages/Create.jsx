import { useState } from "react";
import { useNavigate } from "react-router-dom";
import ToggleSwitch from "../components/ToggleSwitch.jsx";
import BrandPresetSelector from "../components/create/BrandPresetSelector.jsx";
import MediaUploadBar from "../components/create/MediaUploadBar.jsx";
import ValidationMessage from "../components/create/ValidationMessage.jsx";
import BatchGenerationToggle from "../components/create/BatchGenerationToggle.jsx";
import GenerationState from "../components/create/GenerationState.jsx";
import ScenePreviewGrid from "../components/create/ScenePreviewGrid.jsx";
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
    console.log("=".repeat(60));
    console.log("[CREATE] 🎬 User clicked Generate Video button");
    console.log("[CREATE] Prompt:", prompt);
    console.log("[CREATE] Config:", {
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

    if (isGenerating) {
      console.warn("[CREATE] ⚠️ Already generating, ignoring request");
      return;
    }

    // Clear validation error and reset state
    setValidationError("");
    setIsGenerating(true);
    setGenerationError(null);

    try {
      // ============================================
      // STAGE 3: VIDEO GENERATION (VideoGenerator)
      // ============================================
      console.log("\n" + "=".repeat(80));
      console.log(
        "🎥 [STAGE 3] VIDEO GENERATION - Generating video clips with visual continuity"
      );
      console.log("=".repeat(80));
      console.log(
        "[STAGE 3] Purpose: Generate video clips with visual continuity"
      );
      console.log("[STAGE 3] Process: For each scene:");
      console.log(
        "[STAGE 3]   1. Determine input image (last frame from previous clip)"
      );
      console.log("[STAGE 3]   2. Call Replicate API (Minimax/PixVerse/etc.)");
      console.log("[STAGE 3]   3. Poll for completion (~60-180s per clip)");
      console.log("[STAGE 3]   4. Extract last frame for next clip");

      setGenerationState("rendering");
      setGenerationProgress(10);

      // Prepare generate request with required fields
      const generateParams = {
        prompt: prompt.trim(),
        duration: parseInt(selectedDuration),
        aspect_ratio: selectedAspect,
      };

      // Add optional fields if provided
      if (referenceImage) {
        generateParams.start_image = referenceImage.url || referenceImage.name;
      }

      console.log("[STAGE 3] 📡 API Call: POST /api/v1/generate");
      console.log("[STAGE 3] 📦 Request payload:", generateParams);

      const generateResponse = await generate.create(generateParams);

      const jobId = generateResponse.job_id;
      console.log("[STAGE 3] ✅ Video generation job created");
      console.log("[STAGE 3] 🆔 Job ID:", jobId);
      console.log("[STAGE 3] 📊 Status:", generateResponse.status);
      console.log(
        "[STAGE 3] 🎬 Number of clips:",
        generateResponse.num_clips || Math.floor(parseInt(selectedDuration) / 5)
      );
      console.log(
        "[STAGE 3] ⏱️ Estimated completion:",
        generateResponse.estimated_completion_seconds || "N/A",
        "seconds"
      );

      setGeneratedJobId(jobId);
      setGenerationProgress(40);

      // ============================================
      // STAGE 4: VOICEOVER GENERATION (Optional - Handled by backend)
      // ============================================
      console.log("\n" + "=".repeat(80));
      console.log(
        "🎤 [STAGE 4] VOICEOVER GENERATION - Processing (if enabled)"
      );
      console.log("=".repeat(80));
      console.log("[STAGE 4] Purpose: Add professional narration to videos");
      console.log("[STAGE 4] Process: (Handled by backend Step Functions)");
      console.log("[STAGE 4]   1. Script generation (GPT-4o-mini)");
      console.log("[STAGE 4]   2. Text-to-Speech (Minimax Speech 02 HD)");
      console.log("[STAGE 4]   3. Video-Audio merge");

      // ============================================
      // STAGE 5-7: Poll for job progress (covers Download, Stitching, Metadata)
      // ============================================
      console.log("\n" + "=".repeat(80));
      console.log("📊 [STAGES 5-7] POLLING FOR PROGRESS");
      console.log("=".repeat(80));
      console.log("[STAGE 5] Download & Output: Save all generated content");
      console.log(
        "[STAGE 6] Video Stitching: Combine multiple clips into single video"
      );
      console.log("[STAGE 7] Metadata Export: Export scene structure");
      console.log("[POLLING] Starting progress polling...");

      let pollCount = 0;
      const progressPollInterval = 7000; // 7 seconds between polls (stays under 10/min limit)
      let progressPollTimeoutRef = null;

      const pollProgress = async () => {
        pollCount++;
        try {
          console.log(
            `[CREATE] 🔄 Polling job progress (poll #${pollCount})...`
          );
          const progress = await jobs.progress(jobId);
          console.log(`[CREATE] 📊 Progress update:`, {
            status: progress.status,
            progress: progress.progress,
            current_stage: progress.current_stage,
            stages_completed: progress.stages_completed,
            stages_pending: progress.stages_pending,
          });

          setGenerationProgress(Math.min(40 + progress.progress * 0.6, 100));

          if (progress.status === "completed" || progress.status === "ready") {
            console.log("[CREATE] ✅ Video generation completed!");
            if (progressPollTimeoutRef) clearTimeout(progressPollTimeoutRef);
            setGenerationState("ready");
            setGenerationProgress(100);

            // Get final job to get video URL
            console.log("[CREATE] 📥 Fetching final job details...");
            const job = await jobs.get(jobId);
            console.log("[CREATE] 📥 Final job data:", job);

            if (job.video_url || job.video_key) {
              const videoUrl =
                job.video_url ||
                `https://your-s3-bucket.s3.amazonaws.com/${job.video_key}`;
              console.log("[CREATE] 🎬 Video URL:", videoUrl);
              setVideoPreview(videoUrl);
            }
            setIsGenerating(false);
          } else if (progress.status === "failed") {
            console.error("[CREATE] ❌ Video generation failed");
            if (progressPollTimeoutRef) clearTimeout(progressPollTimeoutRef);
            setGenerationState("error");
            setGenerationError("Video generation failed");
            setIsGenerating(false);
          } else {
            // Continue polling
            progressPollTimeoutRef = setTimeout(
              pollProgress,
              progressPollInterval
            );
          }
        } catch (error) {
          // Handle rate limit errors
          if (error.status === 429) {
            const retryAfter = error.details?.reset_in || 60;
            console.warn(
              `[CREATE] ⚠️ Rate limit hit during progress polling. Waiting ${retryAfter} seconds...`
            );
            // Wait and retry
            progressPollTimeoutRef = setTimeout(
              pollProgress,
              retryAfter * 1000
            );
            return;
          }

          console.error(
            `[CREATE] ⚠️ Error polling progress (poll #${pollCount}):`,
            error
          );
          // Continue polling on other errors after delay
          progressPollTimeoutRef = setTimeout(
            pollProgress,
            progressPollInterval
          );
        }
      };

      // Start polling
      progressPollTimeoutRef = setTimeout(pollProgress, progressPollInterval);
      window._createPollTimeout = progressPollTimeoutRef;
    } catch (error) {
      console.error("[CREATE] ❌ Generation failed:", error);
      setGenerationState("error");
      setGenerationError(error.message || "Generation failed");
      setIsGenerating(false);
    }

    console.log("=".repeat(60));
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
          referenceImage: referenceImage?.name || null,
          templateImage: templateImage?.name || null,
        },
        generatedAt: Date.now(),
      },
    });
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
