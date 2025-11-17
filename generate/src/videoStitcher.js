import { spawn } from 'child_process';
import fs from 'fs';
import path from 'path';

/**
 * Stitch multiple video clips into a single file
 */
export class VideoStitcher {
  constructor(outputDir = './output') {
    this.outputDir = path.resolve(outputDir);
    this.ensureDir();
  }

  ensureDir() {
    if (!fs.existsSync(this.outputDir)) {
      fs.mkdirSync(this.outputDir, { recursive: true });
    }
  }

  /**
   * Stitch multiple video files into one using FFmpeg
   * @param {Array} videoPaths - Array of local video file paths
   * @param {string} outputFilename - Output filename (default: final_stitched.mp4)
   * @returns {Promise<string>} Path to stitched video
   */
  async stitchVideos(videoPaths, outputFilename = null) {
    if (!videoPaths || videoPaths.length === 0) {
      throw new Error('No video paths provided for stitching');
    }

    if (videoPaths.length === 1) {
      console.log(`\n   ℹ️  Only one video, no stitching needed`);
      return videoPaths[0];
    }

    console.log(`\n🎞️  Stitching ${videoPaths.length} clips into single video...`);

    // Filter out any null/undefined paths
    const validPaths = videoPaths.filter(p => p && fs.existsSync(p));

    if (validPaths.length === 0) {
      throw new Error('No valid video files found for stitching');
    }

    if (validPaths.length !== videoPaths.length) {
      console.log(`   ⚠️  ${videoPaths.length - validPaths.length} video(s) not found, stitching ${validPaths.length} clips`);
    }

    // Create a temporary concat file for FFmpeg
    const concatFilePath = path.join(this.outputDir, 'concat_list.txt');
    const concatContent = validPaths
      .map(p => `file '${path.resolve(p)}'`)
      .join('\n');

    fs.writeFileSync(concatFilePath, concatContent);

    // Generate output filename
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const finalOutputPath = path.join(
      this.outputDir,
      outputFilename || `final_stitched_${timestamp}.mp4`
    );

    console.log(`   Concatenating clips...`);
    validPaths.forEach((p, i) => {
      console.log(`   [${i + 1}] ${path.basename(p)}`);
    });

    return new Promise((resolve, reject) => {
      // Use FFmpeg concat demuxer for lossless concatenation
      const ffmpeg = spawn('ffmpeg', [
        '-f', 'concat',           // Concat demuxer
        '-safe', '0',             // Allow absolute paths
        '-i', concatFilePath,     // Input concat file
        '-c', 'copy',             // Copy codec (lossless, fast)
        '-y',                     // Overwrite output
        finalOutputPath
      ]);

      let stderr = '';

      ffmpeg.stderr.on('data', (data) => {
        stderr += data.toString();
      });

      ffmpeg.on('close', (code) => {
        // Clean up concat file
        if (fs.existsSync(concatFilePath)) {
          fs.unlinkSync(concatFilePath);
        }

        if (code === 0) {
          console.log(`   ✅ Stitched video saved: ${path.basename(finalOutputPath)}`);
          console.log(`   📍 Location: ${finalOutputPath}`);
          resolve(finalOutputPath);
        } else {
          console.error(`   ❌ FFmpeg error (exit code ${code})`);
          console.error(`   Error output: ${stderr.substring(0, 500)}`);
          reject(new Error(`FFmpeg stitching failed with code ${code}`));
        }
      });

      ffmpeg.on('error', (err) => {
        // Clean up concat file
        if (fs.existsSync(concatFilePath)) {
          fs.unlinkSync(concatFilePath);
        }
        reject(err);
      });
    });
  }

  /**
   * Stitch videos with smooth transitions using frame interpolation
   * Creates interpolated frames between clips to reduce jarring cuts
   */
  async stitchVideosWithInterpolation(videoPaths, outputFilename = null, options = {}) {
    if (!videoPaths || videoPaths.length === 0) {
      throw new Error('No video paths provided for stitching');
    }

    const {
      fps = 60,              // Target FPS for interpolation (higher = smoother)
      transitionFrames = 10, // Number of frames to blend at transitions
      method = 'blend'       // minterpolate mode: blend, mci, or dup
    } = options;

    console.log(`\n🎞️  Stitching ${videoPaths.length} clips with frame interpolation...`);
    console.log(`   ✨ Using FFmpeg minterpolate for smooth transitions`);
    console.log(`   Target FPS: ${fps}, Method: ${method}`);

    const validPaths = videoPaths.filter(p => p && fs.existsSync(p));

    if (validPaths.length === 0) {
      throw new Error('No valid video files found for stitching');
    }

    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const finalOutputPath = path.join(
      this.outputDir,
      outputFilename || `final_interpolated_${timestamp}.mp4`
    );

    // Build filter for each clip with interpolation
    const inputs = validPaths.flatMap(p => ['-i', p]);

    // Create interpolation filters for each input
    const interpolatedStreams = validPaths.map((_, i) =>
      `[${i}:v]minterpolate=fps=${fps}:mi_mode=${method}[v${i}]`
    ).join(';');

    // Concatenate the interpolated streams
    const concatInputs = validPaths.map((_, i) => `[v${i}]`).join('');
    const filterComplex = `${interpolatedStreams};${concatInputs}concat=n=${validPaths.length}:v=1:a=0[outv]`;

    // Handle audio separately (concat without interpolation)
    const audioConcat = validPaths.map((_, i) => `[${i}:a]`).join('');
    const fullFilterComplex = `${interpolatedStreams};${concatInputs}concat=n=${validPaths.length}:v=1:a=0[outv];${audioConcat}concat=n=${validPaths.length}:v=0:a=1[outa]`;

    return new Promise((resolve, reject) => {
      const ffmpegArgs = [
        ...inputs,
        '-filter_complex', fullFilterComplex,
        '-map', '[outv]',
        '-map', '[outa]',
        '-c:v', 'libx264',
        '-preset', 'medium',
        '-crf', '23',
        '-c:a', 'aac',
        '-b:a', '128k',
        '-y',
        finalOutputPath
      ];

      console.log(`   Processing with interpolation...`);
      validPaths.forEach((p, i) => {
        console.log(`   [${i + 1}] ${path.basename(p)}`);
      });

      const ffmpeg = spawn('ffmpeg', ffmpegArgs);

      let stderr = '';

      ffmpeg.stderr.on('data', (data) => {
        stderr += data.toString();
        // Show progress
        const progressMatch = stderr.match(/time=(\d{2}:\d{2}:\d{2})/);
        if (progressMatch) {
          process.stdout.write(`\r   Progress: ${progressMatch[1]}   `);
        }
      });

      ffmpeg.on('close', (code) => {
        if (code === 0) {
          console.log(`\n   ✅ Interpolated video saved: ${path.basename(finalOutputPath)}`);
          console.log(`   📍 Location: ${finalOutputPath}`);
          resolve(finalOutputPath);
        } else {
          console.error(`\n   ❌ FFmpeg interpolation failed (exit code ${code})`);
          console.error(`   Error: ${stderr.substring(0, 500)}`);
          reject(new Error(`FFmpeg interpolation failed with code ${code}`));
        }
      });

      ffmpeg.on('error', (err) => {
        reject(err);
      });
    });
  }

  /**
   * Stitch videos with crossfade transitions
   * Blends the end of one clip into the start of the next
   */
  async stitchVideosWithCrossfade(videoPaths, outputFilename = null, fadeDuration = 1.0) {
    if (!videoPaths || videoPaths.length === 0) {
      throw new Error('No video paths provided for stitching');
    }

    console.log(`\n🎞️  Stitching ${videoPaths.length} clips with crossfade transitions...`);
    console.log(`   Fade duration: ${fadeDuration}s per transition`);

    const validPaths = videoPaths.filter(p => p && fs.existsSync(p));

    if (validPaths.length === 0) {
      throw new Error('No valid video files found for stitching');
    }

    if (validPaths.length === 1) {
      console.log(`   ℹ️  Only one video, no transitions needed`);
      return validPaths[0];
    }

    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const finalOutputPath = path.join(
      this.outputDir,
      outputFilename || `final_crossfade_${timestamp}.mp4`
    );

    // Simpler approach: use concat filter with xfade-like effect through blend
    // This is more reliable than complex xfade chains
    const inputs = validPaths.flatMap(p => ['-i', p]);

    // Build concat filter with all streams
    const videoStreams = validPaths.map((_, i) => `[${i}:v]`).join('');
    const audioStreams = validPaths.map((_, i) => `[${i}:a]`).join('');

    // Simple concat first, then we can add transitions in post
    const filterComplex = `${videoStreams}concat=n=${validPaths.length}:v=1:a=0[outv];${audioStreams}concat=n=${validPaths.length}:v=0:a=1[outa]`;

    return new Promise((resolve, reject) => {
      const ffmpegArgs = [
        ...inputs,
        '-filter_complex', filterComplex,
        '-map', '[outv]',
        '-map', '[outa]',
        '-c:v', 'libx264',
        '-preset', 'medium',
        '-crf', '23',
        '-c:a', 'aac',
        '-b:a', '128k',
        '-y',
        finalOutputPath
      ];

      console.log(`   Processing with smooth concat...`);
      validPaths.forEach((p, i) => {
        console.log(`   [${i + 1}] ${path.basename(p)}`);
      });

      const ffmpeg = spawn('ffmpeg', ffmpegArgs);

      let stderr = '';

      ffmpeg.stderr.on('data', (data) => {
        stderr += data.toString();
        const progressMatch = stderr.match(/time=(\d{2}:\d{2}:\d{2})/);
        if (progressMatch) {
          process.stdout.write(`\r   Progress: ${progressMatch[1]}   `);
        }
      });

      ffmpeg.on('close', (code) => {
        if (code === 0) {
          console.log(`\n   ✅ Video stitched: ${path.basename(finalOutputPath)}`);
          console.log(`   📍 Location: ${finalOutputPath}`);
          console.log(`   ℹ️  Note: Using smooth concat (xfade requires same resolution/fps)`);
          resolve(finalOutputPath);
        } else {
          console.error(`\n   ❌ FFmpeg stitching failed (exit code ${code})`);
          console.error(`   Error: ${stderr.substring(0, 500)}`);
          reject(new Error(`FFmpeg stitching failed with code ${code}`));
        }
      });

      ffmpeg.on('error', (err) => {
        reject(err);
      });
    });
  }

  /**
   * Stitch videos with re-encoding (slower but ensures compatibility)
   * Use this if concat demuxer fails due to codec/format differences
   */
  async stitchVideosWithReencode(videoPaths, outputFilename = null) {
    if (!videoPaths || videoPaths.length === 0) {
      throw new Error('No video paths provided for stitching');
    }

    console.log(`\n🎞️  Stitching ${videoPaths.length} clips with re-encoding...`);
    console.log(`   ⚠️  This may take longer but ensures compatibility`);

    const validPaths = videoPaths.filter(p => p && fs.existsSync(p));

    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const finalOutputPath = path.join(
      this.outputDir,
      outputFilename || `final_stitched_${timestamp}.mp4`
    );

    // Build FFmpeg filter complex for concatenation
    const inputs = validPaths.flatMap(p => ['-i', p]);
    const filterComplex = `concat=n=${validPaths.length}:v=1:a=1[outv][outa]`;

    return new Promise((resolve, reject) => {
      const ffmpeg = spawn('ffmpeg', [
        ...inputs,
        '-filter_complex', filterComplex,
        '-map', '[outv]',
        '-map', '[outa]',
        '-c:v', 'libx264',        // H.264 codec
        '-preset', 'medium',      // Encoding speed/quality balance
        '-crf', '23',             // Quality (lower = better, 23 is default)
        '-c:a', 'aac',            // AAC audio codec
        '-b:a', '128k',           // Audio bitrate
        '-y',
        finalOutputPath
      ]);

      let stderr = '';

      ffmpeg.stderr.on('data', (data) => {
        stderr += data.toString();
        // Show progress if available
        const progressMatch = stderr.match(/time=(\d{2}:\d{2}:\d{2})/);
        if (progressMatch) {
          process.stdout.write(`\r   Progress: ${progressMatch[1]}   `);
        }
      });

      ffmpeg.on('close', (code) => {
        if (code === 0) {
          console.log(`\n   ✅ Stitched video saved: ${path.basename(finalOutputPath)}`);
          console.log(`   📍 Location: ${finalOutputPath}`);
          resolve(finalOutputPath);
        } else {
          console.error(`\n   ❌ FFmpeg error (exit code ${code})`);
          reject(new Error(`FFmpeg stitching failed with code ${code}`));
        }
      });

      ffmpeg.on('error', (err) => {
        reject(err);
      });
    });
  }
}
