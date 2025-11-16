/**
 * Mock data and utilities for simulating video generation flow
 */

// Delay helper
export const delay = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

// Scene description templates
const sceneTemplates = {
  "Ad Creative": [
    "Opening shot showcasing product in elegant setting",
    "Close-up highlighting key features and details",
    "Lifestyle shot showing product in use",
    "Final branding moment with call-to-action",
  ],
  "Product Demo": [
    "Wide shot introducing the product",
    "Step-by-step demonstration of main features",
    "Detail focus on unique selling points",
    "Summary shot with product benefits",
  ],
  Explainer: [
    "Problem statement introduction",
    "Solution presentation with visuals",
    "Step-by-step breakdown",
    "Conclusion with key takeaways",
  ],
  "Social Media": [
    "Attention-grabbing hook shot",
    "Quick product showcase",
    "Lifestyle integration",
    "Strong CTA with branding",
  ],
  Tutorial: [
    "Introduction to topic",
    "Step 1 demonstration",
    "Step 2 demonstration",
    "Final result showcase",
  ],
};

/**
 * Generate mock scene data based on prompt and config
 */
export const generateMockScenes = (prompt, config) => {
  const { duration, category } = config;

  // Calculate number of clips based on duration
  const durationNum = parseInt(duration);
  const numClips =
    durationNum <= 15 ? 1 : durationNum <= 30 ? 2 : durationNum <= 60 ? 3 : 4;

  // Get templates for category
  const templates = sceneTemplates[category] || sceneTemplates["Ad Creative"];

  // Calculate duration per clip
  const clipDuration = Math.floor(durationNum / numClips);

  // Generate scenes
  return Array.from({ length: numClips }, (_, idx) => ({
    id: idx + 1,
    description:
      templates[idx % templates.length] ||
      `Scene ${idx + 1} from: ${prompt.substring(0, 50)}...`,
    status: "pending",
    thumbnailUrl: null,
    duration: `${clipDuration}s`,
  }));
};

/**
 * Mock description generator based on prompt keywords
 */
export const enhanceSceneDescription = (prompt, sceneIndex, totalScenes) => {
  const keywords = prompt.toLowerCase();

  // Extract product name if mentioned
  const productMatch = prompt.match(/\b(for|of|featuring|showcasing)\s+([a-z\s]+)/i);
  const product = productMatch ? productMatch[2].trim() : "product";

  // Scene-specific enhancements
  if (sceneIndex === 0) {
    return `Opening shot introducing ${product} with atmospheric lighting`;
  } else if (sceneIndex === totalScenes - 1) {
    return `Closing shot with ${product} branding and call-to-action`;
  } else {
    return `Mid-sequence highlighting ${product} features and benefits`;
  }
};

/**
 * Simulate video generation flow with state updates
 */
export const simulateGeneration = async (
  prompt,
  config,
  onStateChange,
  onSceneUpdate
) => {
  try {
    // State 1: PENDING (2 seconds)
    onStateChange({ state: "pending", progress: 0 });
    await delay(2000);

    // State 2: PLANNING (3 seconds)
    onStateChange({ state: "planning", progress: 30 });
    await delay(1500);

    const scenes = generateMockScenes(prompt, config);
    onSceneUpdate(scenes);

    onStateChange({
      state: "planning",
      progress: 60,
      sceneCount: scenes.length,
    });
    await delay(1500);

    // State 3: RENDERING (2.5 seconds per clip)
    onStateChange({
      state: "rendering",
      progress: 0,
      sceneCount: scenes.length,
      currentScene: 0,
    });

    let updatedScenes = [...scenes];
    for (let i = 0; i < scenes.length; i++) {
      // Update scene to rendering
      updatedScenes[i] = { ...updatedScenes[i], status: "rendering" };
      onSceneUpdate(updatedScenes);

      await delay(1000);

      // Mark scene as complete
      updatedScenes[i] = { ...updatedScenes[i], status: "complete" };
      onSceneUpdate(updatedScenes);

      const progressPercent = ((i + 1) / scenes.length) * 100;
      onStateChange({
        state: "rendering",
        progress: progressPercent,
        sceneCount: scenes.length,
        currentScene: i + 1,
      });

      await delay(1500);
    }

    // State 4: STITCHING (only if multiple clips)
    if (scenes.length > 1) {
      onStateChange({
        state: "stitching",
        progress: 0,
        sceneCount: scenes.length,
      });
      await delay(1000);

      onStateChange({
        state: "stitching",
        progress: 50,
        sceneCount: scenes.length,
      });
      await delay(1500);

      onStateChange({
        state: "stitching",
        progress: 100,
        sceneCount: scenes.length,
      });
      await delay(1000);
    }

    // State 5: READY
    onStateChange({
      state: "ready",
      progress: 100,
      sceneCount: scenes.length,
      videoPreview: getMockThumbnail(config.style),
    });

    return {
      success: true,
      scenes,
    };
  } catch (error) {
    onStateChange({
      state: "error",
      error: error.message || "Generation failed",
    });
    return { success: false, error: error.message };
  }
};

/**
 * Get mock thumbnail based on style
 */
export const getMockThumbnail = (style) => {
  // Using placeholder images (you can replace with actual thumbnails later)
  const thumbnails = {
    Cinematic: "https://via.placeholder.com/640x360/7cff00/0a0e1a?text=Cinematic+Video",
    Modern: "https://via.placeholder.com/640x360/00ffd1/0a0e1a?text=Modern+Video",
    Minimalist: "https://via.placeholder.com/640x360/b44cff/0a0e1a?text=Minimal+Video",
    Bold: "https://via.placeholder.com/640x360/ff00ff/0a0e1a?text=Bold+Video",
    Playful: "https://via.placeholder.com/640x360/ffa500/0a0e1a?text=Playful+Video",
  };

  return thumbnails[style] || thumbnails.Cinematic;
};

/**
 * Reset generation to initial state
 */
export const resetGeneration = () => ({
  state: "idle",
  progress: 0,
  scenes: [],
  sceneCount: 0,
  currentScene: 0,
  error: null,
  videoPreview: null,
});
