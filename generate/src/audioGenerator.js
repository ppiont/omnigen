import Replicate from 'replicate';
import fs from 'fs';
import path from 'path';
import https from 'https';
import http from 'http';

/**
 * Audio generation and video-audio merging
 */
export class AudioGenerator {
  constructor(apiToken, outputDir = './output', openaiClient = null) {
    this.replicate = new Replicate({
      auth: apiToken,
    });
    this.outputDir = path.resolve(outputDir);
    this.tempDir = path.resolve('./temp');
    this.openaiClient = openaiClient;
    this.ensureDirs();
  }

  ensureDirs() {
    [this.outputDir, this.tempDir].forEach(dir => {
      if (!fs.existsSync(dir)) {
        fs.mkdirSync(dir, { recursive: true });
      }
    });
  }

  /**
   * Convert local file to data URI for Replicate API
   */
  fileToDataURI(filePath) {
    const fileBuffer = fs.readFileSync(filePath);
    const base64 = fileBuffer.toString('base64');
    // Determine MIME type based on file extension
    let mimeType = 'audio/mpeg'; // default for mp3

    // Image formats
    if (filePath.endsWith('.png')) mimeType = 'image/png';
    else if (filePath.endsWith('.jpg') || filePath.endsWith('.jpeg')) mimeType = 'image/jpeg';

    // Audio formats (for XTTS-v2 speaker files)
    else if (filePath.endsWith('.wav')) mimeType = 'audio/wav';
    else if (filePath.endsWith('.mp3')) mimeType = 'audio/mpeg';
    else if (filePath.endsWith('.m4a')) mimeType = 'audio/mp4';
    else if (filePath.endsWith('.ogg')) mimeType = 'audio/ogg';
    else if (filePath.endsWith('.flv')) mimeType = 'video/x-flv';

    return `data:${mimeType};base64,${base64}`;
  }

  /**
   * Generate voiceover audio from text using XTTS-v2 (voice cloning TTS)
   */
  async generateVoiceover(text, speakerAudioPath = null, language = 'en') {
    console.log(`\n🎙️  Generating voiceover with XTTS-v2...`);
    console.log(`   Text: "${text.substring(0, 80)}..."`);
    console.log(`   Language: ${language}`);

    try {
      // Use default speaker audio if none provided
      // You should provide a speaker audio file (at least 6 seconds) for voice cloning
      if (!speakerAudioPath) {
        // Check for default speaker file in multiple formats
        const possibleDefaults = [
          path.join(this.tempDir, 'default_speaker.wav'),
          path.join(this.tempDir, 'default_speaker.mp3'),
          path.join(this.tempDir, 'default_speaker.m4a'),
          path.join(this.tempDir, 'default_speaker.ogg'),
          path.join(this.tempDir, 'default_speaker.flv'),
        ];

        for (const defaultPath of possibleDefaults) {
          if (fs.existsSync(defaultPath)) {
            speakerAudioPath = defaultPath;
            break;
          }
        }

        if (!speakerAudioPath) {
          throw new Error('No speaker audio provided. XTTS-v2 requires a speaker audio file (at least 6 seconds) for voice cloning. Use the --speaker option or place a default_speaker file (wav/mp3/m4a/ogg/flv) in the temp directory.');
        }
      }

      if (!fs.existsSync(speakerAudioPath)) {
        throw new Error(`Speaker audio file not found: ${speakerAudioPath}`);
      }

      console.log(`   Speaker: ${speakerAudioPath}`);

      // Convert speaker audio to data URI
      const speakerDataUri = this.fileToDataURI(speakerAudioPath);

      // Create prediction with proper polling
      let prediction = await this.replicate.predictions.create({
        version: "684bc3855b37866c0c65add2ff39c78f3dea3f4ff103a436465326e0f438d55e",
        input: {
          text: text,
          speaker: speakerDataUri,
          language: language,
          cleanup_voice: false,
        }
      });

      // Poll until complete
      while (prediction.status !== 'succeeded' && prediction.status !== 'failed' && prediction.status !== 'canceled') {
        await new Promise(resolve => setTimeout(resolve, 1000));
        prediction = await this.replicate.predictions.get(prediction.id);
      }

      if (prediction.status === 'failed') {
        throw new Error(prediction.error || 'Voiceover generation failed');
      }

      if (prediction.status === 'canceled') {
        throw new Error('Voiceover generation was canceled');
      }

      const output = prediction.output;

      console.log(`   ✅ Voiceover generated`);

      // Extract audio URL from output (XTTS-v2 returns a URI string)
      let audioUrl;
      if (typeof output === 'string') {
        audioUrl = output;
      } else if (Array.isArray(output) && output.length > 0) {
        audioUrl = output[0];
      } else if (output && typeof output === 'object') {
        // Try common property names
        audioUrl = output.audio || output.url || output.output || output.file;
      }

      // Handle URL objects
      if (audioUrl && typeof audioUrl === 'object' && audioUrl.toString) {
        audioUrl = audioUrl.toString();
      }

      if (!audioUrl || typeof audioUrl !== 'string') {
        console.log('   Debug - output:', JSON.stringify(output, null, 2));
        throw new Error('Could not extract audio URL from output');
      }

      console.log(`   Audio URL: ${audioUrl.substring(0, 60)}...`);

      // Download the audio file (XTTS-v2 outputs WAV format)
      const audioPath = path.join(this.tempDir, `voiceover_${Date.now()}.wav`);
      await this.downloadFile(audioUrl, audioPath);

      return audioPath;

    } catch (error) {
      console.error(`   ❌ Error generating voiceover:`, error.message);
      throw error;
    }
  }

  /**
   * Merge video with audio using Replicate's video-audio-merge model
   */
  async mergeVideoWithAudio(videoUrl, audioPath, outputFilename = null) {
    console.log(`\n🎬 Merging video with voiceover...`);
    console.log(`   Video: ${videoUrl.substring(0, 60)}...`);
    console.log(`   Audio: ${audioPath}`);

    try {
      // Convert audio file to data URI (Replicate requires URIs)
      const audioDataUri = this.fileToDataURI(audioPath);
      console.log(`   Converted audio to data URI`);

      // Create prediction with proper polling
      let prediction = await this.replicate.predictions.create({
        version: "8c3d57c9c9a1aaa05feabafbcd2dff9f68a5cb394e54ec020c1c2dcc42bde109",
        input: {
          video_file: videoUrl,
          audio_file: audioDataUri,
        }
      });

      // Poll until complete
      while (prediction.status !== 'succeeded' && prediction.status !== 'failed' && prediction.status !== 'canceled') {
        await new Promise(resolve => setTimeout(resolve, 1000));
        prediction = await this.replicate.predictions.get(prediction.id);
      }

      if (prediction.status === 'failed') {
        throw new Error(prediction.error || 'Video-audio merge failed');
      }

      if (prediction.status === 'canceled') {
        throw new Error('Video-audio merge was canceled');
      }

      const output = prediction.output;

      console.log(`   ✅ Video and audio merged`);

      // Extract video URL from output
      let mergedVideoUrl;
      if (typeof output === 'string') {
        mergedVideoUrl = output;
      } else if (Array.isArray(output) && output.length > 0) {
        mergedVideoUrl = output[0];
      } else if (output && typeof output === 'object') {
        mergedVideoUrl = output.video || output.url || output.output || output.file;
      }

      // Handle URL objects
      if (mergedVideoUrl && typeof mergedVideoUrl === 'object' && mergedVideoUrl.toString) {
        mergedVideoUrl = mergedVideoUrl.toString();
      }

      if (!mergedVideoUrl || typeof mergedVideoUrl !== 'string') {
        console.log('   Debug - output:', JSON.stringify(output, null, 2));
        throw new Error('Could not extract merged video URL from output');
      }

      console.log(`   Merged video URL: ${mergedVideoUrl.substring(0, 60)}...`);

      // Download the final video
      const finalPath = path.join(
        this.outputDir,
        outputFilename || `final_with_audio_${Date.now()}.mp4`
      );
      await this.downloadFile(mergedVideoUrl, finalPath);

      return {
        url: mergedVideoUrl,
        localPath: finalPath,
      };

    } catch (error) {
      console.error(`   ❌ Error merging video and audio:`, error.message);
      throw error;
    }
  }

  /**
   * Generate voiceover script from scene description with narrative arc awareness
   * @param {Object} scene - Scene object with description and structure
   * @param {string} duration - Scene duration
   * @param {Object} openaiClient - Optional OpenAI client for smart script generation
   * @param {number} sceneIndex - Position in sequence (0-based)
   * @param {number} totalScenes - Total number of scenes
   * @param {Array} allScenes - All scenes for context
   * @returns {string} Voiceover script
   */
  async generateVoiceoverScript(scene, duration = '8-10 seconds', openaiClient = null, sceneIndex = 0, totalScenes = 1, allScenes = []) {
    // If OpenAI available, generate proper ad copy with narrative arc
    if (openaiClient) {
      try {
        // Determine narrative position
        const position = totalScenes === 1 ? 'standalone' :
                        sceneIndex === 0 ? 'opening/hook' :
                        sceneIndex === totalScenes - 1 ? 'closing/CTA' :
                        'middle/development';

        // Build context from all scenes
        const narrativeContext = allScenes.length > 1
          ? `\n\nFULL STORY ARC (${totalScenes} clips):\n${allScenes.map((s, i) =>
              `Clip ${i + 1}: ${s.description}`
            ).join('\n')}`
          : '';

        const response = await openaiClient.chat.completions.create({
          model: 'gpt-4o-mini', // Use mini for cost efficiency
          messages: [
            {
              role: 'system',
              content: `You are an expert advertising copywriter. Create compelling voiceover narration for video ads with NARRATIVE ARC PROGRESSION.

CRITICAL RULES - "SHOW DON'T TELL":
- NEVER describe what's visible on screen (e.g., don't say "Kobe is sitting" or "he pours cereal")
- DO focus on emotion, aspiration, brand messaging, and impact
- Write like premium ad copy: short, punchy, memorable
- Use power words and evocative language
- Max 20-25 words (must fit in ${duration})
- Sound natural when spoken aloud
- Create desire, not description

NARRATIVE ARC PROGRESSION (CRITICAL):
- If this is clip 1/opening: HOOK - Pose question, create tension, or intrigue (NO solutions yet)
  Example: "Every champion faces the moment... when focus is everything"

- If this is middle clip: DEVELOP - Build on hook, deepen problem/desire (NO product pitch yet)
  Example: "But when the pressure peaks, most crash. Not you."

- If this is final clip/CTA: RESOLVE - Deliver solution, product benefit, call-to-action
  Example: "CrackFuel. Zero crash. Pure clutch. Grind unstoppable."

- AVOID REPETITION: Each clip should say something DIFFERENT that advances the story
  ❌ BAD: All clips say "Be unstoppable" or "Fuel your greatness"
  ✅ GOOD: Clip 1 = problem, Clip 2 = consequence, Clip 3 = solution, Clip 4 = CTA

GOOD EXAMPLES OF PROGRESSION:
3-Clip Cereal Ad:
  Clip 1: "Every champion knows: greatness starts with the right fuel"
  Clip 2: "The breakfast that fueled legends. Now it's your turn"
  Clip 3: "Winning tastes better when you share it"

4-Clip Energy Drink:
  Clip 1: "Pros don't crash when it counts"
  Clip 2: "But cheap fuel fails when pressure peaks"
  Clip 3: "CrackFuel. Zero sugar. No crash. Pure focus"
  Clip 4: "Grind unstoppable. Twenty percent off now"

Write ONLY the voiceover text for THIS specific clip, nothing else.`
            },
            {
              role: 'user',
              content: `CURRENT CLIP (${sceneIndex + 1}/${totalScenes}):
Position in story: ${position}
Scene: ${scene.description}
Mood: ${scene.structure?.mood || 'engaging'}
${narrativeContext}

Write compelling voiceover for clip ${sceneIndex + 1} that PROGRESSES the narrative (20-25 words max):`
            }
          ],
          temperature: 0.8,
          max_tokens: 50,
        });

        const script = response.choices[0].message.content.trim();
        console.log(`   💬 Generated ad copy [${sceneIndex + 1}/${totalScenes}]: "${script}"`);
        return script;

      } catch (error) {
        console.log(`   ⚠️  OpenAI script generation failed, using fallback`);
        // Fall through to fallback
      }
    }

    // Fallback: Use scene description (only if no OpenAI)
    return scene.description || scene.prompt.substring(0, 200);
  }

  /**
   * Extract language from scene structure or use default
   * XTTS-v2 supports: en, es, fr, de, it, pt, pl, tr, ru, nl, cs, ar, zh, hu, ko, hi
   */
  extractLanguage(scene) {
    // Default to English
    if (!scene.structure || !scene.structure.language) {
      return 'en';
    }

    const lang = scene.structure.language.toLowerCase();

    // Map common language descriptors to valid XTTS-v2 language codes
    const validLanguages = ['en', 'es', 'fr', 'de', 'it', 'pt', 'pl', 'tr', 'ru', 'nl', 'cs', 'ar', 'zh', 'hu', 'ko', 'hi'];

    if (validLanguages.includes(lang)) {
      return lang;
    }

    // Default to English if invalid
    return 'en';
  }

  /**
   * Process entire video sequence with voiceovers using XTTS-v2
   * NOW WITH NARRATIVE ARC AWARENESS!
   */
  async addVoiceoversToSequence(results, scenes, speakerAudioPath = null) {
    console.log(`\n🎤 Adding voiceovers to ${results.length} clips with narrative arc progression...`);

    const finalVideos = [];
    const totalScenes = scenes.length;

    for (let i = 0; i < results.length; i++) {
      const result = results[i];
      const scene = scenes[i];

      if (result.error) {
        console.log(`\n[${i + 1}/${results.length}] ⏭️  Skipping (generation failed)`);
        finalVideos.push(result);
        continue;
      }

      console.log(`\n[${i + 1}/${results.length}]`);

      try {
        // Generate voiceover script with FULL NARRATIVE CONTEXT
        const script = await this.generateVoiceoverScript(
          scene,
          scene.duration,
          this.openaiClient,
          i,              // Scene index
          totalScenes,    // Total scenes for arc awareness
          scenes          // All scenes for context
        );

        // Extract language from scene structure
        const language = this.extractLanguage(scene);

        // Generate audio using XTTS-v2 (voice cloning)
        const audioPath = await this.generateVoiceover(script, speakerAudioPath, language);

        // Merge with video
        const finalVideo = await this.mergeVideoWithAudio(
          result.url,
          audioPath,
          `clip_${i + 1}_with_audio_${Date.now()}.mp4`
        );

        finalVideos.push({
          ...result,
          audioPath: audioPath,
          finalUrl: finalVideo.url,
          finalPath: finalVideo.localPath,
          voiceoverScript: script,
        });

        // Cleanup temp audio
        if (fs.existsSync(audioPath)) {
          fs.unlinkSync(audioPath);
        }

      } catch (error) {
        console.error(`   Failed to add voiceover to clip ${i + 1}`);
        finalVideos.push({
          ...result,
          audioError: error.message,
        });
      }
    }

    return finalVideos;
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
          fs.unlinkSync(outputPath);
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
   * Upload audio to temporary location (for now, just return local path)
   * In production, you'd upload to a cloud storage service
   */
  async uploadAudioToTemp(audioPath) {
    // For now, return the local file path
    // Replicate models may need a public URL, so this might need enhancement
    return audioPath;
  }

  /**
   * Get information about XTTS-v2 voice cloning
   */
  static getAvailableVoices() {
    return {
      'XTTS-v2 Voice Cloning': 'Uses voice cloning - provide a speaker audio file (at least 6 seconds)',
      'Supported Languages': 'en, es, fr, de, it, pt, pl, tr, ru, nl, cs, ar, zh, hu, ko, hi',
      'How to use': 'Use --speaker <path-to-audio-file> to specify the voice to clone',
      'Default': 'Place a default_speaker file (wav/mp3/m4a/ogg/flv) in ./temp/ directory if no speaker is provided',
      'Supported Formats': 'wav, mp3, m4a, ogg, flv',
    };
  }

  /**
   * Get supported languages for XTTS-v2
   */
  static getSupportedLanguages() {
    return [
      { code: 'en', name: 'English' },
      { code: 'es', name: 'Spanish' },
      { code: 'fr', name: 'French' },
      { code: 'de', name: 'German' },
      { code: 'it', name: 'Italian' },
      { code: 'pt', name: 'Portuguese' },
      { code: 'pl', name: 'Polish' },
      { code: 'tr', name: 'Turkish' },
      { code: 'ru', name: 'Russian' },
      { code: 'nl', name: 'Dutch' },
      { code: 'cs', name: 'Czech' },
      { code: 'ar', name: 'Arabic' },
      { code: 'zh', name: 'Mandarin Chinese' },
      { code: 'hu', name: 'Hungarian' },
      { code: 'ko', name: 'Korean' },
      { code: 'hi', name: 'Hindi' },
    ];
  }
}
