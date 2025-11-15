import { useState } from "react";
import { Link } from "react-router-dom";
import "../styles/auth.css";

function Login() {
  const [values, setValues] = useState({ email: "", password: "" });
  const [errors, setErrors] = useState({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleChange = (event) => {
    const { name, value } = event.target;
    setValues((prev) => ({ ...prev, [name]: value }));
    setErrors((prev) => ({ ...prev, [name]: undefined }));
  };

  const validate = () => {
    const nextErrors = {};
    if (!values.email.trim()) {
      nextErrors.email = "Email is required.";
    } else if (!/^\S+@\S+\.\S+$/.test(values.email)) {
      nextErrors.email = "Enter a valid email address.";
    }

    if (!values.password.trim()) {
      nextErrors.password = "Password is required.";
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
      console.log("Login submitted", values);
      setIsSubmitting(false);
    }, 1200);
  };

  const emailErrorId = errors.email ? "login-email-error" : undefined;
  const passwordErrorId = errors.password ? "login-password-error" : undefined;

  return (
    <main className="auth-page">
      <div className="auth-grid">
        <section className="auth-brand">
          <span className="brand-badge">Omnigen</span>
          <h1>Create professional video ads in minutes.</h1>
          <p>
            Sign in to the Ad Creative Pipeline and keep every campaign on
            schedule—multi-format exports, brand-safe templates, and
            production-ready results are waiting.
          </p>
          <ul className="brand-points">
            <li>10K+ videos generated for leading brands</li>
            <li>98% success rate across all renders</li>
            <li>16:9 · 9:16 · 1:1 outputs in one pass</li>
          </ul>
        </section>

        <section className="auth-card">
          <div className="auth-card-header">
            <h1>Welcome back</h1>
            <p>Sign in to keep generating high-performing ad creatives.</p>
          </div>

          <form className="auth-form" onSubmit={handleSubmit} noValidate>
            <div className="form-field">
              <label htmlFor="login-email">Email</label>
              <input
                id="login-email"
                name="email"
                type="email"
                className={`form-input ${errors.email ? "error" : ""}`}
                value={values.email}
                onChange={handleChange}
                autoComplete="email"
                aria-invalid={Boolean(errors.email)}
                aria-describedby={emailErrorId}
                required
              />
              {errors.email && (
                <p className="input-error" id="login-email-error" role="status">
                  {errors.email}
                </p>
              )}
            </div>

            <div className="form-field">
              <label htmlFor="login-password">Password</label>
              <input
                id="login-password"
                name="password"
                type="password"
                className={`form-input ${errors.password ? "error" : ""}`}
                value={values.password}
                onChange={handleChange}
                autoComplete="current-password"
                aria-invalid={Boolean(errors.password)}
                aria-describedby={passwordErrorId}
                required
              />
              {errors.password && (
                <p
                  className="input-error"
                  id="login-password-error"
                  role="status"
                >
                  {errors.password}
                </p>
              )}
            </div>

            <div className="auth-meta">
              <span>Need help? Contact support</span>
              <Link to="/signup">Create account</Link>
            </div>

            <button
              type="submit"
              className="auth-submit"
              disabled={isSubmitting}
            >
              {isSubmitting ? "Signing In…" : "Sign In"}
            </button>
          </form>

          <p className="auth-support">
            New to Omnigen? <Link to="/signup">Start building</Link>
          </p>
        </section>
      </div>
    </main>
  );
}

export default Login;
