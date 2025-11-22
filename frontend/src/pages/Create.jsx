import { useEffect, useMemo, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import ToggleSwitch from "../components/ToggleSwitch.jsx";
import BrandPresetSelector from "../components/create/BrandPresetSelector.jsx";
import MediaUploadBar from "../components/create/MediaUploadBar.jsx";
import ValidationMessage from "../components/create/ValidationMessage.jsx";
import BatchGenerationToggle from "../components/create/BatchGenerationToggle.jsx";
import GenerationState from "../components/create/GenerationState.jsx";
import ScenePreviewGrid from "../components/create/ScenePreviewGrid.jsx";
import { useJobProgress } from "../hooks/useJobProgress.js";
import { APIError, generate, uploads } from "../utils/api.js";
import { showToast } from "../utils/toast.js";
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
// Valid durations must be achievable with Veo 3.1 clips (4s, 6s, or 8s each)
// 10=4+6, 16=8+8, 20=4+8+8, 24=8+8+8, 30=6+6+6+6+6, 40=8+8+8+8+8, 60=many combos
const durations = ["10s", "16s", "20s", "24s", "30s", "40s", "60s"];
const aspects = ["16:9", "9:16", "1:1"]; // Backend only supports these

const SIDE_EFFECTS_MIN = 10;
const SIDE_EFFECTS_MAX = 500;

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
  const [productImageAssetUrl, setProductImageAssetUrl] = useState(null);
  const [productUploadStatus, setProductUploadStatus] = useState("idle"); // idle | uploading | success | error
  const [productUploadError, setProductUploadError] = useState("");
  const [productUploadAttempt, setProductUploadAttempt] = useState(0);
  const uploadAbortControllerRef = useRef(null);
  const [validationError, setValidationError] = useState("");


  // Phase 2 additions - State Machine
  const [generationState, setGenerationState] = useState("idle"); // idle, rendering, completed, error
  const [scenes, setScenes] = useState([]);
  const [videoPreview, setVideoPreview] = useState(null);
  const [generationError, setGenerationError] = useState(null);
  const [generatedJobId, setGeneratedJobId] = useState(null);

  // Job progress tracking with SSE
  const [activeJobId, setActiveJobId] = useState(null);

  const jobProgress = useJobProgress(activeJobId, {
    onComplete: (finalProgress) => {
      console.log("[CREATE] ✅ Job completed:", finalProgress);

      // Validate that video URL is available
      if (!finalProgress?.assets?.final_video?.url) {
        console.error("[CREATE] ❌ Video completed but URL not available");
        setGenerationState("error");
        setGenerationError("Video generation completed but video URL is not available. Please try again.");
        setIsGenerating(false);
        return;
      }

      setGenerationState("completed");
      setIsGenerating(false);
      setVideoPreview(finalProgress.assets.final_video.url);
    },
    onFailed: (finalProgress) => {
      console.error("[CREATE] ❌ Job failed:", finalProgress);

      // Extract detailed error message from job if available
      let errorMessage = "Video generation failed. Please try again.";
      if (finalProgress?.error_message) {
        errorMessage = finalProgress.error_message;
        console.error("[CREATE] Error details:", finalProgress.error_message);
      } else if (finalProgress?.message) {
        errorMessage = finalProgress.message;
      }

      setGenerationState("error");
      setGenerationError(errorMessage);
      setIsGenerating(false);
    }
  });

  // Handle product image upload to S3 via presigned URL
  useEffect(() => {
    const file = productImage?.file;

    // Reset state when image removed
    if (!file) {
      setProductUploadStatus("idle");
      setProductUploadError("");
      setProductImageAssetUrl(null);
      if (uploadAbortControllerRef.current) {
        uploadAbortControllerRef.current.abort();
        uploadAbortControllerRef.current = null;
      }
      return;
    }

    let isCancelled = false;
    const abortController = new AbortController();
    uploadAbortControllerRef.current = abortController;

    const uploadProductImage = async () => {
      setProductUploadStatus("uploading");
      setProductUploadError("");
      setProductImageAssetUrl(null);

      try {
        console.log("[CREATE] 📤 Requesting presigned URL for product image upload");
        const presignResponse = await uploads.getPresignedUrl({
          type: "product_image",
          filename: file.name,
          contentType: file.type,
          fileSize: file.size,
        });

        if (isCancelled) {
          return;
        }

        const { upload_url: uploadUrl, asset_url: assetUrl } = presignResponse || {};
        if (!uploadUrl || !assetUrl) {
          throw new Error("Invalid presigned URL response from server.");
        }

        console.log("[CREATE] 🚀 Uploading product image to S3");
        const uploadResponse = await fetch(uploadUrl, {
          method: "PUT",
          body: file,
          headers: {
            "Content-Type": file.type || "application/octet-stream",
          },
          signal: abortController.signal,
        });

        if (isCancelled) {
          return;
        }

        if (!uploadResponse.ok) {
          const errorText = await uploadResponse.text().catch(() => uploadResponse.statusText);
          throw new Error(
            `Upload failed with status ${uploadResponse.status}: ${errorText || "Unknown error"}`
          );
        }

        setProductImageAssetUrl(assetUrl);
        setProductUploadStatus("success");
        console.log("[CREATE] ✅ Product image uploaded successfully:", assetUrl);
      } catch (error) {
        if (isCancelled || error.name === "AbortError") {
          console.warn("[CREATE] ⚠️ Product image upload aborted");
          return;
        }

        console.error("[CREATE] ❌ Product image upload failed:", error);
        setProductUploadStatus("error");
        setProductUploadError(
          error.message || "Failed to upload product image. Please try again."
        );
        setProductImageAssetUrl(null);
      } finally {
        if (uploadAbortControllerRef.current === abortController) {
          uploadAbortControllerRef.current = null;
        }
      }
    };

    uploadProductImage();

    return () => {
      isCancelled = true;
      abortController.abort();
    };
  }, [productImage?.file, productUploadAttempt]);

  // Voice selection
  const [voice, setVoice] = useState("Ash"); // Default to Ash (male)
  // Side Effects - Required field for pharmaceutical ads
  const [sideEffects, setSideEffects] = useState("");

  // Voice mapping: UI → Backend
  const voiceMap = useMemo(
    () => ({
      Ash: "male",
      Rebecca: "female",
    }),
    []
  );

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
  const sideEffectsLimit = SIDE_EFFECTS_MAX;
  const sideEffectsCount = sideEffects.length;
  const sideEffectsTrimmedCount = sideEffects.trim().length;
  const isSideEffectsValid =
    sideEffectsTrimmedCount >= SIDE_EFFECTS_MIN &&
    sideEffectsTrimmedCount <= SIDE_EFFECTS_MAX;
  const isGenerateDisabled = useMemo(() => {
    if (!prompt.trim()) return true;
    if (!isSideEffectsValid) return true;
    if (isGenerating) return true;
    if (generationState !== "idle") return true;
    if (activeJobId) return true;
    if (productUploadStatus === "uploading") return true;
    return false;
  }, [
    prompt,
    isSideEffectsValid,
    isGenerating,
    generationState,
    activeJobId,
    productUploadStatus,
  ]);

  const getEstimatedTime = () => {
    const durationNum = parseInt(selectedDuration);
    if (durationNum <= 15) return "~30s";
    if (durationNum <= 30) return "~45s";
    if (durationNum <= 60) return "1 min";
    return "1-2 min";
  };

  // Cost calculation constants (based on pipeline.md)
  const COST_SCRIPT_GPT4O = 0.20;        // Fixed cost for script generation
  const COST_VIDEO_VEO_PER_SECOND = 0.07; // Veo 3.1 cost per second
  const COST_AUDIO_MINIMAX = 0.50;       // Fixed cost for background music
  const COST_TTS_OPENAI = 0.10;          // Fixed cost for narrator TTS (pharmaceutical ads only)
  const COST_STORAGE_S3 = 0.01;          // Estimated S3 storage cost

  // Calculate estimated cost based on user selections (updates in real-time)
  const estimatedCost = useMemo(() => {
    const durationNum = parseInt(selectedDuration) || 30;
    
    // Check if this is a pharmaceutical ad (has voice and side effects)
    const isPharmaceuticalAd = voice && sideEffects.trim().length >= SIDE_EFFECTS_MIN;
    
    // Calculate costs
    const scriptCost = COST_SCRIPT_GPT4O;
    const videoCost = durationNum * COST_VIDEO_VEO_PER_SECOND;
    const audioCost = COST_AUDIO_MINIMAX;
    const ttsCost = isPharmaceuticalAd ? COST_TTS_OPENAI : 0;
    const storageCost = COST_STORAGE_S3;
    
    const totalCost = scriptCost + videoCost + audioCost + ttsCost + storageCost;
    
    return {
      total: totalCost,
      breakdown: {
        script: scriptCost,
        video: videoCost,
        audio: audioCost,
        tts: ttsCost,
        storage: storageCost,
      },
      isPharmaceuticalAd,
    };
  }, [selectedDuration, voice, sideEffects]);

  const getEstimatedCost = () => {
    return `$${estimatedCost.total.toFixed(2)}`;
  };

  const handleGenerate = async () => {
    console.log("=".repeat(80));
    console.log("🎬 [CREATE] VIDEO GENERATION PIPELINE STARTED");
    console.log("=".repeat(80));
    const trimmedPrompt = prompt.trim();
    const trimmedSideEffects = sideEffects.trim();
    const backendVoice = voiceMap[voice] || "male";

    console.log("[CREATE] 📝 User Input:", {
      prompt: trimmedPrompt,
      category: selectedCategory,
      style: selectedStyle,
      duration: selectedDuration,
      aspectRatio: selectedAspect,
      brandPreset: selectedBrandPreset,
      voice: voice,
      backendVoice,
      sideEffectsLength: trimmedSideEffects.length,
    });

    // Validation
    if (!trimmedPrompt) {
      console.warn("[CREATE] ⚠️ Validation failed: Empty prompt");
      setValidationError("Please describe your video to get started");
      return;
    }

    // Validate side effects (required for pharmaceutical ads)
    if (!trimmedSideEffects) {
      console.warn("[CREATE] ⚠️ Validation failed: Empty side effects");
      setValidationError("Side Effects is required. Please enter the side effects information.");
      return;
    }

    if (trimmedSideEffects.length < SIDE_EFFECTS_MIN) {
      console.warn("[CREATE] ⚠️ Validation failed: Side effects too short");
      setValidationError(`Side effects must be at least ${SIDE_EFFECTS_MIN} characters.`);
      return;
    }

    if (trimmedSideEffects.length > SIDE_EFFECTS_MAX) {
      console.warn("[CREATE] ⚠️ Validation failed: Side effects too long");
      setValidationError(`Side effects cannot exceed ${SIDE_EFFECTS_MAX} characters (currently: ${trimmedSideEffects.length}).`);
      return;
    }

    if (!productImage) {
      console.warn("[CREATE] ⚠️ Validation failed: Missing product image");
      setValidationError("Product image is required for pharmaceutical ads.");
      return;
    }

    if (productUploadStatus === "uploading") {
      console.warn("[CREATE] ⚠️ Validation failed: Product image still uploading");
      setValidationError("Product image is still uploading. Please wait for the upload to complete.");
      return;
    }

    if (isGenerating) {
      console.warn("[CREATE] ⚠️ Already generating, ignoring request");
      return;
    }

    // Validate duration is achievable with Veo 3.1 clips (4s, 6s, or 8s)
    const durationNum = parseInt(selectedDuration);
    const validDurations = [10, 16, 20, 24, 30, 40, 60];
    if (!validDurations.includes(durationNum)) {
      console.warn(
        "[CREATE] ⚠️ Validation failed: Duration must be valid for Veo 3.1"
      );
      setValidationError("Duration must be 10, 16, 20, 24, 30, 40, or 60 seconds");
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

      // Prepare generate request with required fields
      const generateParams = {
        prompt: trimmedPrompt,
        duration: durationNum,
        aspect_ratio: selectedAspect,
      };


      // Add style (pharmaceutical ad styles) - MAP to backend-compatible style
      if (selectedStyle) {
        const backendStyle = styleMap[selectedStyle] || 'documentary'; // Default to documentary
        generateParams.style = backendStyle;
        console.log(`[CREATE] Mapped style "${selectedStyle}" to backend style "${backendStyle}"`);
      }

      if (voice) {
        generateParams.voice = backendVoice;
        console.log(`[CREATE] 🎙️ Voice selection mapped to "${backendVoice}"`);
      }

      if (trimmedSideEffects) {
        generateParams.side_effects = trimmedSideEffects;
        console.log(`[CREATE] 💊 Side effects included (${trimmedSideEffects.length} characters)`);
      }

      // Add start_image (Product Image - used ONLY for first scene)
      if (productImageAssetUrl) {
        generateParams.start_image = productImageAssetUrl;
        console.log("[CREATE] 📸 Using uploaded product image asset:", productImageAssetUrl);
      } else if (productImage?.preview) {
        // Fallback to preview data URI if presigned upload not available
        generateParams.start_image = productImage.preview;
        console.warn("[CREATE] ⚠️ Falling back to inline product image preview (upload URL unavailable)");
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

      // Store job ID and start SSE tracking
      setGeneratedJobId(jobId);
      setActiveJobId(jobId); // Start progress tracking
    } catch (error) {
      console.error("\n" + "=".repeat(80));
      console.error("❌ [ERROR] VIDEO GENERATION PIPELINE ERROR");
      console.error("=".repeat(80));
      console.error("[ERROR] Generation failed:", error);

      let message = error?.message || "Video generation failed. Please try again.";

      if (error instanceof APIError) {
        message = error.message || message;

        if (error.status === 400) {
          setGenerationState("idle");
          setIsGenerating(false);

          if (error.details?.field) {
            console.warn(`[CREATE] Validation error on field "${error.details.field}": ${message}`);
            setValidationError(message);
          } else {
            setValidationError("");
            setGenerationError(message);
          }

          showToast(message, "error");
          return;
        }
      }

      setGenerationState("error");
      setValidationError("");
      setGenerationError(message);
      showToast(message, "error");
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

  const getSideEffectsCounterClass = () => {
    if (sideEffectsTrimmedCount >= sideEffectsLimit) return "char-counter danger";
    if (sideEffectsTrimmedCount >= sideEffectsLimit - 9) return "char-counter danger";
    if (sideEffectsTrimmedCount >= sideEffectsLimit - 50) return "char-counter warning";
    return "char-counter normal";
  };

  const retryProductUpload = () => {
    if (!productImage?.file) {
      return;
    }

    console.log("[CREATE] 🔄 Retrying product image upload");
    setProductUploadStatus("idle");
    setProductUploadError("");
    setProductImageAssetUrl(null);
    setProductUploadAttempt((attempt) => attempt + 1);
  };

  const toggleAdvanced = () => setIsAdvancedOpen(!isAdvancedOpen);

  // Handle viewing completed video
  const handleViewVideo = () => {
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

  // Handle retry/generate another
  const handleRetry = () => {
    console.log("[CREATE] 🔄 Resetting generation state");

    setGenerationState("idle");
    setScenes([]);
    setGenerationError(null);
    setVideoPreview(null);
    setGeneratedJobId(null);
    setIsGenerating(false);
    setActiveJobId(null); // Stop SSE connection
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
              disabled={isGenerating || generationState !== "idle"}
            />

            {/* Media Upload Bar - Below prompt */}
            <MediaUploadBar
              productImage={productImage}
              onProductImageSelect={setProductImage}
              durations={durations}
              selectedDuration={selectedDuration}
              onDurationChange={setSelectedDuration}
              uploadStatus={productUploadStatus}
              uploadError={productUploadError}
              onRetryUpload={retryProductUpload}
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
              jobProgress={jobProgress}
              error={generationError}
              videoPreview={videoPreview}
              onRetry={handleRetry}
              onViewVideo={handleViewVideo}
              aspectRatio={selectedAspect}
            />

            {/* Scene Preview Grid - show during RENDERING and COMPLETED */}
            <ScenePreviewGrid
              scenes={scenes}
              isVisible={["rendering", "completed"].includes(generationState)}
              jobProgress={jobProgress}
            />
          </div>

          {/* Estimation - only show when idle */}
          {generationState === "idle" && (
            <div className="estimation-section">
              <div className="estimation-grid">
                <div className="estimation-item">
                  <span className="estimation-label">Estimated time</span>
                  <span className="estimation-value">{getEstimatedTime()}</span>
                </div>
                <div className="estimation-item">
                  <span className="estimation-label">Estimated cost</span>
                  <span 
                    className="estimation-value cost-total" 
                    title={`Script: $${estimatedCost.breakdown.script.toFixed(2)} | Video: $${estimatedCost.breakdown.video.toFixed(2)} | Audio: $${estimatedCost.breakdown.audio.toFixed(2)}${estimatedCost.isPharmaceuticalAd ? ` | TTS: $${estimatedCost.breakdown.tts.toFixed(2)}` : ''} | Storage: $${estimatedCost.breakdown.storage.toFixed(2)}`}
                  >
                    {getEstimatedCost()}
                  </span>
                </div>
              </div>
              
              {/* Cost Breakdown - Expandable */}
              <details className="cost-breakdown">
                <summary className="cost-breakdown-summary">
                  View cost breakdown
                </summary>
                <div className="cost-breakdown-list">
                  <div className="cost-breakdown-item">
                    <span className="cost-item-label">Script Generation (GPT-4o)</span>
                    <span className="cost-item-value">
                      ${estimatedCost.breakdown.script.toFixed(2)}
                    </span>
                  </div>
                  <div className="cost-breakdown-item">
                    <span className="cost-item-label">
                      Video Generation (Veo 3.1)
                      <span className="cost-item-detail">
                        {parseInt(selectedDuration)}s × ${COST_VIDEO_VEO_PER_SECOND.toFixed(2)}/s
                      </span>
                    </span>
                    <span className="cost-item-value">
                      ${estimatedCost.breakdown.video.toFixed(2)}
                    </span>
                  </div>
                  <div className="cost-breakdown-item">
                    <span className="cost-item-label">Background Music (Minimax)</span>
                    <span className="cost-item-value">
                      ${estimatedCost.breakdown.audio.toFixed(2)}
                    </span>
                  </div>
                  {estimatedCost.isPharmaceuticalAd && (
                    <div className="cost-breakdown-item">
                      <span className="cost-item-label">Narrator Voiceover (OpenAI TTS)</span>
                      <span className="cost-item-value">
                        ${estimatedCost.breakdown.tts.toFixed(2)}
                      </span>
                    </div>
                  )}
                  <div className="cost-breakdown-item">
                    <span className="cost-item-label">Storage (S3)</span>
                    <span className="cost-item-value">
                      ${estimatedCost.breakdown.storage.toFixed(2)}
                    </span>
                  </div>
                  <div className="cost-breakdown-item cost-breakdown-total">
                    <span className="cost-item-label">Total</span>
                    <span className="cost-item-value">
                      ${estimatedCost.total.toFixed(2)}
                    </span>
                  </div>
                </div>
              </details>
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
                <label className="option-label">
                  Side Effects <span style={{ color: 'var(--error)' }}>*</span>
                  <span className={getSideEffectsCounterClass()}>
                    {sideEffectsTrimmedCount} / {SIDE_EFFECTS_MAX}
                  </span>
                </label>
                <textarea
                  className="prompt-textarea"
                  placeholder="Enter side effects information (e.g., Common side effects include headache, nausea, dizziness...)"
                  value={sideEffects}
                  onChange={(e) => {
                    setSideEffects(e.target.value);
                    if (validationError) setValidationError("");
                  }}
                  rows={4}
                  required
                  disabled={isGenerating || generationState !== "idle"}
                  maxLength={SIDE_EFFECTS_MAX}
                />
                <p className="option-helper">
                  Required: 10-500 characters (trimmed). This text appears in narration and on-screen disclosures.
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
        disabled={isGenerateDisabled}
        onClick={handleGenerate}
      >
        {isGenerating ? "Generating..." : "Generate Video"}
      </button>
    </>
  );
}

export default Create;
