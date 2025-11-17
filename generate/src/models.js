/**
 * Available video generation models on Replicate
 * Updated with verified working model IDs (2025)
 */
export const VIDEO_MODELS = {
  minimax: {
    name: 'Minimax Hailuo 2.3',
    id: 'minimax/hailuo-2.3',
    duration: '6-10 seconds',
    supportsImage: true,
    cost: '$0.28-$0.56 per clip',
    defaultParams: {
      prompt_optimizer: true,
      duration: 10,
      resolution: '768p',
    },
    promptGuidance: `Minimax Hailuo 2.3 works best with:
- Natural, conversational prompts describing realistic scenes
- Focus on realistic motion and natural physics
- Effective at generating human subjects and facial expressions
- Supports Chinese and English prompts
- Works well with everyday scenarios and realistic settings`
  },
  wan: {
    name: 'Wan 2.5 T2V (Fast)',
    id: 'wan-video/wan-2.5-t2v-fast',
    duration: '5-8 seconds',
    supportsImage: false,
    cost: 'Very cheap',
    defaultParams: {
      fps: 24,
    },
    promptGuidance: `Wan 2.5 T2V (Fast) works best with:
- Fast generation, good for rapid iteration
- Simple, direct prompts work well
- Best for straightforward motion and action
- Budget-friendly option for testing concepts`
  },
  pixverse: {
    name: 'PixVerse v5',
    id: 'pixverse/pixverse-v5',
    duration: '5-8 seconds',
    supportsImage: true,
    cost: 'Cheap (~$0.30/clip)',
    defaultParams: {
      resolution: '720p',
    },
    promptGuidance: `PixVerse v5 works best with:
- Artistic and stylized content
- Creative visual effects and transitions
- Good balance of quality and cost
- Supports both realistic and stylized aesthetics`
  },
  seedance: {
    name: 'Seedance Pro (Fast)',
    id: 'bytedance/seedance-1-pro-fast',
    duration: '5-10 seconds',
    supportsImage: true,
    cost: 'Budget-friendly',
    defaultParams: {},
    promptGuidance: `Seedance Pro (Fast) works best with:
- Fast, efficient generation
- Good for product videos and commercial content
- Handles smooth camera movements well
- Effective with image-to-video conversions`
  },
  veo: {
    name: 'Google Veo 3.1 (Fast)',
    id: 'google/veo-3.1-fast',
    duration: '8 seconds',
    supportsImage: true,
    cost: 'Premium',
    defaultParams: {},
    promptGuidance: `Google Veo 3.1 (Fast) works best with:
- Highly realistic, photorealistic content
- Complex scenes with multiple elements
- Natural physics and realistic motion
- Excellent lighting and shadow accuracy
- Professional-grade cinematic quality`
  },
  kling: {
    name: 'Kling V2.5 Turbo Pro',
    id: 'kwaivgi/kling-v2.5-turbo-pro',
    duration: '5-10 seconds',
    supportsImage: true,
    cost: 'Premium',
    defaultParams: {
      duration: 10,
      aspect_ratio: '16:9'
    },
    promptGuidance: `Kling V2.5 Turbo Pro excels with:

PROMPT STRUCTURE:
- Use multi-step, causal instructions (e.g., "A bird lands on a branch, then tilts its head")
- Include sequential actions with clear cause-and-effect relationships
- Specify detailed camera movements (dolly, pan, zoom, tracking shots)
- Describe high-speed action and dynamic motion

VISUAL DETAILS:
- Provide specific lighting descriptions (golden hour, soft diffused, dramatic side-lighting)
- Include mood and atmosphere (serene, tense, energetic, mysterious)
- When using start images: maintain color, lighting, brushwork, and style consistency
- Add visual context and environmental details for better results

BEST FOR:
- Complex camera movements and cinematography
- High-speed action sequences
- Brand-consistent content (maintains visual style well)
- Professional cinematic quality with refined image conditioning
- Multi-step narrative sequences with coherent action flow`
  }
};

/**
 * Get model configuration by key
 */
export function getModel(modelKey) {
  const model = VIDEO_MODELS[modelKey];
  if (!model) {
    throw new Error(`Unknown model: ${modelKey}. Available: ${Object.keys(VIDEO_MODELS).join(', ')}`);
  }
  return model;
}

/**
 * List all available models
 */
export function listModels() {
  return Object.entries(VIDEO_MODELS).map(([key, model]) => ({
    key,
    name: model.name,
    duration: model.duration,
    supportsImage: model.supportsImage,
    cost: model.cost,
  }));
}
