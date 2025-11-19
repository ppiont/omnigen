import { useState } from "react";
import ToggleSwitch from "../ToggleSwitch.jsx";
import "../../styles/create.css";

/**
 * BatchGenerationToggle - Disabled toggle for future batch generation feature
 * Shows tooltip when user tries to interact
 */
function BatchGenerationToggle() {
  const [showTooltip, setShowTooltip] = useState(false);

  const handleClick = () => {
    setShowTooltip(true);
    setTimeout(() => setShowTooltip(false), 3000);
  };

  return (
    <div className="batch-toggle-container">
      <label className="option-label">
        Batch Generation
        <span className="feature-badge">Coming Soon</span>
      </label>

      <div className="batch-toggle-wrapper" onClick={handleClick}>
        <div className="toggle-item batch-toggle-disabled">
          <ToggleSwitch
            checked={false}
            onChange={() => {}}
            label="Batch Generation"
            disabled={true}
          />
          <span className="toggle-label disabled">
            Generate multiple variations
          </span>
        </div>

        {showTooltip && (
          <div className="batch-tooltip">
            <div className="tooltip-header">
              <span className="tooltip-icon">💡</span>
              <strong>Pro Feature</strong>
            </div>
            <p>
              Generate multiple variations simultaneously - Coming in Pro plan
            </p>
          </div>
        )}
      </div>

      <p className="option-helper disabled">
        Create 3, 5, or 10 video variations at once (Available in Pro plan)
      </p>
    </div>
  );
}

export default BatchGenerationToggle;
