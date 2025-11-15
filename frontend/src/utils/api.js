/**
 * API client utility for making authenticated requests to the backend
 * Automatically includes credentials (httpOnly cookies) with every request
 */

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

/**
 * Custom error class for API errors
 */
export class APIError extends Error {
  constructor(message, status, code, details) {
    super(message);
    this.name = 'APIError';
    this.status = status;
    this.code = code;
    this.details = details;
  }
}

/**
 * Make an authenticated API request
 * @param {string} endpoint - API endpoint (e.g., '/api/v1/jobs')
 * @param {Object} options - Fetch options
 * @returns {Promise<any>} Response data
 */
export async function apiRequest(endpoint, options = {}) {
  const url = `${API_BASE_URL}${endpoint}`;

  const defaultOptions = {
    credentials: 'include', // Include httpOnly cookies
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  };

  const config = {
    ...defaultOptions,
    ...options,
    headers: {
      ...defaultOptions.headers,
      ...options.headers,
    },
  };

  try {
    const response = await fetch(url, config);

    // Handle successful responses
    if (response.ok) {
      // Check if response has content
      const contentType = response.headers.get('content-type');
      if (contentType && contentType.includes('application/json')) {
        return await response.json();
      }
      return null;
    }

    // Handle error responses
    let errorData;
    try {
      errorData = await response.json();
    } catch (e) {
      // Response is not JSON
      throw new APIError(
        response.statusText || 'Request failed',
        response.status,
        'UNKNOWN_ERROR',
        null
      );
    }

    // Extract error details from response
    const error = errorData.error || errorData;
    throw new APIError(
      error.message || error.Message || 'Request failed',
      response.status,
      error.code || error.Code || 'UNKNOWN_ERROR',
      error.details || error.Details || null
    );
  } catch (error) {
    // Re-throw API errors
    if (error instanceof APIError) {
      throw error;
    }

    // Network errors or other exceptions
    throw new APIError(
      error.message || 'Network error',
      0,
      'NETWORK_ERROR',
      null
    );
  }
}

/**
 * Auth API endpoints
 */
export const auth = {
  /**
   * Exchange Cognito tokens for httpOnly cookies
   * @param {Object} tokens - Cognito tokens
   * @param {string} tokens.accessToken - Access token
   * @param {string} tokens.idToken - ID token
   * @param {string} tokens.refreshToken - Refresh token
   * @returns {Promise<{user_id: string, email: string, subscription_tier: string}>}
   */
  login: (tokens) =>
    apiRequest('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify({
        access_token: tokens.accessToken,
        id_token: tokens.idToken,
        refresh_token: tokens.refreshToken,
      }),
    }),

  /**
   * Logout and clear cookies
   * @returns {Promise<void>}
   */
  logout: () =>
    apiRequest('/api/v1/auth/logout', {
      method: 'POST',
    }),

  /**
   * Get current user info
   * @returns {Promise<{user_id: string, email: string, subscription_tier: string}>}
   */
  me: () => apiRequest('/api/v1/auth/me'),

  /**
   * Refresh tokens
   * @returns {Promise<any>}
   */
  refresh: () =>
    apiRequest('/api/v1/auth/refresh', {
      method: 'POST',
    }),
};

/**
 * Jobs API endpoints
 */
export const jobs = {
  /**
   * Get all jobs for the current user
   * @returns {Promise<Array>}
   */
  list: () => apiRequest('/api/v1/jobs'),

  /**
   * Get a specific job by ID
   * @param {string} id - Job ID
   * @returns {Promise<Object>}
   */
  get: (id) => apiRequest(`/api/v1/jobs/${id}`),
};

/**
 * Generation API endpoints
 */
export const generate = {
  /**
   * Submit a new video generation job
   * @param {Object} params - Generation parameters
   * @param {string} params.prompt - Video description
   * @param {number} params.duration - Duration in seconds
   * @param {string} params.aspect_ratio - Aspect ratio (16:9, 9:16, 1:1)
   * @param {string} params.style - Visual style
   * @returns {Promise<{job_id: string, status: string}>}
   */
  create: (params) =>
    apiRequest('/api/v1/generate', {
      method: 'POST',
      body: JSON.stringify(params),
    }),
};

/**
 * Health check endpoint (no auth required)
 */
export const health = {
  check: () =>
    apiRequest('/health', {
      credentials: 'omit', // No cookies needed for health check
    }),
};
