package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/ArtemShalinFe/gophkeeper/internal/models"
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	Login string
}

func NewJWTToken(secretKey []byte, login string, tokenExp time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		Login: login,
	})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", fmt.Errorf("an occured error while retrieving token signed string, err: %w", err)
	}

	return tokenString, nil
}

type key int

var userKey key

func userFromContext(ctx context.Context) (*models.User, bool) {
	u, ok := ctx.Value(userKey).(*models.User)
	return u, ok
}
