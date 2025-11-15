package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	AccessTokenCookie  = "access_token"
	IDTokenCookie      = "id_token"
	RefreshTokenCookie = "refresh_token"

	// Token expiry durations
	AccessTokenExpiry  = 1 * time.Hour
	IDTokenExpiry      = 1 * time.Hour
	RefreshTokenExpiry = 30 * 24 * time.Hour // 30 days
)

// CookieConfig holds configuration for cookie settings
type CookieConfig struct {
	Secure   bool   // HTTPS only
	Domain   string // Cookie domain
	SameSite http.SameSite
}

// SetAuthCookies sets httpOnly cookies for all auth tokens
func SetAuthCookies(c *gin.Context, accessToken, idToken, refreshToken string, config CookieConfig) {
	// Set access token cookie (1 hour expiry)
	c.SetCookie(
		AccessTokenCookie,
		accessToken,
		int(AccessTokenExpiry.Seconds()),
		"/",
		config.Domain,
		config.Secure,
		true, // httpOnly
	)
	c.SetSameSite(config.SameSite)

	// Set ID token cookie (1 hour expiry)
	c.SetCookie(
		IDTokenCookie,
		idToken,
		int(IDTokenExpiry.Seconds()),
		"/",
		config.Domain,
		config.Secure,
		true, // httpOnly
	)

	// Set refresh token cookie (30 days expiry)
	c.SetCookie(
		RefreshTokenCookie,
		refreshToken,
		int(RefreshTokenExpiry.Seconds()),
		"/",
		config.Domain,
		config.Secure,
		true, // httpOnly
	)
}

// ClearAuthCookies removes all auth cookies
func ClearAuthCookies(c *gin.Context, config CookieConfig) {
	c.SetCookie(
		AccessTokenCookie,
		"",
		-1, // Expire immediately
		"/",
		config.Domain,
		config.Secure,
		true,
	)

	c.SetCookie(
		IDTokenCookie,
		"",
		-1,
		"/",
		config.Domain,
		config.Secure,
		true,
	)

	c.SetCookie(
		RefreshTokenCookie,
		"",
		-1,
		"/",
		config.Domain,
		config.Secure,
		true,
	)
}

// GetTokenFromCookie retrieves a token from cookies, falls back to Authorization header
func GetTokenFromCookie(c *gin.Context) string {
	// First, try to get ID token from cookie
	idToken, err := c.Cookie(IDTokenCookie)
	if err == nil && idToken != "" {
		return idToken
	}

	// Fall back to Authorization header (for backwards compatibility)
	authHeader := c.GetHeader("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}

	return ""
}

// GetRefreshTokenFromCookie retrieves the refresh token from cookies
func GetRefreshTokenFromCookie(c *gin.Context) string {
	refreshToken, err := c.Cookie(RefreshTokenCookie)
	if err != nil {
		return ""
	}
	return refreshToken
}
