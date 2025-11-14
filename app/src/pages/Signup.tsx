import { useState } from "react";
import { Link } from "react-router-dom";
import "../styles/auth.css";

function Signup() {
  const [values, setValues] = useState({
    name: "",
    email: "",
    password: "",
    confirmPassword: "",
  });
  const [errors, setErrors] = useState({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleChange = (event) => {
    const { name, value } = event.target;
    setValues((prev) => ({ ...prev, [name]: value }));
    setErrors((prev) => ({ ...prev, [name]: undefined }));
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

  const handleSubmit = (event) => {
    event.preventDefault();
    const nextErrors = validate();
    setErrors(nextErrors);

    if (Object.keys(nextErrors).length > 0) {
      return;
    }

    setIsSubmitting(true);

    setTimeout(() => {
      console.log("Signup submitted", values);
      setIsSubmitting(false);
    }, 1400);
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
            Launch the Ad Creative Pipeline, upload your product details, and
            let Omnigen handle shot lists, captions, and aspect ratios with
            aurora-grade polish.
          </p>
          <ul className="brand-points">
            <li>Free trial with 5 video generations</li>
            <li>No credit card required to start</li>
            <li>Export 16:9 · 9:16 · 1:1 instantly</li>
          </ul>
        </section>

        <section className="auth-card accent-shift">
          <div className="auth-card-header">
            <h1>Create account</h1>
            <p>Spin up your ad studio and publish your first video today.</p>
          </div>

          <form className="auth-form" onSubmit={handleSubmit} noValidate>
            <div className="form-field">
              <label htmlFor="signup-name">Full name</label>
              <input
                id="signup-name"
                name="name"
                type="text"
                className={`form-input ${errors.name ? "error" : ""}`}
                value={values.name}
                onChange={handleChange}
                autoComplete="name"
                aria-invalid={Boolean(errors.name)}
                aria-describedby={fieldErrorId("name")}
                required
              />
              {errors.name && (
                <p className="input-error" id="signup-name-error" role="status">
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
                className={`form-input ${errors.email ? "error" : ""}`}
                value={values.email}
                onChange={handleChange}
                autoComplete="email"
                aria-invalid={Boolean(errors.email)}
                aria-describedby={fieldErrorId("email")}
                required
              />
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
                className={`form-input ${errors.password ? "error" : ""}`}
                value={values.password}
                onChange={handleChange}
                autoComplete="new-password"
                aria-invalid={Boolean(errors.password)}
                aria-describedby={fieldErrorId("password")}
                required
              />
              {errors.password && (
                <p
                  className="input-error"
                  id="signup-password-error"
                  role="status"
                >
                  {errors.password}
                </p>
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
                }`}
                value={values.confirmPassword}
                onChange={handleChange}
                autoComplete="new-password"
                aria-invalid={Boolean(errors.confirmPassword)}
                aria-describedby={fieldErrorId("confirmPassword")}
                required
              />
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
              <span>Password must be 8+ characters</span>
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

          <p className="auth-support">
            Already have an account? <Link to="/login">Sign in</Link>
          </p>
        </section>
      </div>
    </main>
  );
}

export default Signup;
