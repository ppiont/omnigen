import Replicate from 'replicate';
import { getModel } from './models.js';
import { FrameExtractor } from './frameExtractor.js';
import fs from 'fs';

/**
 * Video generator using Replicate API
 */
export class VideoGenerator {
  constructor(apiToken) {
    this.replicate = new Replicate({
      auth: apiToken,
    });
    this.frameExtractor = new FrameExtractor();
  }

  /**
   * Convert local file to data URI for Replicate API
   */
  fileToDataURI(filePath) {
    const fileBuffer = fs.readFileSync(filePath);
    const base64 = fileBuffer.toString('base64');
    const mimeType = filePath.endsWith('.png') ? 'image/png' : 'image/jpeg';
    return `data:${mimeType};base64,${base64}`;
  }

  /**
   * Generate a single video clip
   * @param {string} modelKey - Model to use (minimax, luma, runway)
   * @param {Object} scene - Scene object with prompt and metadata
   * @param {string} imagePath - Optional path to input image
   * @param {Object} previousScene - Previous scene context for continuity
   * @returns {Promise<Object>} Generation result with video URL
   */
  async generateClip(modelKey, scene, imagePath = null, previousScene = null) {
    const model = getModel(modelKey);

    console.log(`\n🎬 Generating with ${model.name}...`);
    console.log(`   Scene: ${scene.description || 'N/A'}`);
    console.log(`   Prompt: ${scene.prompt.substring(0, 100)}...`);

    // Enhance prompt with previous scene context for continuity
    let enhancedPrompt = scene.prompt;
    if (previousScene) {
      console.log(`   📎 Adding continuity context from previous scene...`);
      const continuityContext = `Continuing from previous scene: "${previousScene.description}". `;
      enhancedPrompt = continuityContext + scene.prompt;
    }

    const input = {
      prompt: enhancedPrompt,
      ...model.defaultParams,
    };

    // Add image if provided and model supports it
    if (imagePath && model.supportsImage) {
      console.log(`   Using input image for continuity: ${imagePath}`);

      // Convert local file paths to data URIs (required by most models)
      let imageUri = imagePath;
      if (imagePath.startsWith('/') || imagePath.startsWith('.')) {
        console.log(`   Converting local file to data URI...`);
        imageUri = this.fileToDataURI(imagePath);
      }

      // Different models use different parameter names
      if (model.id.includes('minimax')) {
        input.first_frame_image = imageUri;
      } else {
        input.image = imageUri;
      }
    }

    try {
      const startTime = Date.now();

      console.log(`\n⏳ Starting generation (typically takes 1-3 minutes)...`);

      // Create prediction with progress tracking
      let prediction = await this.replicate.predictions.create({
        version: await this.getModelVersion(model.id),
        input: input,
      });

      // Poll for completion with progress updates
      let lastStatus = '';
      while (prediction.status !== 'succeeded' && prediction.status !== 'failed' && prediction.status !== 'canceled') {
        await new Promise(resolve => setTimeout(resolve, 2000)); // Poll every 2 seconds

        prediction = await this.replicate.predictions.get(prediction.id);

        const elapsed = ((Date.now() - startTime) / 1000).toFixed(0);

        if (prediction.status !== lastStatus) {
          lastStatus = prediction.status;
          console.log(`   Status: ${prediction.status} (${elapsed}s elapsed)`);
        }

        // Show progress bar if available
        if (prediction.logs) {
          const logs = prediction.logs.split('\n').filter(l => l.trim());
          if (logs.length > 0) {
            const lastLog = logs[logs.length - 1];
            if (lastLog.includes('%') || lastLog.includes('step')) {
              process.stdout.write(`\r   ${lastLog.substring(0, 80)}${' '.repeat(20)}`);
            }
          }
        }
      }

      const duration = ((Date.now() - startTime) / 1000).toFixed(1);

      if (prediction.status === 'failed') {
        throw new Error(prediction.error || 'Generation failed');
      }

      if (prediction.status === 'canceled') {
        throw new Error('Generation was canceled');
      }

      console.log(`\n✅ Generated in ${duration}s`);

      const output = prediction.output;

      // Debug: log the output format
      console.log(`   Output type: ${typeof output}, isArray: ${Array.isArray(output)}`);

      // Handle different output formats
      let videoUrl;

      // Helper function to convert URL object to string
      const extractUrl = (value) => {
        if (typeof value === 'string') {
          return value;
        }
        if (value && typeof value === 'object') {
          // Check if it's a URL object with href
          if (value.href) return value.href;
          if (value.toString && value.toString() !== '[object Object]') {
            const str = value.toString();
            if (str.startsWith('http')) return str;
          }
        }
        return null;
      };

      if (typeof output === 'string') {
        videoUrl = output;
      } else if (Array.isArray(output) && output.length > 0) {
        videoUrl = extractUrl(output[0]) || output[0];
      } else if (output && typeof output === 'object') {
        // Try to extract any URL-like property
        const possibleKeys = ['video', 'url', 'output', 'result', 'file', 'video_url', 'mp4'];
        for (const key of possibleKeys) {
          if (output[key]) {
            const extracted = extractUrl(output[key]);
            if (extracted) {
              videoUrl = extracted;
              console.log(`   Found video URL in property: ${key}`);
              break;
            }
          }
        }
        if (!videoUrl) {
          console.log('   Output object:', JSON.stringify(output, null, 2));
          throw new Error('Unexpected output format from Replicate');
        }
      } else {
        console.log('   Raw output:', output);
        throw new Error('Unexpected output format from Replicate');
      }

      return {
        url: videoUrl,
        scene: scene,
        model: model.name,
        generationTime: duration,
      };
    } catch (error) {
      console.error(`❌ Error generating clip:`, error.message);
      throw error;
    }
  }

  /**
   * Generate multiple video clips in sequence with continuity
   * @param {string} modelKey - Model to use
   * @param {Array} scenes - Array of scene objects
   * @param {string} imagePath - Optional input image (used for first scene only)
   * @param {boolean} keepFrames - Keep extracted frames for inspection
   * @param {Array} keyframes - Pre-generated keyframes for style consistency
   * @returns {Promise<Array>} Array of generation results
   */
  async generateSequence(modelKey, scenes, imagePath = null, keepFrames = false, keyframes = null) {
    if (keyframes && keyframes.length > 0) {
      console.log(`\n🎥 Generating ${scenes.length} clip sequence with KEYFRAME-BASED style consistency...`);
      console.log(`   Using pre-generated keyframes to prevent style drift 🎨\n`);
    } else {
      console.log(`\n🎥 Generating ${scenes.length} clip sequence with continuity...`);
      console.log(`   Each clip will use the last frame of the previous clip for seamless flow\n`);
    }

    const model = getModel(modelKey);
    const results = [];
    let previousResult = null;
    let lastFramePath = null;

    for (let i = 0; i < scenes.length; i++) {
      const scene = scenes[i];
      console.log(`\n[${i + 1}/${scenes.length}]`);

      // Determine input image for this scene
      let inputImage = null;
      let previousScene = null;

      // Set previous scene context for continuity (if not first clip)
      if (i > 0 && previousResult && !previousResult.error) {
        previousScene = previousResult.scene;
      }

      // Priority 1: Use pre-generated keyframe (prevents style drift)
      if (keyframes && keyframes[i] && !keyframes[i].error) {
        inputImage = keyframes[i].localPath;
        console.log(`   🎨 Using keyframe for style consistency`);
        console.log(`   📸 Keyframe: ${inputImage}`);
      }
      // Priority 2: Use user-provided image for first scene
      else if (i === 0 && imagePath) {
        inputImage = imagePath;
      }
      // Priority 3: Extract last frame from previous clip (only if no keyframe)
      else if (previousResult && !previousResult.error && model.supportsImage) {
        try {
          lastFramePath = await this.frameExtractor.getLastFrameFromURL(
            previousResult.url,
            i
          );
          inputImage = lastFramePath;

          console.log(`   ✅ Last frame extracted and ready to use`);
          console.log(`   📸 Frame location: ${lastFramePath}`);
        } catch (error) {
          console.log(`   ⚠️  Could not extract frame, continuing without image continuity...`);
        }
      }

      try {
        const result = await this.generateClip(modelKey, scene, inputImage, previousScene);
        results.push(result);
        previousResult = result;

        // Clean up the extracted frame after successful generation
        if (lastFramePath && i > 0) {
          // Keep the frame for now, will clean up at the end
        }
      } catch (error) {
        console.error(`Failed to generate scene ${i + 1}, continuing...`);
        results.push({
          error: error.message,
          scene: scene,
        });
        previousResult = null; // Reset on error
      }
    }

    // Cleanup extracted frames (unless user wants to keep them)
    if (keepFrames) {
      console.log(`\n📁 Keeping extracted frames in: ${this.frameExtractor.tempDir}`);
      console.log(`   You can inspect these frames to see what was sent to the next clip`);
    } else {
      console.log(`\n🧹 Cleaning up temporary frames...`);
      this.frameExtractor.cleanup();
    }

    return results;
  }

  /**
   * Get prediction status (for long-running generations)
   */
  async getStatus(predictionId) {
    try {
      const prediction = await this.replicate.predictions.get(predictionId);
      return prediction;
    } catch (error) {
      console.error('Error getting prediction status:', error.message);
      throw error;
    }
  }

  /**
   * Get the latest version ID for a model
   */
  async getModelVersion(modelId) {
    try {
      const [owner, name] = modelId.split('/');
      const model = await this.replicate.models.get(owner, name);

      if (model.latest_version && model.latest_version.id) {
        return model.latest_version.id;
      }

      // Fallback: fetch versions and get the first one
      const versions = await this.replicate.models.versions.list(owner, name);
      if (versions.results && versions.results.length > 0) {
        return versions.results[0].id;
      }

      throw new Error(`Could not find version for model ${modelId}`);
    } catch (error) {
      console.error(`Error getting model version for ${modelId}:`, error.message);
      throw error;
    }
  }
}
