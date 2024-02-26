package jwt

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
)

const (
	secretLength = 256
)

var (
	secret              []byte
	ErrInvaildToken     = fmt.Errorf("invalid token")
	ErrNoTokenInRequest = fmt.Errorf("no token in request")
)

func init() {
	secret, _ = utils.GetRandBytes(secretLength)
}

type Claims struct {
	jwt.RegisteredClaims
	SessionId string `json:"sessionId"`
}

func GenerateToken(sessionId string) (string, error) {
	claims := Claims{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			Issuer:    "test",
		},
		sessionId,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secret)
}

func parse(tokenStr string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return "", err
	}

	if token.Valid {
		return claims.SessionId, nil
	} else if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
		return "", fmt.Errorf("token expired")
	} else {
		return "", ErrInvaildToken
	}
}

// GetSessionIdFromToken extracts sessionId from http.Request Auth header
func GetSessionIdFromToken(request *http.Request) (sessionId string, err error) {
	authHeader := request.Header.Get("Authorization")
	parts := strings.Split(authHeader, "Bearer")
	if len(parts) != 2 {
		return "", ErrNoTokenInRequest
	}

	tokenStr := strings.TrimSpace(parts[1])
	if len(tokenStr) < 1 {
		return "", ErrNoTokenInRequest
	}
	if tokenStr == "" {
		return "", ErrNoTokenInRequest
	}
	return parse(tokenStr)
}
