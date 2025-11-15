package domain

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// User represents an authenticated user in the system
type User struct {
	ID               string    `json:"id"`                // Cognito sub claim
	Email            string    `json:"email"`             // User email
	EmailVerified    bool      `json:"email_verified"`    // Email verification status
	Username         string    `json:"username"`          // Cognito username
	SubscriptionTier string    `json:"subscription_tier"` // Subscription level (free, pro, enterprise)
	CreatedAt        time.Time `json:"created_at"`        // Account creation time
	LastLogin        time.Time `json:"last_login"`        // Last login timestamp
}

// UserClaims represents the JWT claims extracted from a Cognito token
type UserClaims struct {
	jwt.RegisteredClaims
	Sub              string `json:"sub"`                          // User ID (Cognito sub)
	Email            string `json:"email"`                        // User email
	EmailVerified    bool   `json:"email_verified"`               // Email verification status
	CognitoUsername  string `json:"cognito:username"`             // Cognito username
	SubscriptionTier string `json:"custom:subscription_tier"`     // Custom attribute
	TokenUse         string `json:"token_use"`                    // "access" or "id"
	AuthTime         int64  `json:"auth_time"`                    // Authentication timestamp
}

// ToUser converts UserClaims to a User model
func (uc *UserClaims) ToUser() *User {
	return &User{
		ID:               uc.Sub,
		Email:            uc.Email,
		EmailVerified:    uc.EmailVerified,
		Username:         uc.CognitoUsername,
		SubscriptionTier: uc.SubscriptionTier,
		LastLogin:        time.Unix(uc.AuthTime, 0),
	}
}

// IsAccessToken checks if the token is an access token
func (uc *UserClaims) IsAccessToken() bool {
	return uc.TokenUse == "access"
}

// IsIDToken checks if the token is an ID token
func (uc *UserClaims) IsIDToken() bool {
	return uc.TokenUse == "id"
}

// IsExpired checks if the token is expired
func (uc *UserClaims) IsExpired() bool {
	if uc.ExpiresAt == nil {
		return false
	}
	return uc.ExpiresAt.Before(time.Now())
}

// Usage represents API usage tracking for a user
type Usage struct {
	UserID           string    `json:"user_id" dynamodbav:"user_id"`
	Period           string    `json:"period" dynamodbav:"period"` // Format: YYYY-MM
	RequestCount     int       `json:"request_count" dynamodbav:"request_count"`
	VideoGenerated   int       `json:"video_generated" dynamodbav:"video_generated"`
	TotalDuration    int       `json:"total_duration" dynamodbav:"total_duration"` // Seconds
	LastUpdated      time.Time `json:"last_updated" dynamodbav:"last_updated"`
	MonthlyQuota     int       `json:"monthly_quota" dynamodbav:"monthly_quota"`
	QuotaRemaining   int       `json:"quota_remaining" dynamodbav:"quota_remaining"`
}

// HasQuotaRemaining checks if user has remaining quota
func (u *Usage) HasQuotaRemaining() bool {
	return u.QuotaRemaining > 0
}

// IncrementUsage increments the usage counters
func (u *Usage) IncrementUsage(videoDuration int) {
	u.RequestCount++
	u.VideoGenerated++
	u.TotalDuration += videoDuration
	u.QuotaRemaining--
	u.LastUpdated = time.Now()
}
