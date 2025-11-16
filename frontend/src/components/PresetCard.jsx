import { Sparkles } from "lucide-react";

/**
 * PresetCard component - displays a brand style preset for quick video creation
 * @param {Object} preset - Preset object
 * @param {string} preset.id - Preset ID
 * @param {string} preset.name - Preset name
 * @param {string} preset.description - Preset description
 * @param {string} preset.style - Visual style
 * @param {Array<string>} preset.color_palette - Array of hex color codes
 * @param {string} preset.music_mood - Music mood
 * @param {Function} onClick - Click handler
 */
function PresetCard({ preset, onClick }) {
  return (
    <button className="preset-card" onClick={() => onClick(preset)}>
      <div className="preset-card-header">
        <Sparkles size={20} className="preset-card-icon" />
        <h3 className="preset-card-name">{preset.name}</h3>
      </div>

      <p className="preset-card-description">{preset.description}</p>

      <div className="preset-card-meta">
        <span className="preset-card-style">{preset.style}</span>
        {preset.color_palette && preset.color_palette.length > 0 && (
          <div className="preset-card-colors">
            {preset.color_palette.slice(0, 4).map((color, index) => (
              <div
                key={index}
                className="preset-card-color"
                style={{ backgroundColor: color }}
                title={color}
              />
            ))}
          </div>
        )}
      </div>
    </button>
  );
}

export default PresetCard;
