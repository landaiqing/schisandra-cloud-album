package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AccessJWTPayload struct {
	UserID string `json:"user_id"`
	Type   string `json:"type" default:"access"`
}
type AccessJWTClaims struct {
	AccessJWTPayload
	jwt.RegisteredClaims
}

func GenerateAccessToken(secret string, payload AccessJWTPayload) string {
	claims := AccessJWTClaims{
		AccessJWTPayload: payload,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return ""
	}
	return accessToken
}

// ParseAccessToken parses a JWT token and returns the payload
func ParseAccessToken(secret string, tokenString string) (*AccessJWTPayload, bool) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, false
	}
	if claims, ok := token.Claims.(*AccessJWTClaims); ok && token.Valid {
		return &claims.AccessJWTPayload, true
	}
	return nil, false
}
