import PropTypes from 'prop-types';
import { Play, Pause, Settings, Edit3 } from 'lucide-react';

function JobControls({ isPaused, onPause, onResume, onEditParameters, disabled }) {
  return (
    <div className="job-controls">
      <div className="job-controls-main">
        {isPaused ? (
          <button 
            className="btn-primary btn-resume" 
            onClick={onResume}
            disabled={disabled}
            title="Resume generation"
          >
            <Play size={18} fill="currentColor" />
            <span>Resume</span>
          </button>
        ) : (
          <button 
            className="btn-secondary btn-pause" 
            onClick={onPause}
            disabled={disabled}
            title="Pause generation after current scene"
          >
            <Pause size={18} fill="currentColor" />
            <span>Pause</span>
          </button>
        )}
      </div>
      
      <div className="job-controls-secondary">
        <button 
          className="btn-ghost" 
          onClick={onEditParameters}
          disabled={disabled}
          title="Edit global parameters for remaining scenes"
        >
          <Settings size={18} />
          <span>Parameters</span>
        </button>
      </div>
    </div>
  );
}

JobControls.propTypes = {
  isPaused: PropTypes.bool.isRequired,
  onPause: PropTypes.func.isRequired,
  onResume: PropTypes.func.isRequired,
  onEditParameters: PropTypes.func.isRequired,
  disabled: PropTypes.bool
};

export default JobControls;
