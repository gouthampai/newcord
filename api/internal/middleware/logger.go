package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID, _ := r.Context().Value(ContextKeyRequestID).(string)

		// WebSocket upgrades hijack the connection, so wrapping the
		// ResponseWriter would break http.Hijacker (and the logged
		// status code would be wrong anyway — the 101 is written on
		// the raw conn). Pass through unwrapped.
		if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
			log.Printf("[%s] %s %s %s (websocket upgrade)",
				requestID, r.RemoteAddr, r.Method, r.RequestURI)
			next.ServeHTTP(w, r)
			return
		}

		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)

		userID := r.Context().Value(ContextKeyUserID)
		log.Printf("[%s] %s %s %s %d %s user=%v",
			requestID, r.RemoteAddr, r.Method, r.RequestURI,
			rw.statusCode, time.Since(start), userID)
	})
}
