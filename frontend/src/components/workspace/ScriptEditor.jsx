import { useState } from "react";
import PropTypes from "prop-types";

/**
 * ScriptEditor component for editing the text that appears on the video
 * @param {Object} props - Component props
 * @param {string} props.script - Initial script text
 * @param {Function} props.onChange - Callback when script changes
 */
function ScriptEditor({ script = "", onChange }) {
  const [localScript, setLocalScript] = useState(script);

  const handleChange = (e) => {
    const newScript = e.target.value;
    setLocalScript(newScript);
    if (onChange) {
      onChange(newScript);
    }
  };

  return (
    <div className="script-editor">
      <div className="script-editor-header">
        <h3 className="script-editor-title">Script</h3>
        <p className="script-editor-subtitle">
          Edit the text that appears on your video
        </p>
      </div>
      <textarea
        className="script-editor-textarea"
        value={localScript}
        onChange={handleChange}
        placeholder="Enter your script text here..."
        rows={12}
      />
    </div>
  );
}

ScriptEditor.propTypes = {
  script: PropTypes.string,
  onChange: PropTypes.func,
};

ScriptEditor.defaultProps = {
  script: "",
  onChange: undefined,
};

export default ScriptEditor;

