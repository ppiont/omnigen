import { useState } from "react";
import "../../styles/create.css";

const brandPresets = [
  {
    id: "default",
    name: "Default Style",
    colors: ["#7cff00", "#b44cff"],
  },
  {
    id: "tech-minimal",
    name: "Tech Minimal",
    colors: ["#00ffd1", "#ffffff"],
  },
  {
    id: "bold-vibrant",
    name: "Bold Vibrant",
    colors: ["#ff00ff", "#ffa500"],
  },
  {
    id: "corporate-clean",
    name: "Corporate Clean",
    colors: ["#0066cc", "#333333"],
  },
  {
    id: "warm-organic",
    name: "Warm Organic",
    colors: ["#ff6b35", "#f7931e"],
  },
  {
    id: "custom",
    name: "+ Create Custom Preset",
    colors: [],
    isCustom: true,
  },
];

function BrandPresetSelector({ selectedPreset, onChange }) {
  const [showTooltip, setShowTooltip] = useState(false);

  const handleChange = (e) => {
    const presetId = e.target.value;
    const preset = brandPresets.find((p) => p.id === presetId);

    if (preset?.isCustom) {
      setShowTooltip(true);
      setTimeout(() => setShowTooltip(false), 3000);
      return;
    }

    onChange(presetId);
  };

  const selectedPresetData = brandPresets.find(
    (p) => p.id === selectedPreset
  );

  return (
    <div className="brand-preset-selector">
      <label className="option-label">Brand Preset</label>
      <div className="preset-select-wrapper">
        <select
          className="dropdown-field preset-dropdown"
          value={selectedPreset}
          onChange={handleChange}
        >
          {brandPresets.map((preset) => (
            <option key={preset.id} value={preset.id}>
              {preset.name}
            </option>
          ))}
        </select>

        {/* Color swatches */}
        {selectedPresetData && selectedPresetData.colors.length > 0 && (
          <div className="preset-color-swatches">
            {selectedPresetData.colors.map((color, idx) => (
              <div
                key={idx}
                className="color-swatch"
                style={{ backgroundColor: color }}
                title={color}
              />
            ))}
          </div>
        )}

        {/* Tooltip for custom preset */}
        {showTooltip && (
          <div className="preset-tooltip">
            <span className="tooltip-icon">💡</span>
            Coming soon - Define your brand colors, fonts, and style
          </div>
        )}
      </div>
      <p className="option-helper">
        Apply consistent brand styling across your video
      </p>
    </div>
  );
}

export default BrandPresetSelector;
