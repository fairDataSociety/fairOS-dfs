package auth

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/auth/cookie"
	"github.com/fairdatasociety/fairOS-dfs/pkg/auth/jwt"
)

// GetUniqueSessionId generates a sessionId for each logged-in user
func GetUniqueSessionId() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func GetSessionIdFromRequest(r *http.Request) (string, error) {
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		sessionId, err = jwt.GetSessionIdFromToken(r)
		if err != nil {
			return "", err
		}
	}

	return sessionId, err
}
