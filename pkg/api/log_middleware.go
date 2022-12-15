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
	"net"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// LogMiddleware the server middleware which gets called when an api handler is executed.
// it is used to collect certain information about the api call and log it.
func (h *Handler) LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		rl := &responseLogger{w, 0, 0, logrus.InfoLevel}

		next.ServeHTTP(rl, r)

		if rl.level == 0 {
			return
		}

		status := rl.status
		if status == 0 {
			status = http.StatusOK
		}
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}
		fields := logrus.Fields{
			"ip":       ip,
			"method":   r.Method,
			"uri":      r.RequestURI,
			"proto":    r.Proto,
			"status":   status,
			"size":     rl.size,
			"duration": time.Since(startTime).Seconds(),
		}
		if v := r.Referer(); v != "" {
			fields["referrer"] = v
		}
		if v := r.UserAgent(); v != "" {
			fields["user-agent"] = v
		}
		if v := r.Header.Get("X-Forwarded-For"); v != "" {
			fields["x-forwarded-for"] = v
		}
		if v := r.Header.Get("X-Real-Ip"); v != "" {
			fields["x-real-ip"] = v
		}
		h.logger.WithFields(fields).Log(rl.level, "api access")
	})
}

type responseLogger struct {
	w      http.ResponseWriter
	status int
	size   int
	level  logrus.Level
}

// Header
func (l *responseLogger) Header() http.Header {
	return l.w.Header()
}

// Flush
func (l *responseLogger) Flush() {
	l.w.(http.Flusher).Flush()
}

// CloseNotify
func (l *responseLogger) CloseNotify() <-chan bool {
	// staticcheck SA1019 CloseNotifier interface is required by gorilla compress handler
	// nolint:staticcheck
	return l.w.(http.CloseNotifier).CloseNotify() // skipcq: SCC-SA1019
}

// Push
func (l *responseLogger) Push(target string, opts *http.PushOptions) error {
	return l.w.(http.Pusher).Push(target, opts)
}

func (l *responseLogger) Write(b []byte) (int, error) {
	size, err := l.w.Write(b)
	l.size += size
	return size, err
}

// WriteHeader
func (l *responseLogger) WriteHeader(s int) {
	l.w.WriteHeader(s)
	if l.status == 0 {
		l.status = s
	}
}
