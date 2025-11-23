import { useState } from 'react';
import PropTypes from 'prop-types';
import { X, Check, Trash2 } from 'lucide-react';

function SceneEditor({ scene, onSave, onCancel, onSkip, isSaving }) {
  const [prompt, setPrompt] = useState(scene.generation_prompt || scene.action || '');
  const [duration, setDuration] = useState(scene.duration || 5);

  const handleSave = () => {
    onSave({ prompt, duration: parseFloat(duration) });
  };

  return (
    <div className="scene-editor">
      <div className="scene-editor-header">
        <h4>Edit Scene {scene.scene_number}</h4>
        <button className="btn-icon" onClick={onCancel} disabled={isSaving}>
          <X size={18} />
        </button>
      </div>

      <div className="scene-editor-body">
        <div className="form-group">
          <label>Prompt / Action</label>
          <textarea
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            rows={3}
            disabled={isSaving}
          />
        </div>

        <div className="form-group">
          <label>Duration (seconds)</label>
          <input
            type="number"
            value={duration}
            onChange={(e) => setDuration(e.target.value)}
            min={1}
            max={10}
            step={0.5}
            disabled={isSaving}
          />
        </div>
      </div>

      <div className="scene-editor-footer">
        <button 
          className="btn-danger-ghost" 
          onClick={() => onSkip(scene.scene_number)}
          disabled={isSaving}
        >
          <Trash2 size={16} />
          Skip Scene
        </button>
        <div className="scene-editor-actions">
          <button className="btn-ghost" onClick={onCancel} disabled={isSaving}>
            Cancel
          </button>
          <button className="btn-primary" onClick={handleSave} disabled={isSaving}>
            {isSaving ? 'Saving...' : <><Check size={16} /> Save Changes</>}
          </button>
        </div>
      </div>
    </div>
  );
}

SceneEditor.propTypes = {
  scene: PropTypes.object.isRequired,
  onSave: PropTypes.func.isRequired,
  onCancel: PropTypes.func.isRequired,
  onSkip: PropTypes.func.isRequired,
  isSaving: PropTypes.bool
};

export default SceneEditor;
