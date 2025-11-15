function ToggleSwitch({ checked, onChange, label }) {
  return (
    <button
      type="button"
      className={`toggle-switch ${checked ? "is-active" : ""}`}
      onClick={onChange}
      aria-label={label}
    >
      <span className="toggle-handle" />
    </button>
  );
}

export default ToggleSwitch;

