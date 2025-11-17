import { useMemo } from "react";

function PasswordStrength({ password }) {
  const requirements = useMemo(() => {
    if (!password) return null;

    return {
      length: password.length >= 8,
      lowercase: /[a-z]/.test(password),
      uppercase: /[A-Z]/.test(password),
      number: /[0-9]/.test(password),
      special: /[^a-zA-Z0-9]/.test(password),
    };
  }, [password]);

  const strength = useMemo(() => {
    if (!password) return { level: 0, text: "", bars: 0 };

    let score = 0;

    // Length check
    if (password.length >= 8) score++;
    if (password.length >= 12) score++;

    // Character variety checks
    if (/[a-z]/.test(password)) score++;
    if (/[A-Z]/.test(password)) score++;
    if (/[0-9]/.test(password)) score++;
    if (/[^a-zA-Z0-9]/.test(password)) score++;

    // Determine strength level
    if (score <= 2) return { level: 1, text: "Weak", bars: 1 };
    if (score <= 4) return { level: 2, text: "Medium", bars: 2 };
    return { level: 3, text: "Strong", bars: 3 };
  }, [password]);

  if (!password) return null;

  return (
    <div className="password-strength">
      <div className="strength-bars">
        {[1, 2, 3].map((bar) => (
          <div
            key={bar}
            className={`strength-bar ${
              bar <= strength.bars
                ? `filled ${strength.text.toLowerCase()}`
                : ""
            }`}
          />
        ))}
      </div>
      <p className={`strength-text ${strength.text.toLowerCase()}`}>
        {strength.text} password
      </p>
      <ul className="password-requirements">
        <li className={requirements.length ? "met" : "unmet"}>
          {requirements.length ? "✓" : "○"} At least 8 characters
        </li>
        <li className={requirements.uppercase ? "met" : "unmet"}>
          {requirements.uppercase ? "✓" : "○"} Uppercase letter
        </li>
        <li className={requirements.lowercase ? "met" : "unmet"}>
          {requirements.lowercase ? "✓" : "○"} Lowercase letter
        </li>
        <li className={requirements.number ? "met" : "unmet"}>
          {requirements.number ? "✓" : "○"} Number
        </li>
        <li className={requirements.special ? "met" : "unmet"}>
          {requirements.special ? "✓" : "○"} Special character
        </li>
      </ul>
    </div>
  );
}

export default PasswordStrength;
