package jwt

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logc"
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
		logc.Error(context.Background(), err)
		return ""
	}
	return refreshTokenString
}

// ParseRefreshToken parses a JWT token and returns the payload
func ParseRefreshToken(secret string, refreshTokenString string) *RefreshJWTPayload {
	token, err := jwt.ParseWithClaims(refreshTokenString, &RefreshJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		logc.Error(context.Background(), err)
		return nil
	}
	if claims, ok := token.Claims.(*RefreshJWTClaims); ok && token.Valid {
		return &claims.RefreshJWTPayload
	}
	return nil
}
