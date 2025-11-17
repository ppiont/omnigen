import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Check } from "lucide-react";
import { useAuth } from "../contexts/useAuth.js";
import PasswordStrength from "../components/PasswordStrength";
import "../styles/auth.css";

function Signup() {
  const navigate = useNavigate();
  const { signup, confirmSignup, resendCode } = useAuth();
  const [step, setStep] = useState("signup"); // 'signup' or 'verify'
  const [values, setValues] = useState({
    name: "",
    email: "",
    password: "",
    confirmPassword: "",
  });
  const [verificationCode, setVerificationCode] = useState("");
  const [errors, setErrors] = useState({});
  const [touched, setTouched] = useState({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [successMessage, setSuccessMessage] = useState("");

  const handleChange = (event) => {
    const { name, value } = event.target;
    setValues((prev) => ({ ...prev, [name]: value }));
    setErrors((prev) => ({ ...prev, [name]: undefined }));

    // Validate password field on every keystroke if already touched
    if (name === "password" && touched.password) {
      // Use setTimeout to validate with the new value
      setTimeout(() => validateField("password"), 0);
    }
  };

  const handleBlur = (field) => {
    setTouched((prev) => ({ ...prev, [field]: true }));
    validateField(field);
  };

  const validateField = (field) => {
    const nextErrors = { ...errors };

    switch (field) {
      case "name":
        if (!values.name.trim()) {
          nextErrors.name = "Name is required.";
        } else {
          delete nextErrors.name;
        }
        break;

      case "email":
        if (!values.email.trim()) {
          nextErrors.email = "Email is required.";
        } else if (!/^\S+@\S+\.\S+$/.test(values.email)) {
          nextErrors.email = "Enter a valid email address.";
        } else {
          delete nextErrors.email;
        }
        break;

      case "password":
        if (!values.password.trim()) {
          nextErrors.password = "Password is required.";
        } else if (values.password.length < 8) {
          nextErrors.password = "Password must be at least 8 characters.";
        } else {
          delete nextErrors.password;
        }
        break;

      case "confirmPassword":
        if (!values.confirmPassword.trim()) {
          nextErrors.confirmPassword = "Confirm your password.";
        } else if (values.confirmPassword !== values.password) {
          nextErrors.confirmPassword = "Passwords do not match.";
        } else {
          delete nextErrors.confirmPassword;
        }
        break;
    }

    setErrors(nextErrors);
  };

  const isFieldValid = (field) => {
    return touched[field] && values[field] && !errors[field];
  };

  const validate = () => {
    const nextErrors = {};

    if (!values.name.trim()) {
      nextErrors.name = "Name is required.";
    }

    if (!values.email.trim()) {
      nextErrors.email = "Email is required.";
    } else if (!/^\S+@\S+\.\S+$/.test(values.email)) {
      nextErrors.email = "Enter a valid email address.";
    }

    if (!values.password.trim()) {
      nextErrors.password = "Password is required.";
    } else if (values.password.length < 8) {
      nextErrors.password = "Password must be at least 8 characters.";
    }

    if (!values.confirmPassword.trim()) {
      nextErrors.confirmPassword = "Confirm your password.";
    } else if (values.confirmPassword !== values.password) {
      nextErrors.confirmPassword = "Passwords do not match.";
    }

    return nextErrors;
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    const nextErrors = validate();
    setErrors(nextErrors);

    if (Object.keys(nextErrors).length > 0) {
      return;
    }

    setIsSubmitting(true);
    setErrors({});
    setSuccessMessage("");

    try {
      const result = await signup(values.name, values.email, values.password);

      if (result.success) {
        if (result.userConfirmed) {
          // User is auto-confirmed, redirect to login
          setSuccessMessage(result.message);
          setTimeout(() => navigate("/login"), 2000);
        } else {
          // Need email verification
          setSuccessMessage(result.message);
          setStep("verify");
        }
      } else {
        setErrors({ form: result.error });
      }
    } catch (error) {
      console.error("Signup error:", error);
      setErrors({ form: "An unexpected error occurred. Please try again." });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleVerify = async (event) => {
    event.preventDefault();

    if (!verificationCode.trim()) {
      setErrors({ code: "Verification code is required" });
      return;
    }

    setIsSubmitting(true);
    setErrors({});
    setSuccessMessage("");

    try {
      const result = await confirmSignup(values.email, verificationCode);

      if (result.success) {
        setSuccessMessage(result.message);
        // Redirect to login after 2 seconds
        setTimeout(() => navigate("/login"), 2000);
      } else {
        setErrors({ form: result.error });
      }
    } catch (error) {
      console.error("Verification error:", error);
      setErrors({ form: "An unexpected error occurred. Please try again." });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleResendCode = async () => {
    setIsSubmitting(true);
    setErrors({});
    setSuccessMessage("");

    try {
      const result = await resendCode(values.email);
      if (result.success) {
        setSuccessMessage(result.message);
      } else {
        setErrors({ form: result.error });
      }
    } catch (error) {
      console.error("Resend code error:", error);
      setErrors({ form: "Failed to resend code. Please try again." });
    } finally {
      setIsSubmitting(false);
    }
  };

  const fieldErrorId = (field) =>
    errors[field] ? `signup-${field}-error` : undefined;

  return (
    <main className="auth-page">
      <div className="auth-grid">
        <section className="auth-brand">
          <span className="brand-badge">Omnigen</span>
          <h1>Start generating video ads in under 60 seconds.</h1>
          <p>
            Join the Ad Creative Pipeline and create production-ready video ads
            with multi-format exports, brand-consistent styling, and
            professional quality every time.
          </p>
          <ul className="brand-points">
            <li>10K+ videos generated for leading brands</li>
            <li>98% success rate across all renders</li>
            <li>16:9 · 9:16 · 1:1 outputs in one pass</li>
          </ul>
        </section>

        <section className="auth-card accent-shift">
          <div className="auth-card-header">
            <h1>
              {step === "signup" ? "Create account" : "Verify your email"}
            </h1>
            <p>
              {step === "signup"
                ? "Spin up your ad studio and publish your first video today."
                : "Enter the code we sent to your email"}
            </p>
          </div>

          {successMessage && (
            <div className="auth-success" role="alert">
              {successMessage}
            </div>
          )}

          {errors.form && (
            <div className="auth-error" role="alert" aria-live="polite">
              {errors.form}
            </div>
          )}

          {step === "signup" ? (
            <form className="auth-form" onSubmit={handleSubmit} noValidate>
              <div className="form-field">
                <label htmlFor="signup-name">Full name</label>
                <input
                  id="signup-name"
                  name="name"
                  type="text"
                  className={`form-input ${errors.name ? "error" : ""} ${
                    isFieldValid("name") ? "success" : ""
                  }`}
                  value={values.name}
                  onChange={handleChange}
                  onBlur={() => handleBlur("name")}
                  autoComplete="name"
                  aria-invalid={Boolean(errors.name)}
                  aria-describedby={fieldErrorId("name")}
                  required
                />
                {isFieldValid("name") && (
                  <Check size={18} className="field-icon" />
                )}
                {errors.name && (
                  <p
                    className="input-error"
                    id="signup-name-error"
                    role="status"
                  >
                    {errors.name}
                  </p>
                )}
              </div>

              <div className="form-field">
                <label htmlFor="signup-email">Email</label>
                <input
                  id="signup-email"
                  name="email"
                  type="email"
                  className={`form-input ${errors.email ? "error" : ""} ${
                    isFieldValid("email") ? "success" : ""
                  }`}
                  value={values.email}
                  onChange={handleChange}
                  onBlur={() => handleBlur("email")}
                  autoComplete="email"
                  aria-invalid={Boolean(errors.email)}
                  aria-describedby={fieldErrorId("email")}
                  required
                />
                {isFieldValid("email") && (
                  <Check size={18} className="field-icon" />
                )}
                {errors.email && (
                  <p
                    className="input-error"
                    id="signup-email-error"
                    role="status"
                  >
                    {errors.email}
                  </p>
                )}
              </div>

              <div className="form-field">
                <label htmlFor="signup-password">Password</label>
                <input
                  id="signup-password"
                  name="password"
                  type="password"
                  className={`form-input ${errors.password ? "error" : ""} ${
                    isFieldValid("password") ? "success" : ""
                  }`}
                  value={values.password}
                  onChange={handleChange}
                  onBlur={() => handleBlur("password")}
                  autoComplete="new-password"
                  aria-invalid={Boolean(errors.password)}
                  aria-describedby={fieldErrorId("password")}
                  required
                />
                {isFieldValid("password") && (
                  <Check size={18} className="field-icon" />
                )}
                {errors.password && (
                  <p
                    className="input-error"
                    id="signup-password-error"
                    role="status"
                  >
                    {errors.password}
                  </p>
                )}
                {values.password && (
                  <PasswordStrength password={values.password} />
                )}
              </div>

              <div className="form-field">
                <label htmlFor="signup-confirmPassword">Confirm password</label>
                <input
                  id="signup-confirmPassword"
                  name="confirmPassword"
                  type="password"
                  className={`form-input ${
                    errors.confirmPassword ? "error" : ""
                  } ${isFieldValid("confirmPassword") ? "success" : ""}`}
                  value={values.confirmPassword}
                  onChange={handleChange}
                  onBlur={() => handleBlur("confirmPassword")}
                  autoComplete="new-password"
                  aria-invalid={Boolean(errors.confirmPassword)}
                  aria-describedby={fieldErrorId("confirmPassword")}
                  required
                />
                {isFieldValid("confirmPassword") && (
                  <Check size={18} className="field-icon" />
                )}
                {errors.confirmPassword && (
                  <p
                    className="input-error"
                    id="signup-confirmPassword-error"
                    role="status"
                  >
                    {errors.confirmPassword}
                  </p>
                )}
              </div>

              <div className="auth-meta">
                <span>Password must be 8+ characters with uppercase, lowercase, number, and special character</span>
                <Link to="/login">Sign in</Link>
              </div>

              <button
                type="submit"
                className="auth-submit"
                disabled={isSubmitting}
              >
                {isSubmitting ? "Creating…" : "Create account"}
              </button>
            </form>
          ) : (
            <form className="auth-form" onSubmit={handleVerify} noValidate>
              <div className="form-field">
                <label htmlFor="verification-code">Verification Code</label>
                <input
                  id="verification-code"
                  name="code"
                  type="text"
                  className={`form-input ${errors.code ? "error" : ""}`}
                  value={verificationCode}
                  onChange={(e) => {
                    setVerificationCode(e.target.value);
                    setErrors((prev) => ({ ...prev, code: undefined }));
                  }}
                  placeholder="Enter 6-digit code"
                  aria-invalid={Boolean(errors.code)}
                  aria-describedby={
                    errors.code ? "verification-code-error" : undefined
                  }
                  required
                  disabled={isSubmitting}
                />
                {errors.code && (
                  <p
                    className="input-error"
                    id="verification-code-error"
                    role="status"
                  >
                    {errors.code}
                  </p>
                )}
              </div>

              <div className="auth-meta">
                <button
                  type="button"
                  onClick={handleResendCode}
                  disabled={isSubmitting}
                  className="auth-link-button"
                >
                  Resend code
                </button>
                <button
                  type="button"
                  onClick={() => setStep("signup")}
                  disabled={isSubmitting}
                  className="auth-link-button"
                >
                  Back to signup
                </button>
              </div>

              <button
                type="submit"
                className="auth-submit"
                disabled={isSubmitting}
              >
                {isSubmitting ? "Verifying…" : "Verify Email"}
              </button>
            </form>
          )}

          <p className="auth-support">
            Already have an account? <Link to="/login">Sign in</Link>
          </p>
        </section>
      </div>
    </main>
  );
}

export default Signup;
