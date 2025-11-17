import { useState, useEffect, useRef } from "react";
import { useNavigate, useLocation } from "react-router-dom";
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
const durations = ["10s", "20s", "30s", "40s", "50s", "60s"]; // Must be multiple of 10
const aspects = ["16:9", "9:16", "1:1"]; // Backend only supports these

// Phase 1: Enhanced prompt options
const visualStyles = ["cinematic", "documentary", "energetic", "minimal", "dramatic", "playful"];
const tones = ["premium", "friendly", "edgy", "inspiring", "humorous"];
const tempos = ["slow", "medium", "fast"];
const platforms = ["instagram", "tiktok", "youtube", "facebook"];
const goals = ["awareness", "sales", "engagement", "signups"];

/**
 * Maps a style from preset to brand preset ID
 * @param {string} style - The preset style
 * @returns {string} Brand preset ID
 */
function mapStyleToBrandPreset(style) {
  const mapping = {
    "Minimalist": "tech-minimal",
    "Bold": "bold-vibrant",
    "Modern": "corporate-clean",
    "Cinematic": "default",
    "Playful": "warm-organic",
  };
  return mapping[style] || "default";
}

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
  const location = useLocation();
  const presetAppliedRef = useRef(false);
  
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

  // Phase 1: Enhanced prompt options (all optional)
  const [visualStyle, setVisualStyle] = useState("");
  const [tone, setTone] = useState("");
  const [tempo, setTempo] = useState("");
  const [platform, setPlatform] = useState("");
  const [audience, setAudience] = useState("");
  const [goal, setGoal] = useState("");
  const [callToAction, setCallToAction] = useState("");
  const [proCinematography, setProCinematography] = useState(false);
  const [creativeBoost, setCreativeBoost] = useState(false);

  const characterLimit = 2000;
  const characterCount = prompt.length;

  // Apply preset from location state on mount
  useEffect(() => {
    // Only apply preset once on initial mount
    if (presetAppliedRef.current) {
      return;
    }

    const preset = location.state?.preset;
    if (!preset) {
      presetAppliedRef.current = true;
      return;
    }

    // Validate and map preset style to valid Create page style
    const validStyle = styles.includes(preset.style) 
      ? preset.style 
      : "Cinematic"; // Fallback to default

    // Apply preset settings
    setSelectedStyle(validStyle);
    setSelectedBrandPreset(mapStyleToBrandPreset(validStyle));
    setSelectedCategory("Ad Creative"); // Default as specified
    setSelectedDuration("30s"); // Default as specified
    setSelectedAspect("16:9"); // Default as specified
    // Leave prompt empty for user input

    presetAppliedRef.current = true;
  }, [location.state]);

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

      // Add enhanced prompt options (Phase 1 - all optional)
      if (visualStyle) generateParams.style = visualStyle;
      if (tone) generateParams.tone = tone;
      if (tempo) generateParams.tempo = tempo;
      if (platform) generateParams.platform = platform;
      if (audience) generateParams.audience = audience;
      if (goal) generateParams.goal = goal;
      if (callToAction) generateParams.call_to_action = callToAction;
      if (proCinematography) generateParams.pro_cinematography = true;
      if (creativeBoost) generateParams.creative_boost = true;

      // Add style_reference_image (🎨 Reference Style Img - guides style for ALL clips)
      if (referenceImage) {
        // Use preview (data URI) if available, otherwise skip
        if (
          referenceImage.preview &&
          referenceImage.preview.startsWith("data:image/")
        ) {
          generateParams.style_reference_image = referenceImage.preview;
          console.log("[CREATE] 🎨 Using style reference image (data URI)");
        } else if (
          referenceImage.url &&
          (referenceImage.url.startsWith("http://") ||
            referenceImage.url.startsWith("https://"))
        ) {
          generateParams.style_reference_image = referenceImage.url;
          console.log("[CREATE] 🎨 Using style reference image (URL)");
        } else {
          console.log(
            "[CREATE] ⚠️ Style reference image provided but not a valid URL, skipping"
          );
        }
      }

      // Add start_image (🖼️ Starting Img - used ONLY for first scene)
      if (templateImage) {
        // Use preview (data URI) if available, otherwise skip
        if (
          templateImage.preview &&
          templateImage.preview.startsWith("data:image/")
        ) {
          generateParams.start_image = templateImage.preview;
          console.log("[CREATE] 🖼️ Using start image for first scene (data URI)");
        } else if (
          templateImage.url &&
          (templateImage.url.startsWith("http://") ||
            templateImage.url.startsWith("https://"))
        ) {
          generateParams.start_image = templateImage.url;
          console.log("[CREATE] 🖼️ Using start image for first scene (URL)");
        } else {
          console.log(
            "[CREATE] ⚠️ Start image provided but not a valid URL, skipping"
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

      setGeneratedJobId(jobId);
      setGenerationProgress(5); // Initial progress

      // ============================================
      // Poll for job status
      // ============================================
      console.log("\n" + "=".repeat(80));
      console.log("📊 [POLLING] Starting job status polling");
      console.log("=".repeat(80));
      console.log("[POLLING] Polling GET /api/v1/jobs/:id every 7 seconds");

      let pollCount = 0;
      const progressPollInterval = 7000; // 7 seconds between polls (stays under 10/min limit)
      const maxPollAttempts = 300; // ~35 minutes max (300 * 7s = 2100s = 35min)
      const startTime = Date.now();
      let progressPollTimeoutRef = null;

      const pollProgress = async () => {
        pollCount++;
        const elapsedTime = Date.now() - startTime;

        // Check max polling attempts
        if (pollCount > maxPollAttempts) {
          console.error(
            `[CREATE] ⚠️ Max polling attempts (${maxPollAttempts}) reached. Job may be stuck.`
          );
          if (progressPollTimeoutRef) clearTimeout(progressPollTimeoutRef);
          setGenerationState("error");
          setGenerationError(
            "Video generation is taking longer than expected. Please check back later or try again."
          );
          setIsGenerating(false);
          return;
        }

        try {
          console.log(
            `[CREATE] 🔄 Polling job status (poll #${pollCount}, elapsed: ${Math.round(
              elapsedTime / 1000
            )}s)...`
          );

          // Get job status directly (progress endpoint returns 501)
          const job = await jobs.get(jobId);

          console.log(`[CREATE] 📊 Job status update:`, {
            status: job.status,
            stage: job.stage,
            progress_percent: job.progress_percent,
            metadata: job.metadata,
          });

          // Update progress from backend's calculated progress_percent
          setGenerationProgress(job.progress_percent || 0);

          // Update scene information from metadata if available
          if (job.metadata) {
            if (job.metadata.num_scenes) {
              setSceneCount(job.metadata.num_scenes);
            }
            if (job.metadata.current_scene) {
              setCurrentScene(job.metadata.current_scene);
            }
            if (job.metadata.scenes_complete !== undefined) {
              setCurrentScene(job.metadata.scenes_complete);
            }
          }

          // Update generation state based on stage
          if (job.stage) {
            if (job.stage === "script_generating") {
              setGenerationState("planning");
            } else if (
              job.stage.startsWith("scene_") ||
              job.stage === "audio_generating" ||
              job.stage === "composing"
            ) {
              setGenerationState("rendering");
            }
          }

          if (job.status === "completed") {
            console.log("[CREATE] ✅ Video generation completed!");
            if (progressPollTimeoutRef) clearTimeout(progressPollTimeoutRef);
            setGenerationState("ready");
            setGenerationProgress(100);

            if (job.video_url) {
              console.log("[CREATE] 🎬 Video URL:", job.video_url);
              setVideoPreview(job.video_url);
            } else if (job.video_key) {
              // Fallback: construct URL from key (though backend should provide video_url)
              const videoUrl = `https://your-s3-bucket.s3.amazonaws.com/${job.video_key}`;
              console.log("[CREATE] 🎬 Video URL (constructed):", videoUrl);
              setVideoPreview(videoUrl);
            }
            setIsGenerating(false);
          } else if (job.status === "failed") {
            console.error("[CREATE] ❌ Video generation failed");
            if (progressPollTimeoutRef) clearTimeout(progressPollTimeoutRef);
            setGenerationState("error");
            setGenerationError(job.error_message || "Video generation failed");
            setIsGenerating(false);
          } else {
            // Continue polling
            progressPollTimeoutRef = setTimeout(
              pollProgress,
              progressPollInterval
            );
          }
        } catch (error) {
          // Handle authentication errors - refresh token and retry
          if (error.status === 401) {
            console.warn(
              `[CREATE] 🔄 Authentication expired during polling (poll #${pollCount}). Refreshing token...`
            );
            try {
              // Refresh the session via backend
              await auth.refresh();
              console.log(
                `[CREATE] ✅ Token refreshed successfully. Retrying poll...`
              );
              // Retry immediately after refresh
              progressPollTimeoutRef = setTimeout(pollProgress, 1000);
              return;
            } catch (refreshError) {
              console.error(
                `[CREATE] ❌ Failed to refresh token:`,
                refreshError
              );
              setError(
                "Your session has expired. Please log in again and retry."
              );
              setIsGenerating(false);
              return;
            }
          }

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
              {/* Brand Preset Selector - Commented out for now */}
              {/* <BrandPresetSelector
                selectedPreset={selectedBrandPreset}
                onChange={setSelectedBrandPreset}
              /> */}

              {/* Category Selector - Commented out for now */}
              {/* <div className="option-group">
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
              </div> */}

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

              {/* Phase 1: Enhanced Prompt Options */}
              <div className="option-group">
                <label className="option-label">Visual Style (Optional)</label>
                <select
                  className="dropdown-field"
                  value={visualStyle}
                  onChange={(e) => setVisualStyle(e.target.value)}
                >
                  <option value="">Default</option>
                  {visualStyles.map((style) => (
                    <option key={style} value={style}>
                      {style.charAt(0).toUpperCase() + style.slice(1)}
                    </option>
                  ))}
                </select>
                <p className="option-helper">
                  Choose the overall visual aesthetic (cinematic, documentary, etc.)
                </p>
              </div>

              <div className="option-group">
                <label className="option-label">Tone (Optional)</label>
                <select
                  className="dropdown-field"
                  value={tone}
                  onChange={(e) => setTone(e.target.value)}
                >
                  <option value="">Default</option>
                  {tones.map((t) => (
                    <option key={t} value={t}>
                      {t.charAt(0).toUpperCase() + t.slice(1)}
                    </option>
                  ))}
                </select>
                <p className="option-helper">
                  Set the emotional tone (premium, friendly, inspiring, etc.)
                </p>
              </div>

              <div className="option-group">
                <label className="option-label">Platform (Optional)</label>
                <select
                  className="dropdown-field"
                  value={platform}
                  onChange={(e) => setPlatform(e.target.value)}
                >
                  <option value="">Default</option>
                  {platforms.map((p) => (
                    <option key={p} value={p}>
                      {p.charAt(0).toUpperCase() + p.slice(1)}
                    </option>
                  ))}
                </select>
                <p className="option-helper">
                  Optimize for specific platform (Instagram, TikTok, YouTube, Facebook)
                </p>
              </div>

              <div className="option-group">
                <label className="option-label">Marketing Goal (Optional)</label>
                <select
                  className="dropdown-field"
                  value={goal}
                  onChange={(e) => setGoal(e.target.value)}
                >
                  <option value="">Default</option>
                  {goals.map((g) => (
                    <option key={g} value={g}>
                      {g.charAt(0).toUpperCase() + g.slice(1)}
                    </option>
                  ))}
                </select>
                <p className="option-helper">
                  Campaign objective (awareness, sales, engagement, signups)
                </p>
              </div>

              <div className="option-group">
                <label className="option-label">Target Audience (Optional)</label>
                <input
                  type="text"
                  className="dropdown-field"
                  placeholder="e.g., Tech-savvy millennials, 25-35"
                  value={audience}
                  onChange={(e) => setAudience(e.target.value)}
                  maxLength={200}
                />
                <p className="option-helper">
                  Describe your target demographic
                </p>
              </div>

              <div className="option-group">
                <label className="option-label">Call to Action (Optional)</label>
                <input
                  type="text"
                  className="dropdown-field"
                  placeholder="e.g., Shop Now, Learn More, Sign Up Today"
                  value={callToAction}
                  onChange={(e) => setCallToAction(e.target.value)}
                  maxLength={100}
                />
                <p className="option-helper">
                  Custom call-to-action text
                </p>
              </div>

              <div className="option-group">
                <label className="option-label">Advanced Features</label>
                <div className="toggle-group">
                  <div className="toggle-item">
                    <ToggleSwitch
                      checked={proCinematography}
                      onChange={() => setProCinematography(!proCinematography)}
                      label="Pro Cinematography"
                    />
                    <span className="toggle-label">Professional Film Techniques</span>
                  </div>
                  <div className="toggle-item">
                    <ToggleSwitch
                      checked={creativeBoost}
                      onChange={() => setCreativeBoost(!creativeBoost)}
                      label="Creative Boost"
                    />
                    <span className="toggle-label">Boost Creativity (Higher Temperature)</span>
                  </div>
                </div>
                <p className="option-helper">
                  Use advanced cinematography terms and boost creative output
                </p>
              </div>

              {/* Options - Commented out (not currently implemented in backend) */}
              {/* <div className="option-group">
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
              </div> */}

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
