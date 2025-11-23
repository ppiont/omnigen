import PropTypes from 'prop-types';
import { Clock, RotateCcw } from 'lucide-react';
import TransitionSelector from './TransitionSelector';

function SceneTimingCard({ scene, onUpdate, isApplying }) {
  const sceneNumber = scene.scene_number || scene.SceneNumber;
  const duration = scene.duration || 5;

  return (
    <div className="scene-timing-card">
      <div className="scene-timing-header">
        <h4>Scene {sceneNumber}</h4>
        <span className={`scene-status-badge ${scene.needsRegeneration ? 'status-warning' : 'status-ok'}`}>
          {scene.needsRegeneration ? 'Needs Regen' : 'Ready'}
        </span>
      </div>
      
      <div className="scene-timing-body">
        <div className="form-group duration-control">
          <label><Clock size={14} /> Duration</label>
          <div className="duration-options">
            <button
              className={`duration-btn ${duration === 5 ? 'active' : ''}`}
              onClick={() => onUpdate(sceneNumber, { duration: 5 })}
              disabled={isApplying}
            >
              5s
            </button>
            <button
              className={`duration-btn ${duration === 10 ? 'active' : ''}`}
              onClick={() => onUpdate(sceneNumber, { duration: 10 })}
              disabled={isApplying}
            >
              10s
            </button>
          </div>
        </div>

        <div className="transitions-row">
          <TransitionSelector
            type="in"
            value={scene.transition_in}
            onChange={(val) => onUpdate(sceneNumber, { transition_in: val })}
            disabled={isApplying || sceneNumber === 1}
          />
          <TransitionSelector
            type="out"
            value={scene.transition_out}
            onChange={(val) => onUpdate(sceneNumber, { transition_out: val })}
            disabled={isApplying}
          />
        </div>
      </div>
    </div>
  );
}

SceneTimingCard.propTypes = {
  scene: PropTypes.object.isRequired,
  onUpdate: PropTypes.func.isRequired,
  isApplying: PropTypes.bool,
};

export default SceneTimingCard;
