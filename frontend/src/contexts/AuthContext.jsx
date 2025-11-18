import { useState, useEffect, useCallback, useRef } from "react";
import { useNavigate } from "react-router-dom";
import * as cognito from "../services/cognito";
import { auth as authAPI } from "../utils/api";
import { AuthContext } from "./authContext.js";

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const navigate = useNavigate();
  const tokenRefreshTimerRef = useRef(null);

  // Check if user is already authenticated on mount
  useEffect(() => {
    checkAuth();

    // Cleanup timer on unmount
    return () => {
      if (tokenRefreshTimerRef.current) {
        clearTimeout(tokenRefreshTimerRef.current);
      }
    };
  }, []);

  const checkAuth = async () => {
    try {
      setLoading(true);

      // Development mode: Skip authentication and use mock user
      if (import.meta.env.MODE === 'development' || import.meta.env.DEV) {
        setUser({
          id: 'dev-user-123',
          email: 'dev@localhost',
          subscription_tier: 'pro',
        });
        setLoading(false);
        return;
      }

      // Try to get user info from backend (checks httpOnly cookie)
      const userData = await authAPI.me();
      setUser(userData);

      // Schedule proactive token refresh after successful authentication
      scheduleTokenRefresh();
    } catch {
      // Not authenticated or session expired
      setUser(null);
    } finally {
      setLoading(false);
    }
  };

  /**
   * Login with email and password
   */
  const login = async (email, password) => {
    try {
      setError(null);
      setLoading(true);

      // Authenticate with Cognito
      const tokens = await cognito.loginUser(email, password);

      // Exchange Cognito tokens for httpOnly cookies via backend
      const userData = await authAPI.login(tokens);

      setUser(userData);

      // Schedule proactive token refresh after login
      scheduleTokenRefresh();

      return { success: true };
    } catch (err) {
      console.error("Login error:", err);

      // Handle specific Cognito errors
      if (err.code === "UserNotConfirmedException") {
        setError("Please verify your email before logging in");
        return {
          success: false,
          error: "Please verify your email before logging in",
          code: "UserNotConfirmedException",
          email,
        };
      } else if (err.code === "NotAuthorizedException") {
        setError("Invalid email or password");
        return { success: false, error: "Invalid email or password" };
      } else if (err.code === "NewPasswordRequired") {
        setError("Please set a new password");
        return {
          success: false,
          error: "Please set a new password",
          code: "NewPasswordRequired",
        };
      }

      const errorMsg = err.message || "Login failed";
      setError(errorMsg);
      return { success: false, error: errorMsg };
    } finally {
      setLoading(false);
    }
  };

  /**
   * Signup new user
   */
  const signup = async (name, email, password) => {
    try {
      setError(null);
      setLoading(true);

      const result = await cognito.signupUser(name, email, password);

      return {
        success: true,
        userConfirmed: result.userConfirmed,
        message: result.userConfirmed
          ? "Account created successfully! You can now log in."
          : "Account created! Please check your email for a verification code.",
      };
    } catch (err) {
      console.error("Signup error:", err);

      // Handle specific Cognito errors
      if (err.code === "UsernameExistsException") {
        setError("An account with this email already exists");
        return {
          success: false,
          error: "An account with this email already exists",
        };
      } else if (err.code === "InvalidPasswordException") {
        const errorMsg = "Password must contain: uppercase, lowercase, number, and special character (min 8 chars)";
        setError(errorMsg);
        return { success: false, error: errorMsg };
      }

      const errorMsg = err.message || "Signup failed";
      setError(errorMsg);
      return { success: false, error: errorMsg };
    } finally {
      setLoading(false);
    }
  };

  /**
   * Confirm signup with verification code
   */
  const confirmSignup = async (email, code) => {
    try {
      setError(null);
      setLoading(true);

      await cognito.confirmSignup(email, code);

      return {
        success: true,
        message: "Email verified successfully! You can now log in.",
      };
    } catch (err) {
      console.error("Confirmation error:", err);

      if (err.code === "CodeMismatchException") {
        setError("Invalid verification code");
        return { success: false, error: "Invalid verification code" };
      } else if (err.code === "ExpiredCodeException") {
        setError("Verification code has expired");
        return { success: false, error: "Verification code has expired" };
      }

      const errorMsg = err.message || "Verification failed";
      setError(errorMsg);
      return { success: false, error: errorMsg };
    } finally {
      setLoading(false);
    }
  };

  /**
   * Resend verification code
   */
  const resendCode = async (email) => {
    try {
      setError(null);
      await cognito.resendConfirmationCode(email);
      return { success: true, message: "Verification code sent to your email" };
    } catch (err) {
      console.error("Resend code error:", err);
      const errorMsg = err.message || "Failed to resend code";
      setError(errorMsg);
      return { success: false, error: errorMsg };
    }
  };

  /**
   * Initiate forgot password flow
   */
  const forgotPassword = async (email) => {
    try {
      setError(null);
      setLoading(true);

      await cognito.forgotPassword(email);

      return {
        success: true,
        message: "Password reset code sent to your email",
      };
    } catch (err) {
      console.error("Forgot password error:", err);

      const errorMsg = err.message || "Failed to send reset code";
      setError(errorMsg);
      return { success: false, error: errorMsg };
    } finally {
      setLoading(false);
    }
  };

  /**
   * Confirm password reset with code
   */
  const resetPassword = async (email, code, newPassword) => {
    try {
      setError(null);
      setLoading(true);

      await cognito.confirmPassword(email, code, newPassword);

      return {
        success: true,
        message: "Password reset successfully! You can now log in.",
      };
    } catch (err) {
      console.error("Reset password error:", err);

      if (err.code === "CodeMismatchException") {
        setError("Invalid reset code");
        return { success: false, error: "Invalid reset code" };
      } else if (err.code === "ExpiredCodeException") {
        setError("Reset code has expired");
        return { success: false, error: "Reset code has expired" };
      } else if (err.code === "InvalidPasswordException") {
        const errorMsg = "Password must contain: uppercase, lowercase, number, and special character (min 8 chars)";
        setError(errorMsg);
        return { success: false, error: errorMsg };
      }

      const errorMsg = err.message || "Password reset failed";
      setError(errorMsg);
      return { success: false, error: errorMsg };
    } finally {
      setLoading(false);
    }
  };

  /**
   * Logout user
   */
  const logout = async () => {
    try {
      // Clear token refresh timer
      if (tokenRefreshTimerRef.current) {
        clearTimeout(tokenRefreshTimerRef.current);
      }

      // Sign out from Cognito
      cognito.signOut();

      // Clear backend cookies
      await authAPI.logout();

      setUser(null);
      navigate("/login");
    } catch (err) {
      console.error("Logout error:", err);
      // Force logout even if backend call fails
      setUser(null);
      navigate("/login");
    }
  };

  /**
   * Proactively refresh tokens before they expire
   * Using direct implementation to avoid circular dependency
   */
  const proactiveRefresh = useCallback(async () => {
    try {
      console.log("[AUTH] 🔄 Proactive token refresh initiated");

      // Refresh Cognito session
      const newTokens = await cognito.refreshSession();

      // Update backend cookies
      await authAPI.login(newTokens);

      console.log("[AUTH] ✅ Proactive token refresh successful");

      // Schedule next refresh (50 minutes from now, tokens expire in 60 minutes)
      // Inline scheduling to avoid circular dependency
      if (tokenRefreshTimerRef.current) {
        clearTimeout(tokenRefreshTimerRef.current);
      }

      const refreshInterval = 50 * 60 * 1000; // 50 minutes in milliseconds
      tokenRefreshTimerRef.current = setTimeout(() => {
        proactiveRefresh();
      }, refreshInterval);

      console.log("[AUTH] ⏰ Next token refresh scheduled for 50 minutes from now");
    } catch (err) {
      console.error("[AUTH] ❌ Proactive token refresh failed:", err);
      // If refresh fails, user will be logged out on next API call via interceptor
    }
  }, []);

  /**
   * Schedule automatic token refresh before expiry
   * Tokens expire in 1 hour, refresh at 50 minutes
   */
  const scheduleTokenRefresh = useCallback(() => {
    // Clear existing timer
    if (tokenRefreshTimerRef.current) {
      clearTimeout(tokenRefreshTimerRef.current);
    }

    // Schedule refresh for 50 minutes (3000 seconds)
    const refreshInterval = 50 * 60 * 1000; // 50 minutes in milliseconds
    tokenRefreshTimerRef.current = setTimeout(() => {
      proactiveRefresh();
    }, refreshInterval);

    console.log("[AUTH] ⏰ Token refresh scheduled for 50 minutes from now");
  }, [proactiveRefresh]);

  /**
   * Refresh authentication (useful for checking if still authenticated)
   * Note: checkAuth() internally schedules token refresh if authenticated
   */
  const refresh = useCallback(async () => {
    await checkAuth();
  }, []);

  const value = {
    user,
    loading,
    error,
    isAuthenticated: !!user,
    login,
    signup,
    confirmSignup,
    resendCode,
    forgotPassword,
    resetPassword,
    logout,
    refresh,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
