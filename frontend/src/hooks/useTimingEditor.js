import { useState, useCallback, useMemo, useEffect } from 'react';
import { timingAPI } from '../services/timingAPI';
import { showToast } from '../utils/toast';

export function useTimingEditor(jobId, initialScenes = []) {
  const [originalScenes, setOriginalScenes] = useState(initialScenes);
  const [editedScenes, setEditedScenes] = useState(initialScenes);
  const [isApplying, setIsApplying] = useState(false);

  useEffect(() => {
    setOriginalScenes(initialScenes);
    setEditedScenes(initialScenes);
  }, [initialScenes]);

  const totalDuration = useMemo(() => {
    return editedScenes.reduce((total, scene) => total + (scene.duration || 0), 0);
  }, [editedScenes]);

  const isDirty = useMemo(() => {
    if (originalScenes.length !== editedScenes.length) return true;
    return editedScenes.some((scene, index) => {
      const original = originalScenes[index];
      return (
        scene.duration !== original.duration ||
        scene.transition_in !== original.transition_in ||
        scene.transition_out !== original.transition_out
      );
    });
  }, [originalScenes, editedScenes]);

  const changedScenes = useMemo(() => {
    return editedScenes
      .map((scene, index) => {
        const original = originalScenes[index];
        if (!original) return null;
        if (
          scene.duration !== original.duration ||
          scene.transition_in !== original.transition_in ||
          scene.transition_out !== original.transition_out
        ) {
          return { 
            ...scene, 
            needsRegeneration: scene.duration !== original.duration 
          };
        }
        return null;
      })
      .filter(Boolean);
  }, [originalScenes, editedScenes]);

  const updateScene = useCallback((sceneNumber, updates) => {
    setEditedScenes((prev) =>
      prev.map((scene) => {
        const currentSceneNumber = scene.scene_number || scene.SceneNumber;
        if (currentSceneNumber === sceneNumber) {
          return { ...scene, ...updates };
        }
        return scene;
      })
    );
  }, []);

  const resetChanges = useCallback(() => {
    setEditedScenes(originalScenes);
  }, [originalScenes]);

  const saveChanges = useCallback(async (regenerateDurationChanges = false) => {
    setIsApplying(true);
    try {
      // 1. Update scene metadata
      await Promise.all(
        changedScenes.map((scene) =>
          timingAPI.updateSceneTiming(jobId, scene.scene_number || scene.SceneNumber, {
            duration: scene.duration,
            transition_in: scene.transition_in,
            transition_out: scene.transition_out,
          })
        )
      );

      // 2. Regenerate scenes if requested and needed
      if (regenerateDurationChanges) {
        const scenesToRegen = changedScenes.filter((s) => s.needsRegeneration);
        if (scenesToRegen.length > 0) {
          showToast(`Regenerating ${scenesToRegen.length} scenes...`, 'info');
          await Promise.all(
            scenesToRegen.map((scene) =>
              timingAPI.regenerateScene(jobId, scene.scene_number || scene.SceneNumber)
            )
          );
        }
      }

      showToast('Timing and transitions updated', 'success');
      setOriginalScenes(editedScenes); // Commit changes locally
    } catch (error) {
      console.error('Failed to save timing changes:', error);
      showToast('Failed to save changes', 'error');
    } finally {
      setIsApplying(false);
    }
  }, [jobId, changedScenes, editedScenes]);

  return {
    editedScenes,
    totalDuration,
    isDirty,
    isApplying,
    changedScenes,
    updateScene,
    resetChanges,
    saveChanges,
  };
}
