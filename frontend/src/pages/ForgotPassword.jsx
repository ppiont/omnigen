import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import '../styles/Auth.css';

export default function ForgotPassword() {
  const navigate = useNavigate();
  const { forgotPassword, resetPassword } = useAuth();

  const [step, setStep] = useState('email'); // 'email' or 'reset'
  const [email, setEmail] = useState('');
  const [code, setCode] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');

  const [errors, setErrors] = useState({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [successMessage, setSuccessMessage] = useState('');

  const validateEmail = () => {
    const newErrors = {};

    if (!email) {
      newErrors.email = 'Email is required';
    } else if (!/\S+@\S+\.\S+/.test(email)) {
      newErrors.email = 'Please enter a valid email';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const validateReset = () => {
    const newErrors = {};

    if (!code) {
      newErrors.code = 'Verification code is required';
    }

    if (!newPassword) {
      newErrors.newPassword = 'New password is required';
    } else if (newPassword.length < 8) {
      newErrors.newPassword = 'Password must be at least 8 characters';
    }

    if (!confirmPassword) {
      newErrors.confirmPassword = 'Please confirm your password';
    } else if (newPassword !== confirmPassword) {
      newErrors.confirmPassword = 'Passwords do not match';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleRequestCode = async (e) => {
    e.preventDefault();

    if (!validateEmail()) return;

    setIsSubmitting(true);
    setErrors({});
    setSuccessMessage('');

    const result = await forgotPassword(email);

    if (result.success) {
      setSuccessMessage(result.message);
      setStep('reset');
    } else {
      setErrors({ form: result.error });
    }

    setIsSubmitting(false);
  };

  const handleResetPassword = async (e) => {
    e.preventDefault();

    if (!validateReset()) return;

    setIsSubmitting(true);
    setErrors({});
    setSuccessMessage('');

    const result = await resetPassword(email, code, newPassword);

    if (result.success) {
      setSuccessMessage(result.message);
      // Redirect to login after 2 seconds
      setTimeout(() => {
        navigate('/login');
      }, 2000);
    } else {
      setErrors({ form: result.error });
    }

    setIsSubmitting(false);
  };

  return (
    <div className="auth-container">
      <div className="auth-card">
        <div className="auth-header">
          <h1>Reset Password</h1>
          <p className="auth-subtitle">
            {step === 'email'
              ? 'Enter your email to receive a reset code'
              : 'Enter the code sent to your email'}
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

        {step === 'email' ? (
          <form onSubmit={handleRequestCode} className="auth-form" noValidate>
            <div className="form-group">
              <label htmlFor="email">Email Address</label>
              <input
                type="email"
                id="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className={errors.email ? 'error' : ''}
                placeholder="you@example.com"
                aria-invalid={!!errors.email}
                aria-describedby={errors.email ? 'email-error' : undefined}
                disabled={isSubmitting}
              />
              {errors.email && (
                <span className="error-message" id="email-error" role="alert">
                  {errors.email}
                </span>
              )}
            </div>

            <button
              type="submit"
              className="auth-submit"
              disabled={isSubmitting}
            >
              {isSubmitting ? 'Sending...' : 'Send Reset Code'}
            </button>
          </form>
        ) : (
          <form onSubmit={handleResetPassword} className="auth-form" noValidate>
            <div className="form-group">
              <label htmlFor="code">Verification Code</label>
              <input
                type="text"
                id="code"
                value={code}
                onChange={(e) => setCode(e.target.value)}
                className={errors.code ? 'error' : ''}
                placeholder="Enter 6-digit code"
                aria-invalid={!!errors.code}
                aria-describedby={errors.code ? 'code-error' : undefined}
                disabled={isSubmitting}
              />
              {errors.code && (
                <span className="error-message" id="code-error" role="alert">
                  {errors.code}
                </span>
              )}
            </div>

            <div className="form-group">
              <label htmlFor="newPassword">New Password</label>
              <input
                type="password"
                id="newPassword"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                className={errors.newPassword ? 'error' : ''}
                placeholder="Minimum 8 characters"
                aria-invalid={!!errors.newPassword}
                aria-describedby={errors.newPassword ? 'newPassword-error' : undefined}
                disabled={isSubmitting}
              />
              {errors.newPassword && (
                <span className="error-message" id="newPassword-error" role="alert">
                  {errors.newPassword}
                </span>
              )}
            </div>

            <div className="form-group">
              <label htmlFor="confirmPassword">Confirm New Password</label>
              <input
                type="password"
                id="confirmPassword"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                className={errors.confirmPassword ? 'error' : ''}
                placeholder="Re-enter new password"
                aria-invalid={!!errors.confirmPassword}
                aria-describedby={errors.confirmPassword ? 'confirmPassword-error' : undefined}
                disabled={isSubmitting}
              />
              {errors.confirmPassword && (
                <span className="error-message" id="confirmPassword-error" role="alert">
                  {errors.confirmPassword}
                </span>
              )}
            </div>

            <button
              type="submit"
              className="auth-submit"
              disabled={isSubmitting}
            >
              {isSubmitting ? 'Resetting...' : 'Reset Password'}
            </button>

            <button
              type="button"
              className="auth-link-button"
              onClick={() => setStep('email')}
              disabled={isSubmitting}
            >
              ← Back to email entry
            </button>
          </form>
        )}

        <div className="auth-footer">
          <p>
            Remember your password?{' '}
            <Link to="/login" className="auth-link">
              Sign in
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
