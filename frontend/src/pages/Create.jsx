import { useState } from "react";
import { useNavigate } from "react-router-dom";
import ToggleSwitch from "../components/ToggleSwitch.jsx";
import BrandPresetSelector from "../components/create/BrandPresetSelector.jsx";
import MediaUploadBar from "../components/create/MediaUploadBar.jsx";
import ValidationMessage from "../components/create/ValidationMessage.jsx";
import BatchGenerationToggle from "../components/create/BatchGenerationToggle.jsx";
import GenerationState from "../components/create/GenerationState.jsx";
import ScenePreviewGrid from "../components/create/ScenePreviewGrid.jsx";
import { generate, jobs, scripts } from "../utils/api.js";
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
      // STEP 1: Parse prompt into script
      // ============================================
      console.log("\n[CREATE] 📝 STEP 1: Starting script generation (parse)");
      setGenerationState("planning");
      setGenerationProgress(10);

      // Extract product name and audience from prompt (simple extraction)
      // In a real app, you might want to use an LLM or have separate fields
      const productMatch = prompt.match(/\b(for|of|featuring|showcasing)\s+([a-z\s]+)/i);
      const productName = productMatch ? productMatch[2].trim() : "Product";
      const targetAudience = "General audience"; // Could be extracted or made a field

      const parseParams = {
        prompt: prompt.trim(),
        duration: parseInt(selectedDuration),
        product_name: productName,
        target_audience: targetAudience,
        brand_vibe: selectedStyle,
      };

      console.log("[CREATE] 📝 Calling POST /api/v1/parse with params:", parseParams);
      const parseResponse = await scripts.parse(parseParams);
      console.log("[CREATE] ✅ Script generation started:", parseResponse);

      const scriptId = parseResponse.script_id;
      console.log("[CREATE] 📝 Script ID:", scriptId);
      setGenerationProgress(30);

      // ============================================
      // STEP 2: Poll for script completion
      // ============================================
      console.log("\n[CREATE] 🔄 STEP 2: Polling for script completion");
      let script = null;
      let attempts = 0;
      const maxAttempts = 30; // 30 seconds max wait

      while (attempts < maxAttempts) {
        await new Promise(resolve => setTimeout(resolve, 1000));
        attempts++;
        console.log(`[CREATE] 🔄 Polling script status (attempt ${attempts}/${maxAttempts})...`);
        
        try {
          script = await scripts.get(scriptId);
          console.log(`[CREATE] 📄 Script status: ${script.status}`, script);
          
          if (script.status === "draft" || script.status === "ready") {
            console.log("[CREATE] ✅ Script generation completed!");
            break;
          }
          if (script.status === "failed") {
            console.error("[CREATE] ❌ Script generation failed");
            throw new Error("Script generation failed");
          }
        } catch (error) {
          console.warn(`[CREATE] ⚠️ Error fetching script (attempt ${attempts}):`, error);
          // Continue polling on transient errors
          if (attempts >= maxAttempts) {
            throw error;
          }
        }
      }

      if (!script || script.status !== "draft") {
        console.error("[CREATE] ❌ Script generation timed out or failed");
        throw new Error("Script generation timed out or failed");
      }

      // Convert script scenes to our format
      const scenes = script.scenes.map((scene, idx) => ({
        id: idx + 1,
        description: scene.description || scene.prompt || `Scene ${idx + 1}`,
        status: "pending",
        thumbnailUrl: null,
        duration: `${scene.duration || Math.floor(parseInt(selectedDuration) / script.scenes.length)}s`,
      }));

      console.log("[CREATE] 🎬 Scenes extracted:", scenes);
      setScenes(scenes);
      setSceneCount(scenes.length);
      setGenerationProgress(60);
      setGenerationState("rendering");

      // ============================================
      // STEP 3: Generate video from script
      // ============================================
      console.log("\n[CREATE] 🎥 STEP 3: Starting video generation from script");
      console.log("[CREATE] 🎥 Calling POST /api/v1/generate with script_id:", scriptId);
      
      const generateResponse = await generate.create({
        script_id: scriptId,
      });

      const jobId = generateResponse.job_id;
      console.log("[CREATE] ✅ Video generation job created:", generateResponse);
      console.log("[CREATE] 🎥 Job ID:", jobId);
      
      setGeneratedJobId(jobId);
      setGenerationProgress(70);

      // ============================================
      // STEP 4: Poll for job progress
      // ============================================
      console.log("\n[CREATE] 🔄 STEP 4: Polling for job progress");
      let pollCount = 0;
      const pollInterval = setInterval(async () => {
        pollCount++;
        try {
          console.log(`[CREATE] 🔄 Polling job progress (poll #${pollCount})...`);
          const progress = await jobs.progress(jobId);
          console.log(`[CREATE] 📊 Progress update:`, {
            status: progress.status,
            progress: progress.progress,
            current_stage: progress.current_stage,
            stages_completed: progress.stages_completed,
            stages_pending: progress.stages_pending,
          });
          
          setGenerationProgress(Math.min(70 + (progress.progress * 0.3), 100));
          
          // Update scene statuses based on progress
          if (progress.current_stage === "rendering") {
            const updatedScenes = scenes.map((scene, idx) => {
              const sceneProgress = (idx + 1) / scenes.length;
              if (progress.progress / 100 >= sceneProgress) {
                return { ...scene, status: "complete" };
              } else if (progress.progress / 100 >= sceneProgress - 0.1) {
                return { ...scene, status: "rendering" };
              }
              return scene;
            });
            setScenes(updatedScenes);
          }

          if (progress.status === "completed" || progress.status === "ready") {
            console.log("[CREATE] ✅ Video generation completed!");
            clearInterval(pollInterval);
            setGenerationState("ready");
            setGenerationProgress(100);
            
            // Get final job to get video URL
            console.log("[CREATE] 📥 Fetching final job details...");
            const job = await jobs.get(jobId);
            console.log("[CREATE] 📥 Final job data:", job);
            
            if (job.video_url || job.video_key) {
              const videoUrl = job.video_url || `https://your-s3-bucket.s3.amazonaws.com/${job.video_key}`;
              console.log("[CREATE] 🎬 Video URL:", videoUrl);
              setVideoPreview(videoUrl);
            }
            setIsGenerating(false);
          } else if (progress.status === "failed") {
            console.error("[CREATE] ❌ Video generation failed");
            clearInterval(pollInterval);
            setGenerationState("error");
            setGenerationError("Video generation failed");
            setIsGenerating(false);
          }
        } catch (error) {
          console.error(`[CREATE] ⚠️ Error polling progress (poll #${pollCount}):`, error);
          // Continue polling on error
        }
      }, 2000); // Poll every 2 seconds

      // Store interval reference for cleanup
      window._createPollInterval = pollInterval;

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
    
    // Clear any polling intervals
    if (window._createPollInterval) {
      clearInterval(window._createPollInterval);
      window._createPollInterval = null;
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
