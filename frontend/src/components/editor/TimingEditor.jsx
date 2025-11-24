import PropTypes from 'prop-types';
import { Save, RotateCcw, AlertTriangle } from 'lucide-react';
import { useTimingEditor } from '../../hooks/useTimingEditor';
import TimelineVisualization from './TimelineVisualization';
import SceneTimingCard from './SceneTimingCard';

function TimingEditor({ jobId, initialScenes, onSaveComplete }) {
  const {
    editedScenes,
    totalDuration,
    isDirty,
    isApplying,
    changedScenes,
    updateScene,
    resetChanges,
    saveChanges,
  } = useTimingEditor(jobId, initialScenes);

  const scenesToRegenerate = changedScenes.filter(s => s.needsRegeneration).length;

  const handleSave = async () => {
    if (scenesToRegenerate > 0) {
      if (!window.confirm(`This will regenerate ${scenesToRegenerate} scenes. Continue?`)) return;
    }
    await saveChanges(true);
    onSaveComplete?.();
  };

  return (
    <div className="timing-editor">
      <div className="timing-controls">
        <div className="timing-stats">
          <span className="total-duration">Total Duration: <strong>{totalDuration}s</strong></span>
          {scenesToRegenerate > 0 && (
            <span className="regen-warning">
              <AlertTriangle size={14} /> {scenesToRegenerate} scenes need regen
            </span>
          )}
        </div>
        <div className="timing-actions">
          <button 
            className="btn-ghost" 
            onClick={resetChanges} 
            disabled={!isDirty || isApplying}
          >
            <RotateCcw size={16} /> Reset
          </button>
          <button 
            className="btn-primary" 
            onClick={handleSave} 
            disabled={!isDirty || isApplying}
          >
            <Save size={16} /> {isApplying ? 'Saving...' : 'Apply Changes'}
          </button>
        </div>
      </div>

      <TimelineVisualization scenes={editedScenes} totalDuration={totalDuration} />

      <div className="scene-timing-list">
        {editedScenes.map((scene) => (
          <SceneTimingCard
            key={scene.scene_number || scene.SceneNumber}
            scene={scene}
            onUpdate={updateScene}
            isApplying={isApplying}
          />
        ))}
      </div>
    </div>
  );
}

TimingEditor.propTypes = {
  jobId: PropTypes.string.isRequired,
  initialScenes: PropTypes.arrayOf(PropTypes.object).isRequired,
  onSaveComplete: PropTypes.func,
};

export default TimingEditor;
