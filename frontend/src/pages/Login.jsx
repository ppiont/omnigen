import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";
import "../styles/auth.css";

function Login() {
  const navigate = useNavigate();
  const { login } = useAuth();
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

  const handleSubmit = async (event) => {
    event.preventDefault();
    const nextErrors = validate();
    setErrors(nextErrors);

    if (Object.keys(nextErrors).length > 0) {
      return;
    }

    setIsSubmitting(true);

    try {
      const result = await login(values.email, values.password);

      if (result.success) {
        // Redirect to dashboard on successful login
        navigate('/dashboard');
      } else {
        // Handle specific error codes
        if (result.code === 'UserNotConfirmedException') {
          // Redirect to verification page with email
          navigate(`/verify?email=${encodeURIComponent(values.email)}`);
        } else {
          setErrors({ form: result.error });
        }
      }
    } catch (error) {
      console.error('Login error:', error);
      setErrors({ form: 'An unexpected error occurred. Please try again.' });
    } finally {
      setIsSubmitting(false);
    }
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
            {errors.form && (
              <div className="auth-error" role="alert" aria-live="polite">
                {errors.form}
              </div>
            )}

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
              <Link to="/forgot-password">Forgot password?</Link>
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
