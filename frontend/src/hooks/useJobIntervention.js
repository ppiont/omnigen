import { useState, useCallback } from 'react';
import { jobIntervention } from '../services/jobInterventionAPI';
import { showToast } from '../utils/toast';

export function useJobIntervention(jobId, onUpdate) {
  const [isUpdating, setIsUpdating] = useState(false);

  const pauseJob = useCallback(async () => {
    setIsUpdating(true);
    try {
      await jobIntervention.pause(jobId);
      showToast('Job paused', 'info');
      onUpdate?.();
    } catch (err) {
      console.error('Failed to pause job:', err);
      showToast('Failed to pause job', 'error');
    } finally {
      setIsUpdating(false);
    }
  }, [jobId, onUpdate]);

  const resumeJob = useCallback(async () => {
    setIsUpdating(true);
    try {
      await jobIntervention.resume(jobId);
      showToast('Job resumed', 'success');
      onUpdate?.();
    } catch (err) {
      console.error('Failed to resume job:', err);
      showToast('Failed to resume job', 'error');
    } finally {
      setIsUpdating(false);
    }
  }, [jobId, onUpdate]);

  const modifyScene = useCallback(async (sceneNumber, data) => {
    setIsUpdating(true);
    try {
      await jobIntervention.modifyScene(jobId, sceneNumber, data);
      showToast(`Scene ${sceneNumber} updated`, 'success');
      onUpdate?.();
    } catch (err) {
      console.error('Failed to update scene:', err);
      showToast('Failed to update scene', 'error');
    } finally {
      setIsUpdating(false);
    }
  }, [jobId, onUpdate]);

  const skipScene = useCallback(async (sceneNumber) => {
    if (!window.confirm(`Are you sure you want to skip scene ${sceneNumber}?`)) return;
    
    setIsUpdating(true);
    try {
      await jobIntervention.skipScene(jobId, sceneNumber);
      showToast(`Scene ${sceneNumber} skipped`, 'info');
      onUpdate?.();
    } catch (err) {
      console.error('Failed to skip scene:', err);
      showToast('Failed to skip scene', 'error');
    } finally {
      setIsUpdating(false);
    }
  }, [jobId, onUpdate]);

  const updateParameters = useCallback(async (params) => {
    setIsUpdating(true);
    try {
      await jobIntervention.updateParameters(jobId, params);
      showToast('Parameters updated for remaining scenes', 'success');
      onUpdate?.();
    } catch (err) {
      console.error('Failed to update parameters:', err);
      showToast('Failed to update parameters', 'error');
    } finally {
      setIsUpdating(false);
    }
  }, [jobId, onUpdate]);

  return {
    isUpdating,
    pauseJob,
    resumeJob,
    modifyScene,
    skipScene,
    updateParameters
  };
}
