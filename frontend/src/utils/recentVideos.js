/**
 * Utility functions for tracking recently opened videos
 */

const STORAGE_KEY = "recently-opened-videos";
const MAX_RECENT_VIDEOS = 20;

/**
 * Get recently opened videos from localStorage
 * @returns {Array<{jobId: string, openedAt: number}>} Array of recently opened videos
 */
export function getRecentlyOpenedVideos() {
  if (typeof window === "undefined") return [];
  
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      return JSON.parse(stored);
    }
  } catch (error) {
    console.error("[RECENT_VIDEOS] Failed to load recently opened videos:", error);
  }
  
  return [];
}

/**
 * Add a video to recently opened list
 * @param {string} jobId - The job ID of the video
 */
export function addRecentlyOpenedVideo(jobId) {
  if (typeof window === "undefined" || !jobId) return;
  
  try {
    const recent = getRecentlyOpenedVideos();
    
    // Remove if already exists (to avoid duplicates)
    const filtered = recent.filter((item) => item.jobId !== jobId);
    
    // Add to beginning with current timestamp
    const updated = [
      { jobId, openedAt: Date.now() },
      ...filtered
    ].slice(0, MAX_RECENT_VIDEOS); // Keep only the most recent N videos
    
    localStorage.setItem(STORAGE_KEY, JSON.stringify(updated));
  } catch (error) {
    console.error("[RECENT_VIDEOS] Failed to save recently opened video:", error);
  }
}

/**
 * Remove a video from recently opened list
 * @param {string} jobId - The job ID to remove
 */
export function removeRecentlyOpenedVideo(jobId) {
  if (typeof window === "undefined" || !jobId) return;
  
  try {
    const recent = getRecentlyOpenedVideos();
    const filtered = recent.filter((item) => item.jobId !== jobId);
    localStorage.setItem(STORAGE_KEY, JSON.stringify(filtered));
  } catch (error) {
    console.error("[RECENT_VIDEOS] Failed to remove recently opened video:", error);
  }
}

/**
 * Clear all recently opened videos
 */
export function clearRecentlyOpenedVideos() {
  if (typeof window === "undefined") return;
  
  try {
    localStorage.removeItem(STORAGE_KEY);
  } catch (error) {
    console.error("[RECENT_VIDEOS] Failed to clear recently opened videos:", error);
  }
}

