import { useMemo } from "react";

function PasswordStrength({ password }) {
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
    </div>
  );
}

export default PasswordStrength;
