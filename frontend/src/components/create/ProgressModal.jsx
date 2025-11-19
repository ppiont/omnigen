import { useState, useEffect, useCallback } from 'react';
import { connectToProgress } from '../../utils/sse';
import { formatTimeRemaining } from '../../utils/format';
import '../../styles/progress-modal.css';

/**
 * ProgressModal - Real-time progress tracking modal for video generation
 */
function ProgressModal({ jobId, onComplete, onCancel, isOpen }) {
  const [progress, setProgress] = useState(null);
  const [error, setError] = useState(null);
  const [connectionStatus, setConnectionStatus] = useState('connecting'); // 'connecting', 'connected', 'error'
  const [retryCount, setRetryCount] = useState(0);
  const [isRetrying, setIsRetrying] = useState(false);

  // Retry connection
  const handleRetry = useCallback(() => {
    if (isRetrying || retryCount >= 3) return;

    setIsRetrying(true);
    setRetryCount(prev => prev + 1);
    setError(null);
    setConnectionStatus('connecting');

    // Reset states for retry
    setProgress(null);

    console.log(`Retrying SSE connection (attempt ${retryCount + 1})`);
  }, [isRetrying, retryCount]);

  // Handle SSE progress updates
  const handleProgressUpdate = useCallback((data) => {
    setProgress(data);
    setError(null);
    setConnectionStatus('connected');
  }, []);

  // Handle job completion
  const handleJobDone = useCallback((status) => {
    if (status === 'completed') {
      // Navigate to workspace after a brief delay to show completion
      setTimeout(() => {
        onComplete?.(progress);
      }, 2000);
    } else if (status === 'failed') {
      setError('Video generation failed. Please try again.');
    }
  }, [progress, onComplete]);

  // Handle SSE errors
  const handleSSEError = useCallback((errorMessage) => {
    console.error('SSE Error:', errorMessage);
    setError(errorMessage);
    setConnectionStatus('error');
  }, []);

  // Connect to SSE when modal opens and jobId is available
  useEffect(() => {
    if (!isOpen || !jobId) return;

    // Only reset states if not retrying
    if (!isRetrying) {
      setError(null);
      setConnectionStatus('connecting');
      setProgress(null);
      setRetryCount(0);
    }

    const cleanup = connectToProgress(jobId, {
      onUpdate: handleProgressUpdate,
      onDone: handleJobDone,
      onError: handleSSEError,
    });

    return cleanup; // Cleanup function to close SSE connection
  }, [isOpen, jobId, handleProgressUpdate, handleJobDone, handleSSEError, isRetrying]);

  // Handle retry state reset
  useEffect(() => {
    if (isRetrying && connectionStatus === 'connecting') {
      setIsRetrying(false);
    }
  }, [isRetrying, connectionStatus]);

  // Don't render if modal is not open
  if (!isOpen) return null;

  const renderStageList = () => {
    if (!progress) return null;

    const allStages = [
      ...progress.stages_completed.map(stage => ({ ...stage, status: 'completed' })),
      ...(progress.current_stage ? [{
        name: progress.current_stage,
        display_name: progress.current_stage_display,
        status: 'current'
      }] : []),
      ...progress.stages_pending.map(stage => ({ ...stage, status: 'pending' })),
    ];

    return (
      <div className="progress-stages">
        {allStages.map((stage, index) => (
          <div key={stage.name} className={`stage-item stage-${stage.status}`}>
            {stage.status === 'completed' && <span className="stage-check">✓</span>}
            {stage.status === 'current' && <span className="stage-spinner">⟳</span>}
            {stage.status === 'pending' && <span className="stage-pending">⋯</span>}
            <span className="stage-text">{stage.display_name}</span>
          </div>
        ))}
      </div>
    );
  };

  const renderProgressContent = () => {
    // Connection error state
    if (connectionStatus === 'error' && !progress) {
      const canRetry = retryCount < 3;
      return (
        <div className="progress-error">
          <div className="error-icon">⚠️</div>
          <h3 className="error-title">Connection Failed</h3>
          <p className="error-message">
            {error || 'Unable to connect to progress stream. Please check your connection and try again.'}
            {retryCount > 0 && ` (Attempt ${retryCount + 1}/3)`}
          </p>
          <div className="error-actions">
            {canRetry && (
              <button
                className="btn-retry"
                onClick={handleRetry}
                disabled={isRetrying}
              >
                {isRetrying ? 'Retrying...' : 'Retry Connection'}
              </button>
            )}
            <button className="btn-cancel" onClick={onCancel}>
              Cancel
            </button>
            {!canRetry && (
              <button className="btn-retry" onClick={() => window.location.reload()}>
                Refresh Page
              </button>
            )}
          </div>
        </div>
      );
    }

    // Job error state
    if (error && progress?.status === 'failed') {
      return (
        <div className="progress-error">
          <div className="error-icon">❌</div>
          <h3 className="error-title">Generation Failed</h3>
          <p className="error-message">{error}</p>
          <div className="error-actions">
            <button className="btn-retry" onClick={onCancel}>
              Try Again
            </button>
            <button className="btn-cancel" onClick={onCancel}>
              Cancel
            </button>
          </div>
        </div>
      );
    }

    // Success state
    if (progress?.status === 'completed') {
      return (
        <div className="progress-success">
          <div className="success-icon">✅</div>
          <h3 className="success-title">Video Ready!</h3>
          <p className="success-message">Your video has been generated successfully.</p>
          <p className="success-detail">Redirecting to workspace...</p>
        </div>
      );
    }

    // Loading/Progress state
    return (
      <>
        <div className="progress-header">
          <h3 className="progress-title">Generating Your Video</h3>
          <div className="progress-bar-container">
            <div
              className="progress-bar"
              style={{ width: `${progress?.progress || 0}%` }}
            />
            <span className="progress-text">{progress?.progress || 0}%</span>
          </div>
        </div>

        <div className="progress-details">
          <div className="current-stage">
            {progress?.current_stage_display || 'Initializing...'}
          </div>
          {progress?.estimated_time_remaining && (
            <div className="time-remaining">
              Est. {formatTimeRemaining(progress.estimated_time_remaining)} remaining
            </div>
          )}
        </div>

        {renderStageList()}

        <div className="progress-actions">
          <button className="btn-cancel" onClick={onCancel}>
            Cancel
          </button>
        </div>
      </>
    );
  };

  return (
    <div className="progress-modal-overlay">
      <div className="progress-modal">
        {renderProgressContent()}
      </div>
    </div>
  );
}

export default ProgressModal;
