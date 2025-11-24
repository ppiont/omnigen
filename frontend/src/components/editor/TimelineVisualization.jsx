import PropTypes from 'prop-types';

function TimelineVisualization({ scenes, totalDuration }) {
  return (
    <div className="timeline-visualization">
      <div className="timeline-track">
        {scenes.map((scene, index) => {
          const duration = scene.duration || 5;
          const widthPercent = (duration / totalDuration) * 100;
          const sceneNumber = scene.scene_number || scene.SceneNumber || index + 1;
          
          return (
            <div 
              key={sceneNumber} 
              className={`timeline-block duration-${duration}s`}
              style={{ width: `${widthPercent}%` }}
              title={`Scene ${sceneNumber}: ${duration}s`}
            >
              <span className="scene-label">S{sceneNumber}</span>
              <span className="duration-label">{duration}s</span>
              
              {scene.transition_out && scene.transition_out !== 'none' && (
                <div className="transition-indicator" title={`Transition: ${scene.transition_out}`} />
              )}
            </div>
          );
        })}
      </div>
      <div className="timeline-ruler">
        <span>0s</span>
        <span>{totalDuration}s</span>
      </div>
    </div>
  );
}

TimelineVisualization.propTypes = {
  scenes: PropTypes.arrayOf(PropTypes.object).isRequired,
  totalDuration: PropTypes.number.isRequired,
};

export default TimelineVisualization;
