/**
 * Format seconds into human-readable time
 * @param {number} seconds
 * @returns {string} e.g., "2 min 30 sec", "45 sec", "1 min"
 */
export function formatTimeRemaining(seconds) {
  if (seconds < 60) {
    return `${seconds} sec`;
  }
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;
  if (remainingSeconds === 0) {
    return `${minutes} min`;
  }
  return `${minutes} min ${remainingSeconds} sec`;
}
