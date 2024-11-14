package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type RefreshJWTPayload struct {
	UserID string `json:"user_id"`
	Type   string `json:"type" default:"refresh"`
}
type RefreshJWTClaims struct {
	RefreshJWTPayload
	jwt.RegisteredClaims
}

// GenerateRefreshToken generates a JWT token with the given payload, and returns the accessToken and refreshToken
func GenerateRefreshToken(secret string, payload RefreshJWTPayload, days time.Duration) string {
	refreshClaims := RefreshJWTClaims{
		RefreshJWTPayload: payload,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(days)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(secret))
	if err != nil {
		return ""
	}
	return refreshTokenString
}

// ParseRefreshToken parses a JWT token and returns the payload
func ParseRefreshToken(secret string, refreshTokenString string) (*RefreshJWTPayload, bool) {
	token, err := jwt.ParseWithClaims(refreshTokenString, &RefreshJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, false
	}
	if claims, ok := token.Claims.(*RefreshJWTClaims); ok && token.Valid {
		return &claims.RefreshJWTPayload, true
	}
	return nil, false
}
