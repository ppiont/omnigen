import fs from 'fs';
import path from 'path';
import https from 'https';
import http from 'http';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

/**
 * Download videos from URLs
 */
export class VideoDownloader {
  constructor(outputDir = './output') {
    this.outputDir = path.resolve(outputDir);
    this.ensureOutputDir();
  }

  ensureOutputDir() {
    if (!fs.existsSync(this.outputDir)) {
      fs.mkdirSync(this.outputDir, { recursive: true });
      console.log(`📁 Created output directory: ${this.outputDir}`);
    }
  }

  /**
   * Download a single video from URL
   * @param {string} url - Video URL
   * @param {string} filename - Output filename
   * @returns {Promise<string>} Path to downloaded file
   */
  async downloadVideo(url, filename) {
    return new Promise((resolve, reject) => {
      const outputPath = path.join(this.outputDir, filename);

      console.log(`⬇️  Downloading: ${filename}...`);

      const protocol = url.startsWith('https') ? https : http;

      const file = fs.createWriteStream(outputPath);

      protocol.get(url, (response) => {
        if (response.statusCode === 200) {
          response.pipe(file);

          file.on('finish', () => {
            file.close();
            console.log(`✅ Downloaded: ${outputPath}`);
            resolve(outputPath);
          });
        } else if (response.statusCode === 302 || response.statusCode === 301) {
          // Handle redirect
          file.close();
          fs.unlinkSync(outputPath);
          this.downloadVideo(response.headers.location, filename)
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
   * Download multiple videos
   * @param {Array} results - Array of generation results
   * @returns {Promise<Array>} Array of downloaded file paths
   */
  async downloadAll(results) {
    console.log(`\n📥 Downloading ${results.length} video(s)...`);

    const downloads = [];

    for (let i = 0; i < results.length; i++) {
      const result = results[i];

      if (result.error) {
        console.log(`⏭️  Skipping scene ${i + 1} (generation failed)`);
        downloads.push(null);
        continue;
      }

      try {
        const timestamp = Date.now();
        // Use voiceover version if available, otherwise silent version
        const url = result.finalUrl || result.url;
        const filename = result.finalPath
          ? path.basename(result.finalPath)
          : `video_${i + 1}_${timestamp}.mp4`;

        const filePath = await this.downloadVideo(url, filename);
        downloads.push({ localPath: filePath, url: url });
      } catch (error) {
        console.error(`❌ Failed to download scene ${i + 1}:`, error.message);
        downloads.push(null);
      }
    }

    const successCount = downloads.filter(d => d !== null).length;
    console.log(`\n✅ Downloaded ${successCount}/${results.length} video(s)`);

    return downloads;
  }

  /**
   * Generate a summary file with URLs and metadata
   */
  async writeSummary(results, downloads) {
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const summaryPath = path.join(this.outputDir, `summary_${timestamp}.json`);

    const summary = {
      generatedAt: new Date().toISOString(),
      totalClips: results.length,
      clips: results.map((result, i) => {
        const download = downloads[i];
        let localFile = null;

        if (download) {
          // Handle both string paths and objects with localPath property
          if (typeof download === 'string') {
            localFile = path.basename(download);
          } else if (download.localPath) {
            localFile = path.basename(download.localPath);
          }
        }

        return {
          index: i + 1,
          url: result.url,
          localFile: localFile,
          scene: result.scene,
          model: result.model,
          generationTime: result.generationTime,
          error: result.error || null,
        };
      })
    };

    fs.writeFileSync(summaryPath, JSON.stringify(summary, null, 2));
    console.log(`📄 Summary written to: ${summaryPath}`);

    return summaryPath;
  }
}
