import PropTypes from 'prop-types';

const TRANSITION_OPTIONS = [
  { value: "none", label: "None" },
  { value: "cut", label: "Cut (Instant)" },
  { value: "fade", label: "Fade" },
  { value: "cross_fade", label: "Cross Fade" },
  { value: "wipe_left", label: "Wipe Left" },
  { value: "wipe_right", label: "Wipe Right" },
  { value: "iris_in", label: "Iris In" },
  { value: "iris_out", label: "Iris Out" },
  { value: "match_cut", label: "Match Cut" },
  { value: "jump_cut", label: "Jump Cut" },
  { value: "zoom_transition", label: "Zoom Transition" },
];

function TransitionSelector({ value, onChange, type, disabled }) {
  return (
    <div className="transition-selector">
      <label className="transition-label">
        {type === 'in' ? 'Transition In' : 'Transition Out'}
      </label>
      <select
        value={value || "none"}
        onChange={(e) => onChange(e.target.value)}
        disabled={disabled}
        className="transition-select"
      >
        {TRANSITION_OPTIONS.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </div>
  );
}

TransitionSelector.propTypes = {
  value: PropTypes.string,
  onChange: PropTypes.func.isRequired,
  type: PropTypes.oneOf(['in', 'out']).isRequired,
  disabled: PropTypes.bool,
};

export default TransitionSelector;
