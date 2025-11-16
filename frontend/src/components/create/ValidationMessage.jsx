import "../../styles/create.css";

/**
 * ValidationMessage component for displaying inline validation errors
 * @param {string} message - Error message to display
 * @param {string} type - Type of message: 'error', 'warning', 'info', 'success'
 */
function ValidationMessage({ message, type = "error" }) {
  if (!message) return null;

  const icons = {
    error: "❌",
    warning: "⚠️",
    info: "ℹ️",
    success: "✅",
  };

  return (
    <div className={`validation-message validation-${type}`} role="alert">
      <span className="validation-icon">{icons[type]}</span>
      <span className="validation-text">{message}</span>
    </div>
  );
}

export default ValidationMessage;
