import Replicate from 'replicate';
import fs from 'fs';
import path from 'path';
import https from 'https';
import http from 'http';

/**
 * Generate keyframes to ensure style consistency across video sequence
 */
export class KeyframeGenerator {
  constructor(apiToken, outputDir = './output') {
    this.replicate = new Replicate({
      auth: apiToken,
    });
    this.outputDir = path.resolve(outputDir);
    this.keyframeDir = path.resolve('./keyframes');
    this.ensureDirs();
  }

  ensureDirs() {
    [this.outputDir, this.keyframeDir].forEach(dir => {
      if (!fs.existsSync(dir)) {
        fs.mkdirSync(dir, { recursive: true });
      }
    });
  }

  /**
   * Generate keyframes for all scenes using image generation
   * This ensures consistent style across the sequence
   */
  async generateKeyframes(scenes, styleReference = null) {
    console.log(`\n🎨 Generating keyframes for ${scenes.length} scenes to ensure style consistency...`);

    if (styleReference) {
      console.log(`   Using style reference: ${styleReference}`);
    }

    const keyframes = [];

    for (let i = 0; i < scenes.length; i++) {
      const scene = scenes[i];
      console.log(`\n[Keyframe ${i + 1}/${scenes.length}]`);
      console.log(`   Scene: ${scene.description}`);

      try {
        const keyframe = await this.generateKeyframe(scene, styleReference, i);
        keyframes.push(keyframe);
      } catch (error) {
        console.error(`   ❌ Failed to generate keyframe ${i + 1}:`, error.message);
        keyframes.push({ error: error.message, scene });
      }
    }

    return keyframes;
  }

  /**
   * Generate a single keyframe image
   */
  async generateKeyframe(scene, styleReference = null, index = 0) {
    console.log(`   Generating keyframe image...`);

    // Build prompt for keyframe
    let keyframePrompt = this.buildKeyframePrompt(scene);

    // Add style consistency instruction
    if (styleReference) {
      keyframePrompt = `In the exact same visual style as the reference image. ${keyframePrompt}`;
    }

    const input = {
      prompt: keyframePrompt,
      width: 1024,
      height: 576,
      num_outputs: 1,
      num_inference_steps: 25, // SDXL default
      guidance_scale: 7.5,
    };

    // Add style reference if provided
    if (styleReference) {
      // Convert local file to data URI if it's a local path
      if (styleReference.startsWith('/') || styleReference.startsWith('.')) {
        input.image = await this.fileToDataURI(styleReference);
      } else {
        input.image = styleReference;
      }
      input.prompt_strength = 0.7; // Balance between prompt and style reference
    }

    try {
      // Use SDXL for reliable keyframe generation
      console.log('   Starting keyframe generation...');

      let prediction = await this.replicate.predictions.create({
        version: "39ed52f2a78e934b3ba6e2a89f5b1c712de7dfea535525255b1aa35c5565e08b",
        input: input
      });

      // Poll until complete
      while (prediction.status !== 'succeeded' && prediction.status !== 'failed') {
        await new Promise(resolve => setTimeout(resolve, 1000));
        prediction = await this.replicate.predictions.get(prediction.id);
      }

      if (prediction.status === 'failed') {
        throw new Error(prediction.error || 'Keyframe generation failed');
      }

      const output = prediction.output;

      let imageUrl;
      if (typeof output === 'string') {
        imageUrl = output;
      } else if (Array.isArray(output) && output.length > 0) {
        imageUrl = output[0];
      } else if (output && typeof output === 'object') {
        // Handle URL object or other formats
        imageUrl = output.toString();
      } else {
        console.log('   Debug - output type:', typeof output);
        console.log('   Debug - output:', JSON.stringify(output));
        throw new Error('Unexpected output format from image model');
      }

      // Ensure imageUrl is a string
      if (typeof imageUrl !== 'string') {
        throw new Error('Could not extract image URL from output');
      }

      console.log(`   ✅ Keyframe generated`);

      // Download keyframe
      const keyframePath = path.join(this.keyframeDir, `keyframe_${index + 1}.jpg`);
      await this.downloadFile(imageUrl, keyframePath);

      return {
        url: imageUrl,
        localPath: keyframePath,
        scene: scene,
        index: index,
      };

    } catch (error) {
      console.error(`   ❌ Error generating keyframe:`, error.message);
      throw error;
    }
  }

  /**
   * Build optimized prompt for keyframe generation
   */
  buildKeyframePrompt(scene) {
    // Extract key visual elements from scene
    let prompt = scene.prompt;

    // Add keyframe-specific instructions
    const keyframeInstructions = [
      'single frame',
      'high detail',
      'cinematic composition',
      'professional photography',
      'sharp focus',
      '4K quality'
    ].join(', ');

    // Combine
    return `${prompt}. ${keyframeInstructions}`;
  }

  /**
   * Generate style reference from user's concept
   * Creates a "master" style image that all keyframes will reference
   */
  async generateStyleReference(conceptPrompt, userStyleImage = null) {
    console.log(`\n🎨 Generating master style reference...`);

    if (userStyleImage) {
      console.log(`   Using user-provided style image: ${userStyleImage}`);
      return userStyleImage;
    }

    console.log(`   Creating style reference from concept...`);

    const stylePrompt = `${conceptPrompt}. Consistent art style, cohesive visual aesthetic, professional, cinematic, high quality reference image.`;

    const input = {
      prompt: stylePrompt,
      width: 1024,
      height: 576,
      num_outputs: 1,
      num_inference_steps: 30, // Higher for quality
      guidance_scale: 7.5,
    };

    try {
      // Create prediction and wait for completion
      console.log('   Starting image generation...');

      let prediction = await this.replicate.predictions.create({
        version: "39ed52f2a78e934b3ba6e2a89f5b1c712de7dfea535525255b1aa35c5565e08b",
        input: input
      });

      // Poll until complete
      while (prediction.status !== 'succeeded' && prediction.status !== 'failed') {
        await new Promise(resolve => setTimeout(resolve, 1000));
        prediction = await this.replicate.predictions.get(prediction.id);
        console.log(`   Status: ${prediction.status}`);
      }

      if (prediction.status === 'failed') {
        throw new Error(prediction.error || 'Image generation failed');
      }

      const output = prediction.output;
      console.log('   Generation complete!');

      // Enhanced URL extraction with better debugging
      let imageUrl;

      if (typeof output === 'string') {
        imageUrl = output;
      } else if (Array.isArray(output)) {
        if (output.length > 0) {
          const firstItem = output[0];

          // Check if it's a FileOutput object with url property
          if (firstItem && typeof firstItem === 'object' && firstItem.url) {
            imageUrl = firstItem.url;
          }
          // Check if it's already a string
          else if (typeof firstItem === 'string') {
            imageUrl = firstItem;
          }
          // Try toString on the object
          else if (firstItem && firstItem.toString && typeof firstItem.toString === 'function') {
            const strValue = firstItem.toString();
            if (strValue.startsWith('http')) {
              imageUrl = strValue;
            }
          }
        }
      } else if (output && typeof output === 'object') {
        // Check for direct url property
        if (output.url) {
          imageUrl = output.url;
        }
      }

      // Ensure imageUrl is a valid string
      if (!imageUrl || typeof imageUrl !== 'string' || !imageUrl.startsWith('http')) {
        console.log('   ❌ Could not extract valid URL from output');
        console.log('   Output structure:', JSON.stringify(output, null, 2));
        throw new Error('Could not extract image URL from output');
      }

      console.log(`   ✅ Style reference generated`);

      // Download and save
      const stylePath = path.join(this.keyframeDir, 'style_reference.jpg');
      await this.downloadFile(imageUrl, stylePath);

      return stylePath;

    } catch (error) {
      console.error(`   ❌ Error generating style reference:`, error.message);
      throw error;
    }
  }

  /**
   * Convert local file to data URI for Replicate API
   */
  async fileToDataURI(filePath) {
    const fileBuffer = fs.readFileSync(filePath);
    const base64 = fileBuffer.toString('base64');
    const mimeType = filePath.endsWith('.png') ? 'image/png' : 'image/jpeg';
    return `data:${mimeType};base64,${base64}`;
  }

  /**
   * Download file from URL
   */
  async downloadFile(url, outputPath) {
    return new Promise((resolve, reject) => {
      const file = fs.createWriteStream(outputPath);
      const protocol = url.startsWith('https') ? https : http;

      protocol.get(url, (response) => {
        if (response.statusCode === 200) {
          response.pipe(file);
          file.on('finish', () => {
            file.close();
            console.log(`   📥 Downloaded: ${path.basename(outputPath)}`);
            resolve(outputPath);
          });
        } else if (response.statusCode === 302 || response.statusCode === 301) {
          file.close();
          fs.unlinkSync(outputPath);
          this.downloadFile(response.headers.location, outputPath)
            .then(resolve)
            .catch(reject);
        } else {
          file.close();
          if (fs.existsSync(outputPath)) {
            fs.unlinkSync(outputPath);
          }
          reject(new Error(`Failed to download: HTTP ${response.statusCode}`));
        }
      }).on('error', (err) => {
        file.close();
        if (fs.existsSync(outputPath)) {
          fs.unlinkSync(outputPath);
        }
        reject(err);
      });
    });
  }

  /**
   * Cleanup keyframe directory
   */
  cleanup() {
    if (fs.existsSync(this.keyframeDir)) {
      const files = fs.readdirSync(this.keyframeDir);
      files.forEach(file => {
        fs.unlinkSync(path.join(this.keyframeDir, file));
      });
      console.log(`🧹 Cleaned up keyframes`);
    }
  }
}
