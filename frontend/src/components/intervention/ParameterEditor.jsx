import { useState } from 'react';
import PropTypes from 'prop-types';
import { X, Check } from 'lucide-react';

function ParameterEditor({ currentParams, onSave, onCancel, isSaving }) {
  const [aspectRatio, setAspectRatio] = useState(currentParams.aspectRatio || '16:9');
  const [style, setStyle] = useState(currentParams.style || 'cinematic');

  const handleSave = () => {
    onSave({ aspectRatio, style });
  };

  return (
    <div className="modal-backdrop">
      <div className="modal-card parameter-editor-modal">
        <div className="modal-header">
          <h3>Edit Parameters</h3>
          <button className="btn-icon" onClick={onCancel} disabled={isSaving}>
            <X size={20} />
          </button>
        </div>

        <div className="modal-body">
          <p className="modal-description">
            Changes will apply to all remaining ungenerated scenes.
          </p>

          <div className="form-group">
            <label>Aspect Ratio</label>
            <select 
              value={aspectRatio} 
              onChange={(e) => setAspectRatio(e.target.value)}
              disabled={isSaving}
            >
              <option value="16:9">16:9 (Widescreen)</option>
              <option value="9:16">9:16 (Vertical)</option>
              <option value="1:1">1:1 (Square)</option>
            </select>
          </div>

          <div className="form-group">
            <label>Visual Style</label>
            <select 
              value={style} 
              onChange={(e) => setStyle(e.target.value)}
              disabled={isSaving}
            >
              <option value="cinematic">Cinematic</option>
              <option value="documentary">Documentary</option>
              <option value="animated">Animated</option>
              <option value="minimalist">Minimalist</option>
            </select>
          </div>
        </div>

        <div className="modal-footer">
          <button className="btn-ghost" onClick={onCancel} disabled={isSaving}>
            Cancel
          </button>
          <button className="btn-primary" onClick={handleSave} disabled={isSaving}>
            {isSaving ? 'Applying...' : <><Check size={16} /> Apply Changes</>}
          </button>
        </div>
      </div>
    </div>
  );
}

ParameterEditor.propTypes = {
  currentParams: PropTypes.object.isRequired,
  onSave: PropTypes.func.isRequired,
  onCancel: PropTypes.func.isRequired,
  isSaving: PropTypes.bool
};

export default ParameterEditor;
