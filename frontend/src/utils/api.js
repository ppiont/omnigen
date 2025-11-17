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
  const method = options.method || 'GET';
  const timestamp = new Date().toISOString();

  // Log API request
  console.log(`[API] ${timestamp} → ${method} ${endpoint}`, {
    url,
    method,
    hasBody: !!options.body,
    body: options.body ? JSON.parse(options.body) : undefined,
  });

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
    const startTime = performance.now();
    const response = await fetch(url, config);
    const duration = Math.round(performance.now() - startTime);

    // Handle successful responses
    if (response.ok) {
      // Check if response has content
      const contentType = response.headers.get('content-type');
      if (contentType && contentType.includes('application/json')) {
        const data = await response.json();
        console.log(`[API] ${new Date().toISOString()} ← ${method} ${endpoint} (${response.status}) [${duration}ms]`, data);
        return data;
      }
      console.log(`[API] ${new Date().toISOString()} ← ${method} ${endpoint} (${response.status}) [${duration}ms] (no content)`);
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
    const apiError = new APIError(
      error.message || error.Message || 'Request failed',
      response.status,
      error.code || error.Code || 'UNKNOWN_ERROR',
      error.details || error.Details || null
    );

    // Don't log 401 errors as errors (they're expected when not logged in)
    if (response.status === 401) {
      console.log(`[API] ${new Date().toISOString()} ← ${method} ${endpoint} (${response.status}) [${duration}ms] (Unauthenticated)`);
    } else {
      console.error(`[API] ${new Date().toISOString()} ✗ ${method} ${endpoint} (${response.status}) [${duration}ms]`, apiError);
    }

    throw apiError;
  } catch (error) {
    // Re-throw API errors
    if (error instanceof APIError) {
      throw error;
    }

    // Network errors or other exceptions
    const networkError = new APIError(
      error.message || 'Network error',
      0,
      'NETWORK_ERROR',
      null
    );
    console.error(`[API] ${new Date().toISOString()} ✗ ${method} ${endpoint} (NETWORK_ERROR)`, networkError);
    throw networkError;
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
   * @param {Object} params - Query parameters
   * @param {number} params.page - Page number
   * @param {number} params.page_size - Page size
   * @param {string} params.status - Filter by status
   * @returns {Promise<{jobs: Array, total_count: number, page: number, page_size: number}>}
   */
  list: (params = {}) => {
    const queryParams = new URLSearchParams();
    if (params.page) queryParams.append('page', params.page);
    if (params.page_size) queryParams.append('page_size', params.page_size);
    if (params.status) queryParams.append('status', params.status);

    const query = queryParams.toString();
    return apiRequest(`/api/v1/jobs${query ? `?${query}` : ''}`);
  },

  /**
   * Get a specific job by ID
   * @param {string} id - Job ID
   * @returns {Promise<Object>}
   */
  get: (id) => apiRequest(`/api/v1/jobs/${id}`),

  /**
   * Get detailed progress for a job
   * @param {string} id - Job ID
   * @returns {Promise<{job_id: string, status: string, progress: number, current_stage: string, estimated_time_remaining: number, stages_completed: Array, stages_pending: Array}>}
   */
  progress: (id) => apiRequest(`/api/v1/jobs/${id}/progress`),

  /**
   * Delete a job
   * @param {string} id - Job ID
   * @returns {Promise<void>}
   */
  delete: (id) =>
    apiRequest(`/api/v1/jobs/${id}`, {
      method: 'DELETE',
    }),
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
   * @param {string} params.title - Optional video title
   * @returns {Promise<{job_id: string, status: string}>}
   */
  create: (params) =>
    apiRequest('/api/v1/generate', {
      method: 'POST',
      body: JSON.stringify(params),
    }),

  /**
   * Generate a catchy video title from prompt
   * @param {Object} params - Title generation parameters
   * @param {string} params.prompt - Video description
   * @returns {Promise<{title: string}>}
   */
  title: (params) =>
    apiRequest('/api/v1/generate-title', {
      method: 'POST',
      body: JSON.stringify(params),
    }),
};

/**
 * Scripts/Parser API endpoints
 */
export const scripts = {
  /**
   * Generate a script from user input
   * @param {Object} params - Script generation parameters
   * @param {string} params.prompt - Video description
   * @param {number} params.duration - Duration in seconds (15, 30, or 60)
   * @param {string} params.product_name - Product name
   * @param {string} params.target_audience - Target audience
   * @param {string} params.brand_vibe - Optional brand vibe/style
   * @returns {Promise<{script_id: string, status: string, message: string}>}
   */
  parse: (params) =>
    apiRequest('/api/v1/parse', {
      method: 'POST',
      body: JSON.stringify(params),
    }),

  /**
   * Get a script by ID
   * @param {string} id - Script ID
   * @returns {Promise<Object>}
   */
  get: (id) => apiRequest(`/api/v1/scripts/${id}`),
};

/**
 * Presets API endpoints
 */
export const presets = {
  /**
   * Get all brand style presets
   * @returns {Promise<{presets: Array}>}
   */
  list: () => apiRequest('/api/v1/presets'),
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
