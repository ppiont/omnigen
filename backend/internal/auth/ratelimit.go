package auth

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/pkg/errors"
	"go.uber.org/zap"
)

// Rate limits per subscription tier (requests per minute)
var tierRateLimits = map[string]int{
	"free":       10,  // 10 requests per minute
	"pro":        60,  // 60 requests per minute (1 per second)
	"enterprise": 300, // 300 requests per minute (5 per second)
}

// RateLimiter tracks request rates per user
type RateLimiter struct {
	requests map[string]*userRateLimit
	mu       sync.RWMutex
	logger   *zap.Logger
	window   time.Duration
}

// userRateLimit tracks rate limit state for a single user
type userRateLimit struct {
	count      int
	resetAt    time.Time
	tier       string
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter with specified window duration
func NewRateLimiter(window time.Duration, logger *zap.Logger) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*userRateLimit),
		logger:   logger,
		window:   window,
	}

	// Start cleanup goroutine to remove old entries
	go rl.cleanup()

	return rl
}

// cleanup periodically removes expired rate limit entries
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for userID, limit := range rl.requests {
			limit.mu.Lock()
			if now.After(limit.resetAt.Add(time.Minute)) {
				delete(rl.requests, userID)
			}
			limit.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// isAllowed checks if a request from the user is allowed based on their tier
func (rl *RateLimiter) isAllowed(userID, tier string) (bool, int, time.Duration) {
	rl.mu.Lock()
	userLimit, exists := rl.requests[userID]
	if !exists {
		userLimit = &userRateLimit{
			count:   0,
			resetAt: time.Now().Add(rl.window),
			tier:    tier,
		}
		rl.requests[userID] = userLimit
	}
	rl.mu.Unlock()

	userLimit.mu.Lock()
	defer userLimit.mu.Unlock()

	now := time.Now()

	// Reset window if expired
	if now.After(userLimit.resetAt) {
		userLimit.count = 0
		userLimit.resetAt = now.Add(rl.window)
		userLimit.tier = tier
	}

	// Get rate limit for tier
	limit := tierRateLimits[tier]
	if limit == 0 {
		limit = tierRateLimits["free"] // Default to free tier
	}

	// Check if under limit
	if userLimit.count >= limit {
		resetIn := time.Until(userLimit.resetAt)
		return false, limit, resetIn
	}

	// Increment counter
	userLimit.count++
	remaining := limit - userLimit.count
	resetIn := time.Until(userLimit.resetAt)

	return true, remaining, resetIn
}

// RateLimitMiddleware creates a Gin middleware for rate limiting
func RateLimitMiddleware(rateLimiter *RateLimiter, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user claims from context (set by JWT middleware)
		claims, ok := GetUserClaims(c)
		if !ok {
			// If no user claims, this endpoint should require auth
			// Let it pass to the auth middleware to handle
			c.Next()
			return
		}

		userID := claims.Sub
		tier := claims.SubscriptionTier
		if tier == "" {
			tier = "free"
		}

		allowed, remaining, resetIn := rateLimiter.isAllowed(userID, tier)

		// Set rate limit headers
		limit := tierRateLimits[tier]
		if limit == 0 {
			limit = tierRateLimits["free"]
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(resetIn).Unix()))

		if !allowed {
			logger.Warn("Rate limit exceeded",
				zap.String("user_id", userID),
				zap.String("tier", tier),
				zap.Int("limit", limit),
			)

			c.JSON(http.StatusTooManyRequests, errors.NewAPIError(
				&errors.APIError{
					Code:    "RATE_LIMIT_EXCEEDED",
					Message: "Rate limit exceeded",
					Status:  http.StatusTooManyRequests,
				},
				fmt.Sprintf("Rate limit of %d requests per minute exceeded", limit),
				map[string]interface{}{
					"limit":     limit,
					"reset_in":  resetIn.Seconds(),
					"tier":      tier,
					"upgrade":   "Upgrade your subscription for higher rate limits",
				},
			))
			c.Abort()
			return
		}

		c.Next()
	}
}
