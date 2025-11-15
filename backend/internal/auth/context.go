package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/internal/domain"
)

// Context keys for storing auth information in Gin context
const (
	UserClaimsKey = "user_claims"
	UserIDKey     = "user_id"
	UserKey       = "user"
)

// SetUserClaims stores user claims in the Gin context
func SetUserClaims(c *gin.Context, claims *domain.UserClaims) {
	c.Set(UserClaimsKey, claims)
	c.Set(UserIDKey, claims.Sub)
	c.Set(UserKey, claims.ToUser())
}

// GetUserClaims retrieves user claims from the Gin context
func GetUserClaims(c *gin.Context) (*domain.UserClaims, bool) {
	claims, exists := c.Get(UserClaimsKey)
	if !exists {
		return nil, false
	}

	userClaims, ok := claims.(*domain.UserClaims)
	return userClaims, ok
}

// MustGetUserClaims retrieves user claims or panics if not found
func MustGetUserClaims(c *gin.Context) *domain.UserClaims {
	claims, ok := GetUserClaims(c)
	if !ok {
		panic("user claims not found in context")
	}
	return claims
}

// GetUserID retrieves the user ID from the Gin context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return "", false
	}

	id, ok := userID.(string)
	return id, ok
}

// MustGetUserID retrieves the user ID or panics if not found
func MustGetUserID(c *gin.Context) string {
	userID, ok := GetUserID(c)
	if !ok {
		panic("user ID not found in context")
	}
	return userID
}

// GetUser retrieves the user model from the Gin context
func GetUser(c *gin.Context) (*domain.User, bool) {
	user, exists := c.Get(UserKey)
	if !exists {
		return nil, false
	}

	u, ok := user.(*domain.User)
	return u, ok
}

// MustGetUser retrieves the user model or panics if not found
func MustGetUser(c *gin.Context) *domain.User {
	user, ok := GetUser(c)
	if !ok {
		panic("user not found in context")
	}
	return user
}

// GetUserEmail retrieves the user's email from context
func GetUserEmail(c *gin.Context) (string, bool) {
	claims, ok := GetUserClaims(c)
	if !ok {
		return "", false
	}
	return claims.Email, true
}

// GetSubscriptionTier retrieves the user's subscription tier from context
func GetSubscriptionTier(c *gin.Context) (string, bool) {
	claims, ok := GetUserClaims(c)
	if !ok {
		return "", false
	}
	return claims.SubscriptionTier, true
}

// IsAuthenticated checks if the current request is authenticated
func IsAuthenticated(c *gin.Context) bool {
	_, exists := c.Get(UserIDKey)
	return exists
}
