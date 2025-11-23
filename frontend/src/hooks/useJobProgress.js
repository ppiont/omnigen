import { useState, useEffect, useCallback, useRef, useMemo } from 'react';
import { connectToProgress } from '../utils/sse';

/**
 * Custom hook for tracking job progress via SSE
 * @param {string} jobId - The job ID to track
 * @param {Object} options - Configuration options
 * @param {Function} options.onComplete - Called when job completes successfully
 * @param {Function} options.onFailed - Called when job fails
 * @param {boolean} options.autoConnect - Auto-connect to SSE (default: true)
 * @param {number} options.maxRetries - Maximum retry attempts (default: 3)
 * @returns {Object} Progress state and helpers
 */
const DEFAULT_STAGE_ORDER = [
  { key: 'script', label: 'Script Generation' },
  { key: 'audio', label: 'Audio Mix' },
  { key: 'scenes', label: 'Scene Rendering' },
  { key: 'compose', label: 'Final Composition' },
];

export function useJobProgress(jobId, options = {}) {
  const {
    onComplete,
    onFailed,
    autoConnect = true,
    maxRetries = 3
  } = options;

  const [progress, setProgress] = useState(null);
  const [status, setStatus] = useState('idle'); // idle | connecting | connected | completed | failed | error
  const [error, setError] = useState(null);

  const retriesRef = useRef(0);
  const cleanupRef = useRef(null);
  const isRetryingRef = useRef(false);

  // Handle SSE progress updates
  const handleProgressUpdate = useCallback((data) => {
    setProgress(data);
    setError(null);
    setStatus('connected');
    retriesRef.current = 0; // Reset retries on successful update
  }, []);

  // Handle job completion
  const handleJobDone = useCallback((finalStatus, finalData) => {
    setStatus(finalStatus); // 'completed' or 'failed'

    // Use finalData if provided (full ProgressResponse), otherwise fall back to current progress
    const finalProgress = finalData && finalData.job_id ? finalData : progress;

    if (finalStatus === 'completed') {
      onComplete?.(finalProgress);
    } else {
      // Extract error message from job data if available
      const errorMessage = finalProgress?.error_message || 
                          finalProgress?.message || 
                          'Video generation failed. Please try again.';
      setError(errorMessage);
      onFailed?.(finalProgress);
    }
  }, [progress, onComplete, onFailed]);

  // Handle SSE errors
  const handleSSEError = useCallback((errorMessage) => {
    console.error('SSE Error:', errorMessage);
    setError(errorMessage);

    // Retry logic with exponential backoff
    if (retriesRef.current < maxRetries && !isRetryingRef.current) {
      retriesRef.current++;
      isRetryingRef.current = true;

      const backoffDelay = 2000 * retriesRef.current; // 2s, 4s, 6s
      console.warn(`Connection lost. Retry ${retriesRef.current}/${maxRetries} in ${backoffDelay}ms...`);

      setTimeout(() => {
        isRetryingRef.current = false;
        setStatus('connecting');
        setError(null);
      }, backoffDelay);
    } else {
      setStatus('error');
    }
  }, [maxRetries]);

  // Connect to SSE
  const connect = useCallback(() => {
    if (!jobId) return;

    setStatus('connecting');
    setError(null);

    cleanupRef.current = connectToProgress(jobId, {
      onUpdate: handleProgressUpdate,
      onDone: handleJobDone,
      onError: handleSSEError,
    });
  }, [jobId, handleProgressUpdate, handleJobDone, handleSSEError]);

  // Auto-connect when jobId changes (with small delay to avoid race condition)
  useEffect(() => {
    if (!jobId || !autoConnect) return;

    // Small delay to ensure job is written to DynamoDB before first SSE poll
    const delayTimer = setTimeout(() => {
      connect();
    }, 100);

    return () => {
      clearTimeout(delayTimer);
      cleanupRef.current?.();
    };
  }, [jobId, autoConnect, connect]);

  // Reconnect manually (for retry buttons)
  const reconnect = useCallback(() => {
    retriesRef.current = 0;
    isRetryingRef.current = false;
    cleanupRef.current?.();
    setStatus('idle');
    setError(null);
    connect();
  }, [connect]);

  const stageTimeline = useMemo(() => {
    const completedSet = new Set(progress?.stages_completed || progress?.completed_stages || []);
    const pendingSet = new Set(progress?.stages_pending || progress?.pending_stages || []);
    const currentStageKey = progress?.current_stage || progress?.currentStage || '';

    return DEFAULT_STAGE_ORDER.map((stage) => {
      let status = 'pending';
      if (completedSet.has(stage.key)) {
        status = 'completed';
      } else if (pendingSet.has(stage.key)) {
        status = 'pending';
      }

      if (stage.key === currentStageKey && status !== 'completed') {
        status = 'active';
      }

      return {
        ...stage,
        status,
      };
    });
  }, [progress]);

  return {
    // Raw progress data
    progress,
    status,  // Hook connection status (idle | connecting | connected | completed | failed | error)
    error,

    // Actions
    reconnect,

    // Computed helpers
    isConnecting: status === 'connecting',
    isConnected: status === 'connected',
    isCompleted: status === 'completed',
    isFailed: status === 'failed',
    hasError: status === 'error',

    // Granular progress data (with fallbacks)
    jobStatus: progress?.status ?? 'pending',  // Backend job status (pending | processing | completed | failed)
    percentage: progress?.progress ?? 0,
    currentStage: progress?.current_stage_display ?? 'Initializing...',
    currentStageRaw: progress?.current_stage ?? '',
    completedStages: progress?.stages_completed ?? [],
    pendingStages: progress?.stages_pending ?? [],
    estimatedTimeRemaining: progress?.estimated_time_remaining,
    assets: progress?.assets ?? {},
    stageTimeline,

    // Retry info
    retryCount: retriesRef.current,
    canRetry: retriesRef.current < maxRetries
  };
}
