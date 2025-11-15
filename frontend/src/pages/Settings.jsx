import { useState } from "react";
import { Eye, EyeOff, Check } from "lucide-react";
import AppLayout from "../components/AppLayout.jsx";
import PasswordStrength from "../components/PasswordStrength";
import "../styles/settings.css";

function Settings() {
  const [values, setValues] = useState({
    currentPassword: "",
    newPassword: "",
    confirmPassword: "",
  });
  const [touched, setTouched] = useState({});
  const [errors, setErrors] = useState({});
  const [showCurrentPassword, setShowCurrentPassword] = useState(false);
  const [showNewPassword, setShowNewPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [successMessage, setSuccessMessage] = useState("");

  const handleChange = (event) => {
    const { name, value } = event.target;

    setValues((prev) => {
      const updated = { ...prev, [name]: value };

      if (touched[name]) {
        validateField(name, updated);
      }

      if (name === "newPassword" && touched.confirmPassword) {
        validateField("confirmPassword", updated);
      }

      return updated;
    });
  };

  const handleBlur = (field) => {
    setTouched((prev) => ({ ...prev, [field]: true }));
    validateField(field);
  };

  const validateField = (field, nextValues = values) => {
    setErrors((prev) => {
      const updated = { ...prev };

      switch (field) {
        case "currentPassword":
          if (!nextValues.currentPassword.trim()) {
            updated.currentPassword = "Current password is required.";
          } else {
            delete updated.currentPassword;
          }
          break;
        case "newPassword":
          if (!nextValues.newPassword.trim()) {
            updated.newPassword = "New password is required.";
          } else if (nextValues.newPassword.length < 8) {
            updated.newPassword = "Password must be at least 8 characters.";
          } else {
            delete updated.newPassword;
          }
          break;
        case "confirmPassword":
          if (!nextValues.confirmPassword.trim()) {
            updated.confirmPassword = "Confirm your new password.";
          } else if (nextValues.confirmPassword !== nextValues.newPassword) {
            updated.confirmPassword = "Passwords do not match.";
          } else {
            delete updated.confirmPassword;
          }
          break;
        default:
          break;
      }

      return updated;
    });
  };

  const validateForm = (nextValues = values) => {
    const nextErrors = {};

    if (!nextValues.currentPassword.trim()) {
      nextErrors.currentPassword = "Current password is required.";
    }

    if (!nextValues.newPassword.trim()) {
      nextErrors.newPassword = "New password is required.";
    } else if (nextValues.newPassword.length < 8) {
      nextErrors.newPassword = "Password must be at least 8 characters.";
    }

    if (!nextValues.confirmPassword.trim()) {
      nextErrors.confirmPassword = "Confirm your new password.";
    } else if (nextValues.confirmPassword !== nextValues.newPassword) {
      nextErrors.confirmPassword = "Passwords do not match.";
    }

    return nextErrors;
  };

  const isFieldValid = (field) =>
    Boolean(touched[field] && values[field]?.trim() && !errors[field]);

  const hasValidationErrors =
    !values.currentPassword.trim() ||
    !values.newPassword.trim() ||
    values.newPassword.length < 8 ||
    !values.confirmPassword.trim() ||
    values.confirmPassword !== values.newPassword ||
    Object.keys(errors).length > 0;

  const handleSubmit = (e) => {
    e.preventDefault();
    const nextErrors = validateForm(values);
    setErrors(nextErrors);
    setTouched({
      currentPassword: true,
      newPassword: true,
      confirmPassword: true,
    });

    if (Object.keys(nextErrors).length > 0) {
      return;
    }

    setIsSaving(true);
    setSuccessMessage("");

    // Simulate API call
    setTimeout(() => {
      setIsSaving(false);
      setValues({
        currentPassword: "",
        newPassword: "",
        confirmPassword: "",
      });
      setErrors({});
      setTouched({});
      setSuccessMessage("Password updated successfully!");
    }, 1000);
  };

  return (
    <AppLayout>
      <section className="settings-section">
        <h2 className="section-title">Settings</h2>
        <div className="settings-card">
          <div className="settings-card-header">
            <h3 className="settings-subtitle">Change Password</h3>
            <p className="settings-description">
              Keep your account secure with a strong password.
            </p>
          </div>
          <form className="password-form" onSubmit={handleSubmit} noValidate>
            {successMessage && <p className="form-success">{successMessage}</p>}
            <div className="form-group">
              <label className="form-label" htmlFor="current-password">
                Current Password
              </label>
              <div className="password-input-wrapper">
                <input
                  id="current-password"
                  name="currentPassword"
                  type={showCurrentPassword ? "text" : "password"}
                  className={`form-input ${
                    errors.currentPassword ? "error" : ""
                  } ${isFieldValid("currentPassword") ? "success" : ""}`}
                  value={values.currentPassword}
                  onChange={handleChange}
                  onBlur={() => handleBlur("currentPassword")}
                  placeholder="Enter your current password"
                  required
                />
                {isFieldValid("currentPassword") && (
                  <span className="field-icon" aria-hidden="true">
                    <Check size={16} />
                  </span>
                )}
                <button
                  type="button"
                  className="password-toggle"
                  onClick={() => setShowCurrentPassword(!showCurrentPassword)}
                  aria-label={
                    showCurrentPassword ? "Hide password" : "Show password"
                  }
                >
                  {showCurrentPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                </button>
              </div>
              {errors.currentPassword && touched.currentPassword && (
                <p className="form-error">{errors.currentPassword}</p>
              )}
            </div>

            <div className="form-group">
              <label className="form-label" htmlFor="new-password">
                New Password
              </label>
              <div className="password-input-wrapper">
                <input
                  id="new-password"
                  name="newPassword"
                  type={showNewPassword ? "text" : "password"}
                  className={`form-input ${
                    errors.newPassword ? "error" : ""
                  } ${isFieldValid("newPassword") ? "success" : ""}`}
                  value={values.newPassword}
                  onChange={handleChange}
                  onBlur={() => handleBlur("newPassword")}
                  placeholder="Enter your new password"
                  required
                  minLength={8}
                />
                {isFieldValid("newPassword") && (
                  <span className="field-icon" aria-hidden="true">
                    <Check size={16} />
                  </span>
                )}
                <button
                  type="button"
                  className="password-toggle"
                  onClick={() => setShowNewPassword(!showNewPassword)}
                  aria-label={
                    showNewPassword ? "Hide password" : "Show password"
                  }
                >
                  {showNewPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                </button>
              </div>
              {touched.newPassword && (
                <PasswordStrength password={values.newPassword} />
              )}
              {errors.newPassword && touched.newPassword && (
                <p className="form-error">{errors.newPassword}</p>
              )}
            </div>

            <div className="form-group">
              <label className="form-label" htmlFor="confirm-password">
                Confirm New Password
              </label>
              <div className="password-input-wrapper">
                <input
                  id="confirm-password"
                  name="confirmPassword"
                  type={showConfirmPassword ? "text" : "password"}
                  className={`form-input ${
                    errors.confirmPassword ? "error" : ""
                  } ${isFieldValid("confirmPassword") ? "success" : ""}`}
                  value={values.confirmPassword}
                  onChange={handleChange}
                  onBlur={() => handleBlur("confirmPassword")}
                  placeholder="Confirm your new password"
                  required
                />
                {isFieldValid("confirmPassword") && (
                  <span className="field-icon" aria-hidden="true">
                    <Check size={16} />
                  </span>
                )}
                <button
                  type="button"
                  className="password-toggle"
                  onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                  aria-label={
                    showConfirmPassword ? "Hide password" : "Show password"
                  }
                >
                  {showConfirmPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                </button>
              </div>
              {errors.confirmPassword && touched.confirmPassword && (
                <p className="form-error">{errors.confirmPassword}</p>
              )}
            </div>

            <button
              type="submit"
              className="save-button"
              disabled={isSaving || hasValidationErrors}
            >
              {isSaving ? "Saving..." : "Save Changes"}
            </button>
          </form>
        </div>
      </section>
    </AppLayout>
  );
}

export default Settings;
