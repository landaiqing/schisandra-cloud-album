package jwt

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zeromicro/go-zero/core/logc"
)

type AccessJWTPayload struct {
	UserID *string `json:"user_id"`
	Type   *string `json:"type" default:"access"`
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
		logc.Error(context.Background(), err)
		return ""
	}
	return accessToken
}

// ParseAccessToken parses a JWT token and returns the payload
func ParseAccessToken(secret string, tokenString string) *AccessJWTPayload {
	token, err := jwt.ParseWithClaims(tokenString, &AccessJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		logc.Error(context.Background(), err)
		return nil
	}
	if claims, ok := token.Claims.(*AccessJWTClaims); ok && token.Valid {
		return &claims.AccessJWTPayload
	}
	return nil
}
