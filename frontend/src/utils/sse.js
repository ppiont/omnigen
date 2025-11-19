/**
 * Connect to SSE progress endpoint
 * @param {string} jobId - The job ID to track
 * @param {Object} callbacks - Event callbacks
 * @param {Function} callbacks.onUpdate - Called with ProgressResponse on update
 * @param {Function} callbacks.onDone - Called with status on completion
 * @param {Function} callbacks.onError - Called with error message
 * @returns {Function} cleanup function to close connection
 */
export function connectToProgress(jobId, { onUpdate, onDone, onError }) {
  const eventSource = new EventSource(
    `http://localhost:8080/api/v1/jobs/${jobId}/progress`,
    { withCredentials: true }
  );

  eventSource.addEventListener('update', (event) => {
    const data = JSON.parse(event.data);
    onUpdate(data);
  });

  eventSource.addEventListener('done', (event) => {
    const data = JSON.parse(event.data);
    onDone(data.status);
    eventSource.close();
  });

  eventSource.addEventListener('error', (event) => {
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
    onError('Failed to connect to progress stream');
    eventSource.close();
  };

  // Return cleanup function
  return () => {
    eventSource.close();
  };
}
