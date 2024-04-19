/*
Copyright © 2020 FairOS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/auth/jwt"

	"github.com/fairdatasociety/fairOS-dfs/pkg/auth/cookie"
	"resenje.org/jsonhttp"
)

// LoginMiddleware is a middleware that gets called before executing any of the protected handlers.
// this check if a there is a valid session id the request and then allows the request handler to
// proceed for execution.
func (h *Handler) LoginMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		sessionId, err := jwt.GetSessionIdFromRequest(r)
		if err != nil && !errors.Is(err, jwt.ErrNoTokenInRequest) {
			h.logger.Errorf("jwt: invalid token: %v", err)
			jsonhttp.Unauthorized(w, &response{Message: "jwt: invalid token"})
			return
		}

		if sessionId != "" {
			next.ServeHTTP(w, r)
			return
		}

		sessionId, loginTimeout, err := cookie.GetSessionIdAndLoginTimeFromCookie(r)
		if err != nil {
			h.logger.Errorf("cookie: invalid cookie: %v", err)
			jsonhttp.Unauthorized(w, &response{Message: "cookie: invalid cookie: " + err.Error()})
			return
		}

		// if the expiry time is over, logout the user
		loginTime, err := time.Parse(time.RFC3339, loginTimeout)
		if err != nil {
			h.logger.Errorf("cookie: invalid login timeout")
			jsonhttp.Unauthorized(w, &response{Message: "cookie: invalid login timeout"})
			return
		}
		if loginTime.Before(time.Now()) {
			err = h.dfsAPI.LogoutUser(sessionId)
			if err == nil {
				h.logger.Errorf("Logging out as cookie login timeout expired")
				jsonhttp.Unauthorized(w, &response{Message: "logging out as cookie login timeout expired"})
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
