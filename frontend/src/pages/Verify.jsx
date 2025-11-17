import { useState, useEffect } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { useAuth } from "../contexts/useAuth.js";
import "../styles/auth.css";

function Verify() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { confirmSignup, resendCode } = useAuth();

  const [email, setEmail] = useState(searchParams.get("email") || "");
  const [verificationCode, setVerificationCode] = useState("");
  const [errors, setErrors] = useState({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [successMessage, setSuccessMessage] = useState("");

  useEffect(() => {
    const emailParam = searchParams.get("email");
    if (emailParam) {
      setEmail(emailParam);
    }
  }, [searchParams]);

  const handleVerify = async (event) => {
    event.preventDefault();

    // Validation
    const nextErrors = {};
    if (!email.trim()) {
      nextErrors.email = "Email is required";
    } else if (!/^\S+@\S+\.\S+$/.test(email)) {
      nextErrors.email = "Enter a valid email address";
    }

    if (!verificationCode.trim()) {
      nextErrors.code = "Verification code is required";
    }

    if (Object.keys(nextErrors).length > 0) {
      setErrors(nextErrors);
      return;
    }

    setIsSubmitting(true);
    setErrors({});
    setSuccessMessage("");

    try {
      const result = await confirmSignup(email, verificationCode);

      if (result.success) {
        setSuccessMessage("Email verified successfully! Redirecting to login...");
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
    if (!email.trim()) {
      setErrors({ email: "Email is required to resend code" });
      return;
    }

    setIsSubmitting(true);
    setErrors({});
    setSuccessMessage("");

    try {
      const result = await resendCode(email);
      if (result.success) {
        setSuccessMessage("Verification code sent! Check your email.");
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

  return (
    <main className="auth-page">
      <div className="auth-grid">
        <section className="auth-brand">
          <span className="brand-badge">Omnigen</span>
          <h1>Almost there!</h1>
          <p>
            Verify your email to unlock the full power of the Ad Creative
            Pipeline. Your account is just one step away from generating
            production-ready video ads.
          </p>
          <ul className="brand-points">
            <li>Secure account activation</li>
            <li>Instant access after verification</li>
            <li>Code valid for 24 hours</li>
          </ul>
        </section>

        <section className="auth-card accent-shift">
          <div className="auth-card-header">
            <h1>Verify your email</h1>
            <p>Enter the verification code sent to your email</p>
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

          <form className="auth-form" onSubmit={handleVerify} noValidate>
            <div className="form-field">
              <label htmlFor="verify-email">Email</label>
              <input
                id="verify-email"
                name="email"
                type="email"
                className={`form-input ${errors.email ? "error" : ""}`}
                value={email}
                onChange={(e) => {
                  setEmail(e.target.value);
                  setErrors((prev) => ({ ...prev, email: undefined }));
                }}
                autoComplete="email"
                aria-invalid={Boolean(errors.email)}
                aria-describedby={errors.email ? "verify-email-error" : undefined}
                required
                disabled={isSubmitting}
              />
              {errors.email && (
                <p className="input-error" id="verify-email-error" role="status">
                  {errors.email}
                </p>
              )}
            </div>

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
              <Link to="/login" className="auth-link-button">
                Back to login
              </Link>
            </div>

            <button
              type="submit"
              className="auth-submit"
              disabled={isSubmitting}
            >
              {isSubmitting ? "Verifying…" : "Verify Email"}
            </button>
          </form>

          <p className="auth-support">
            Don't have an account? <Link to="/signup">Sign up</Link>
          </p>
        </section>
      </div>
    </main>
  );
}

export default Verify;
