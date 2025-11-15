package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/omnigen/backend/internal/domain"
	"go.uber.org/zap"
)

// JWK represents a JSON Web Key
type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS represents a JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWTValidator handles JWT token validation
type JWTValidator struct {
	jwksURL      string
	issuer       string
	clientID     string
	logger       *zap.Logger
	keys         map[string]*rsa.PublicKey
	keysMu       sync.RWMutex
	lastFetchTime time.Time
}

// NewJWTValidator creates a new JWT validator
func NewJWTValidator(jwksURL, issuer, clientID string, logger *zap.Logger) *JWTValidator {
	return &JWTValidator{
		jwksURL:  jwksURL,
		issuer:   issuer,
		clientID: clientID,
		logger:   logger,
		keys:     make(map[string]*rsa.PublicKey),
	}
}

// FetchJWKS fetches the JWKS from Cognito
func (v *JWTValidator) FetchJWKS() error {
	v.logger.Info("Fetching JWKS", zap.String("url", v.jwksURL))

	resp, err := http.Get(v.jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	v.keysMu.Lock()
	defer v.keysMu.Unlock()

	// Convert JWKs to RSA public keys
	for _, key := range jwks.Keys {
		if key.Kty != "RSA" {
			continue
		}

		pubKey, err := v.jwkToRSAPublicKey(key)
		if err != nil {
			v.logger.Warn("Failed to convert JWK to RSA public key",
				zap.String("kid", key.Kid),
				zap.Error(err))
			continue
		}

		v.keys[key.Kid] = pubKey
	}

	v.lastFetchTime = time.Now()
	v.logger.Info("JWKS fetched successfully", zap.Int("key_count", len(v.keys)))

	return nil
}

// jwkToRSAPublicKey converts a JWK to an RSA public key
func (v *JWTValidator) jwkToRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	// Decode the modulus
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	// Decode the exponent
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert to big integers
	n := new(big.Int).SetBytes(nBytes)
	var e int
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	return &rsa.PublicKey{
		N: n,
		E: e,
	}, nil
}

// getPublicKey retrieves the public key for the given kid
func (v *JWTValidator) getPublicKey(kid string) (*rsa.PublicKey, error) {
	v.keysMu.RLock()
	key, exists := v.keys[kid]
	v.keysMu.RUnlock()

	if exists {
		return key, nil
	}

	// Key not found, try refreshing JWKS if it's been a while
	if time.Since(v.lastFetchTime) > 5*time.Minute {
		if err := v.FetchJWKS(); err != nil {
			return nil, fmt.Errorf("failed to refresh JWKS: %w", err)
		}

		v.keysMu.RLock()
		key, exists = v.keys[kid]
		v.keysMu.RUnlock()

		if exists {
			return key, nil
		}
	}

	return nil, fmt.Errorf("public key not found for kid: %s", kid)
}

// ValidateToken validates a JWT token and extracts claims
func (v *JWTValidator) ValidateToken(tokenString string) (*domain.UserClaims, error) {
	// Parse token without validation first to get the kid
	token, err := jwt.ParseWithClaims(tokenString, &domain.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing algorithm
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get kid from header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid not found in token header")
		}

		// Get public key for this kid
		return v.getPublicKey(kid)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Extract claims
	claims, ok := token.Claims.(*domain.UserClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Verify issuer
	issuer, err := claims.GetIssuer()
	if err != nil || issuer != v.issuer {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", v.issuer, issuer)
	}

	// Verify audience (client ID)
	audience, err := claims.GetAudience()
	if err != nil || len(audience) == 0 || audience[0] != v.clientID {
		aud := ""
		if len(audience) > 0 {
			aud = audience[0]
		}
		return nil, fmt.Errorf("invalid audience: expected %s, got %s", v.clientID, aud)
	}

	// Verify expiration
	if claims.IsExpired() {
		return nil, fmt.Errorf("token has expired")
	}

	// Verify token use (should be access or id token)
	if !claims.IsAccessToken() && !claims.IsIDToken() {
		return nil, fmt.Errorf("invalid token use: %s", claims.TokenUse)
	}

	return claims, nil
}

// ValidateAccessToken validates an access token specifically
func (v *JWTValidator) ValidateAccessToken(tokenString string) (*domain.UserClaims, error) {
	claims, err := v.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if !claims.IsAccessToken() {
		return nil, fmt.Errorf("token is not an access token")
	}

	return claims, nil
}

// ValidateIDToken validates an ID token specifically
func (v *JWTValidator) ValidateIDToken(tokenString string) (*domain.UserClaims, error) {
	claims, err := v.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if !claims.IsIDToken() {
		return nil, fmt.Errorf("token is not an ID token")
	}

	return claims, nil
}
