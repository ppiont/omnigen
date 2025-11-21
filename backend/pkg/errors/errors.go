package errors

import "net/http"

// APIError represents a standardized API error response
type APIError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	Status  int                    `json:"-"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return e.Message
}

// WithDetails adds details to an error
func (e *APIError) WithDetails(details map[string]interface{}) *APIError {
	newErr := *e
	newErr.Details = details
	return &newErr
}

// Common error definitions
var (
	// Validation errors (400)
	ErrInvalidRequest = &APIError{
		Code:    "INVALID_REQUEST",
		Message: "Invalid request body",
		Status:  http.StatusBadRequest,
	}

	ErrInvalidDuration = &APIError{
		Code:    "INVALID_DURATION",
		Message: "Duration must be between 15 and 180 seconds",
		Status:  http.StatusBadRequest,
	}

	ErrInvalidPrompt = &APIError{
		Code:    "INVALID_PROMPT",
		Message: "Prompt must be between 10 and 1000 characters",
		Status:  http.StatusBadRequest,
	}

	ErrInvalidAspectRatio = &APIError{
		Code:    "INVALID_ASPECT_RATIO",
		Message: "Aspect ratio must be one of: 16:9, 9:16, 1:1",
		Status:  http.StatusBadRequest,
	}

	// Authentication errors (401)
	ErrUnauthorized = &APIError{
		Code:    "UNAUTHORIZED",
		Message: "Authentication required",
		Status:  http.StatusUnauthorized,
	}

	ErrMissingAPIKey = &APIError{
		Code:    "MISSING_API_KEY",
		Message: "API key is required",
		Status:  http.StatusUnauthorized,
	}

	ErrInvalidAPIKey = &APIError{
		Code:    "INVALID_API_KEY",
		Message: "Invalid API key",
		Status:  http.StatusUnauthorized,
	}

	// Authorization errors (403)
	ErrForbidden = &APIError{
		Code:    "FORBIDDEN",
		Message: "You do not have permission to access this resource",
		Status:  http.StatusForbidden,
	}

	// Not found errors (404)
	ErrJobNotFound = &APIError{
		Code:    "JOB_NOT_FOUND",
		Message: "Job not found",
		Status:  http.StatusNotFound,
	}

	ErrNotFound = &APIError{
		Code:    "NOT_FOUND",
		Message: "Resource not found",
		Status:  http.StatusNotFound,
	}

	// Not implemented (501)
	ErrNotImplemented = &APIError{
		Code:    "NOT_IMPLEMENTED",
		Message: "Feature not yet implemented",
		Status:  http.StatusNotImplemented,
	}

	// Server errors (500)
	ErrInternalServer = &APIError{
		Code:    "INTERNAL_SERVER_ERROR",
		Message: "An internal server error occurred",
		Status:  http.StatusInternalServerError,
	}

	ErrDatabaseError = &APIError{
		Code:    "DATABASE_ERROR",
		Message: "Database operation failed",
		Status:  http.StatusInternalServerError,
	}

	ErrStorageError = &APIError{
		Code:    "STORAGE_ERROR",
		Message: "Storage operation failed",
		Status:  http.StatusInternalServerError,
	}

	ErrWorkflowError = &APIError{
		Code:    "WORKFLOW_ERROR",
		Message: "Failed to start workflow execution",
		Status:  http.StatusInternalServerError,
	}

	ErrPromptParsingFailed = &APIError{
		Code:    "PROMPT_PARSING_FAILED",
		Message: "Failed to parse prompt",
		Status:  http.StatusInternalServerError,
	}

	// Service unavailable (503)
	ErrServiceUnavailable = &APIError{
		Code:    "SERVICE_UNAVAILABLE",
		Message: "Service temporarily unavailable",
		Status:  http.StatusServiceUnavailable,
	}
)

// ErrorResponse is the JSON response for errors
type ErrorResponse struct {
	Error *APIError `json:"error"`
}

// NewAPIError creates a new API error
func NewAPIError(base *APIError, message string, details map[string]interface{}) *APIError {
	err := *base
	if message != "" {
		err.Message = message
	}
	if details != nil {
		err.Details = details
	}
	return &err
}

// NewValidationError creates a field-specific validation error derived from ErrInvalidRequest.
func NewValidationError(field, message string) *APIError {
	return NewAPIError(ErrInvalidRequest, message, map[string]interface{}{
		"field": field,
	})
}

// NewServiceError creates a sanitized internal server error for downstream service failures.
func NewServiceError(service, message string) *APIError {
	return NewAPIError(ErrInternalServer, message, map[string]interface{}{
		"service": service,
	})
}
