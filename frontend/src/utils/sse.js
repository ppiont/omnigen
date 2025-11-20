/**
 * Connect to SSE progress endpoint
 * @param {string} jobId - The job ID to track
 * @param {Object} callbacks - Event callbacks
 * @param {Function} callbacks.onUpdate - Called with ProgressResponse on update
 * @param {Function} callbacks.onDone - Called with status on completion
 * @param {Function} callbacks.onError - Called with error message
 * @param {number} callbacks.timeout - Optional timeout in milliseconds (default: 30 min)
 * @returns {Function} cleanup function to close connection
 */
export function connectToProgress(jobId, { onUpdate, onDone, onError, timeout = 30 * 60 * 1000 }) {
  const API_BASE_URL = import.meta.env.VITE_API_URL || '';
  const eventSource = new EventSource(
    `${API_BASE_URL}/api/v1/jobs/${jobId}/progress`,
    { withCredentials: true }
  );

  // Set timeout to prevent zombie connections
  const timeoutId = setTimeout(() => {
    console.warn(`[SSE] Connection timeout after ${timeout}ms for job ${jobId}`);
    eventSource.close();
    onError('Connection timeout - video generation is taking longer than expected');
  }, timeout);

  eventSource.addEventListener('update', (event) => {
    const data = JSON.parse(event.data);
    onUpdate(data);
  });

  eventSource.addEventListener('done', (event) => {
    const data = JSON.parse(event.data);
    clearTimeout(timeoutId); // Clear timeout on completion
    // Backend now sends full ProgressResponse, extract status from it
    const status = data.status || (typeof data === 'string' ? data : 'completed');
    onDone(status, data);  // Pass full data as second argument
    eventSource.close();
  });

  eventSource.addEventListener('error', (event) => {
    clearTimeout(timeoutId); // Clear timeout on error
    if (event.data) {
      const data = JSON.parse(event.data);
      onError(data.error);
    } else {
      // Connection error
      onError('Connection lost');
      eventSource.close();
    }
  });

  eventSource.onerror = (error) => {
    console.error('SSE connection error:', error);
    clearTimeout(timeoutId); // Clear timeout on error
    onError('Failed to connect to progress stream');
    eventSource.close();
  };

  // Return cleanup function
  return () => {
    clearTimeout(timeoutId);
    eventSource.close();
  };
}
