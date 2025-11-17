#!/usr/bin/env node

import { Command } from 'commander';
import dotenv from 'dotenv';
import OpenAI from 'openai';
import { PromptParser } from './src/promptParser.js';
import { VideoGenerator } from './src/videoGenerator.js';
import { VideoDownloader } from './src/downloader.js';
import { XMLExporter } from './src/xmlExporter.js';
import { AudioGenerator } from './src/audioGenerator.js';
import { KeyframeGenerator } from './src/keyframeGenerator.js';
import { VideoStitcher } from './src/videoStitcher.js';
import { listModels } from './src/models.js';
import fs from 'fs';

dotenv.config();

const program = new Command();

program
  .name('video-gen')
  .description('AI-powered cinematic video generator with story continuity and structured output')
  .version('1.0.0');

program
  .command('generate')
  .description('Generate cinematic video sequence with narrative flow')
  .argument('<prompt>', 'Story, concept, or scene to visualize')
  .option('-m, --model <model>', 'Model to use: minimax, luma, or runway', 'minimax')
  .option('-i, --image <path>', 'Input image path (optional)')
  .option('-c, --clips <number>', 'Number of clips in sequence (default: 1)', '1')
  .option('-o, --output <dir>', 'Output directory for downloads', './output')
  .option('--no-download', 'Skip downloading videos, just show URLs')
  .option('--no-parse', 'Skip OpenAI prompt parsing, use raw prompt')
  .option('--keep-frames', 'Keep extracted frames for inspection (in ./temp/)')
  .option('--voiceover', 'Add AI voiceover narration to videos using XTTS-v2 voice cloning')
  .option('--speaker <path>', 'Speaker audio file for voice cloning (at least 6 seconds, wav/mp3/m4a/ogg/flv)')
  .option('--language <lang>', 'Language for voiceover (en, es, fr, de, it, pt, pl, tr, ru, nl, cs, ar, zh, hu, ko, hi)', 'en')
  .option('--style-ref <path>', 'Style reference image to prevent style drift across clips')
  .option('--keyframes', 'Generate keyframes first for consistent style (prevents drift)')
  .option('--style <style>', 'Creative style: cinematic, documentary, energetic, minimal, dramatic, playful', 'cinematic')
  .option('--tone <tone>', 'Advertisement tone: premium, friendly, edgy, inspiring, humorous', 'premium')
  .option('--tempo <tempo>', 'Pacing: slow, medium, fast', 'medium')
  .option('--creative', 'Boost creativity (more unexpected, artistic interpretations)')
  .option('--audience <audience>', 'Target audience (e.g., "Gen Z gamers", "busy professionals 25-40")')
  .option('--goal <goal>', 'Marketing goal: awareness, sales, engagement, signups', 'awareness')
  .option('--cta <text>', 'Call-to-action text (e.g., "Shop Now", "Learn More")')
  .option('--platform <platform>', 'Optimize for platform: instagram, tiktok, youtube, facebook', 'youtube')
  .option('--pro-cinematography', 'Use advanced film terminology for professional camera work')
  .option('--smooth-transitions', 'Use frame interpolation for smoother transitions between clips')
  .option('--crossfade <duration>', 'Add crossfade transitions between clips (seconds, e.g., 1.0)', parseFloat)
  .option('--interpolation-fps <fps>', 'Target FPS for frame interpolation (default: 60)', '60')
  .action(async (prompt, options) => {
    try {
      console.log('\n🎬 Cinematic Video Generator\n');
      console.log('━'.repeat(50));

      // Validate API keys
      if (!process.env.REPLICATE_API_TOKEN) {
        console.error('❌ Error: REPLICATE_API_TOKEN not found in .env file');
        process.exit(1);
      }

      if (options.parse && !process.env.OPENAI_API_KEY) {
        console.error('⚠️  Warning: OPENAI_API_KEY not found, falling back to simple parsing');
        options.parse = false;
      }

      // Validate image file if provided
      if (options.image && !fs.existsSync(options.image)) {
        console.error(`❌ Error: Image file not found: ${options.image}`);
        process.exit(1);
      }

      // Validate speaker audio file if provided
      if (options.voiceover && options.speaker && !fs.existsSync(options.speaker)) {
        console.error(`❌ Error: Speaker audio file not found: ${options.speaker}`);
        process.exit(1);
      }

      const numClips = parseInt(options.clips);
      if (isNaN(numClips) || numClips < 1) {
        console.error('❌ Error: Number of clips must be a positive integer');
        process.exit(1);
      }

      // Step 1: Parse and plan content
      let scenes;
      if (options.parse) {
        const parser = new PromptParser(process.env.OPENAI_API_KEY);

        // Get model-specific guidance
        const { getModel } = await import('./src/models.js');
        const modelConfig = getModel(options.model);
        const modelGuidance = modelConfig.promptGuidance || null;

        // Pass creative direction to parser (including new marketing parameters)
        const creativeOptions = {
          style: options.style,
          tone: options.tone,
          tempo: options.tempo,
          creative: options.creative,
          audience: options.audience,
          goal: options.goal,
          cta: options.cta,
          platform: options.platform,
          proCinematography: options.proCinematography,
        };

        scenes = await parser.planContent(prompt, numClips, creativeOptions, modelGuidance);
      } else {
        console.log('\n📝 Using raw prompt (OpenAI parsing disabled)\n');
        scenes = PromptParser.simpleParse(prompt, numClips);
      }

      // Export scene structure to XML
      const xmlExporter = new XMLExporter(options.output);
      xmlExporter.exportToXML(scenes, prompt, {
        model: options.model,
        total_duration: `${scenes.length * 8}-${scenes.length * 10} seconds`
      });

      // Display planned scenes
      console.log('\n📋 Planned Scenes:');
      console.log('━'.repeat(50));
      scenes.forEach((scene, i) => {
        console.log(`\n[Scene ${i + 1}]`);
        console.log(`Description: ${scene.description}`);
        console.log(`Prompt: ${scene.prompt.substring(0, 100)}...`);
        console.log(`Duration: ${scene.duration}`);

        // Show structure if available
        if (scene.structure) {
          console.log(`\nStructure:`);
          console.log(`  Camera: ${scene.structure.camera || 'N/A'}`);
          console.log(`  Lighting: ${scene.structure.lighting || 'N/A'}`);
          console.log(`  Mood: ${scene.structure.mood || 'N/A'}`);
          if (scene.structure.transition) {
            console.log(`  Transition: ${scene.structure.transition}`);
          }
        }
      });
      console.log('\n' + '━'.repeat(50));

      // Step 2: Generate keyframes if requested (prevents style drift)
      let keyframes = null;
      if (options.keyframes && numClips > 1) {
        const keyframeGen = new KeyframeGenerator(process.env.REPLICATE_API_TOKEN, options.output);

        // Generate or use style reference
        let styleRef = options.styleRef || null;
        if (options.keyframes && !styleRef) {
          styleRef = await keyframeGen.generateStyleReference(prompt);
        }

        // Generate keyframes for all scenes
        keyframes = await keyframeGen.generateKeyframes(scenes, styleRef);
      }

      // Step 3: Generate videos
      const generator = new VideoGenerator(process.env.REPLICATE_API_TOKEN);

      let results;
      if (numClips === 1) {
        const result = await generator.generateClip(options.model, scenes[0], options.image);
        results = [result];
      } else {
        results = await generator.generateSequence(options.model, scenes, options.image, options.keepFrames, keyframes);
      }

      // Step 3: Add voiceovers if requested
      let finalResults = results;
      if (options.voiceover) {
        // Pass OpenAI client for smart ad copy generation
        const openaiClient = process.env.OPENAI_API_KEY ? new OpenAI({ apiKey: process.env.OPENAI_API_KEY }) : null;
        const audioGen = new AudioGenerator(process.env.REPLICATE_API_TOKEN, options.output, openaiClient);

        console.log('\n🎙️  Voiceover mode enabled (XTTS-v2 Voice Cloning)');
        console.log(`   Language: ${options.language}`);
        if (options.speaker) {
          console.log(`   Speaker: ${options.speaker}`);
        } else {
          console.log(`   Speaker: Using default (./temp/default_speaker.[wav/mp3/m4a/ogg/flv])`);
        }
        if (openaiClient) {
          console.log(`   ✨ Using AI-generated ad copy ("show don't tell")`);
        } else {
          console.log(`   ⚠️  OpenAI not available, using basic narration`);
        }

        finalResults = await audioGen.addVoiceoversToSequence(results, scenes, options.speaker);
      }

      // Step 4: Display URLs
      console.log('\n\n🎥 Generated Videos:');
      console.log('━'.repeat(50));
      finalResults.forEach((result, i) => {
        if (result.error) {
          console.log(`\n[${i + 1}] ❌ Failed: ${result.error}`);
        } else {
          console.log(`\n[${i + 1}] ${result.scene.description || 'Video'}`);

          if (result.finalUrl) {
            console.log(`    Final (with audio): ${result.finalUrl}`);
            console.log(`    Original (silent): ${result.url}`);
            if (result.voiceoverScript) {
              console.log(`    Voiceover: "${result.voiceoverScript.substring(0, 60)}..."`);
            }
          } else {
            console.log(`    URL: ${result.url}`);
          }

          console.log(`    Model: ${result.model}`);
          console.log(`    Time: ${result.generationTime}s`);
        }
      });
      console.log('\n' + '━'.repeat(50));

      // Step 5: Download videos if requested
      let downloads = [];
      if (options.download) {
        const downloader = new VideoDownloader(options.output);
        downloads = await downloader.downloadAll(finalResults);

        // Write summary file
        await downloader.writeSummary(finalResults, downloads);

        // Step 6: Stitch clips together if multiple clips were generated
        if (numClips > 1) {
          try {
            const stitcher = new VideoStitcher(options.output);

            // Collect video paths from downloads array
            const videoPaths = downloads
              .filter(d => d !== null && d.localPath)
              .map(d => d.localPath)
              .filter(p => p && typeof p === 'string' && fs.existsSync(p));

            if (videoPaths.length > 1) {
              let stitchedPath;

              // Choose stitching method based on options
              if (options.smoothTransitions) {
                console.log(`\n🎬 Stitching with frame interpolation...`);
                const fps = parseInt(options.interpolationFps) || 60;
                stitchedPath = await stitcher.stitchVideosWithInterpolation(
                  videoPaths,
                  null,
                  { fps, method: 'blend' }
                );
              } else if (options.crossfade) {
                console.log(`\n🎬 Stitching with crossfade transitions (${options.crossfade}s)...`);
                stitchedPath = await stitcher.stitchVideosWithCrossfade(
                  videoPaths,
                  null,
                  options.crossfade
                );
              } else {
                // Default: fast concat without transitions
                stitchedPath = await stitcher.stitchVideos(videoPaths);
              }

              console.log(`\n🎬 Final stitched video ready!`);
              console.log(`   Watch: ${stitchedPath}`);
            } else {
              console.log(`\n⚠️  Need at least 2 clips to stitch (found ${videoPaths.length})`);
            }
          } catch (error) {
            console.error(`\n⚠️  Could not stitch videos: ${error.message}`);
            console.log(`   Individual clips are still available in ${options.output}`);
          }
        }
      } else {
        console.log('\n⏭️  Skipping downloads (--no-download flag used)');
      }

      console.log('\n✅ Done!\n');

    } catch (error) {
      console.error('\n❌ Error:', error.message);
      if (process.env.DEBUG) {
        console.error(error);
      }
      process.exit(1);
    }
  });

program
  .command('models')
  .description('List available video generation models')
  .action(() => {
    console.log('\n📹 Available Models:\n');
    console.log('━'.repeat(50));

    const models = listModels();
    models.forEach(model => {
      console.log(`\n${model.key}`);
      console.log(`  Name: ${model.name}`);
      console.log(`  Duration: ${model.duration}`);
      console.log(`  Cost: ${model.cost}`);
      console.log(`  Image Support: ${model.supportsImage ? 'Yes' : 'No'}`);
    });

    console.log('\n' + '━'.repeat(50));
    console.log('\nUsage: video-gen generate "your prompt" --model <model-key>\n');
  });

program
  .command('voices')
  .description('Show XTTS-v2 voice cloning information')
  .action(() => {
    console.log('\n🎙️  XTTS-v2 Voice Cloning:\n');
    console.log('━'.repeat(50));

    const info = AudioGenerator.getAvailableVoices();
    Object.entries(info).forEach(([key, description]) => {
      console.log(`\n${key}:`);
      console.log(`  ${description}`);
    });

    console.log('\n\nSupported Languages:');
    console.log('━'.repeat(50));
    const languages = AudioGenerator.getSupportedLanguages();
    languages.forEach(lang => {
      console.log(`  ${lang.code} - ${lang.name}`);
    });

    console.log('\n' + '━'.repeat(50));
    console.log('\nUsage: video-gen generate "your prompt" --voiceover --speaker <path-to-audio-file> --language <lang-code>\n');
    console.log('Example: video-gen generate "Product demo" --voiceover --speaker ./my_voice.wav --language en\n');
  });

program.parse();
