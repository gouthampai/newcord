package middleware

import (
	"log"
	"net/http"
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

		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)

		requestID, _ := r.Context().Value(ContextKeyRequestID).(string)
		userID := r.Context().Value(ContextKeyUserID)

		log.Printf(
			"[%s] %s %s %s %d %s user=%v",
			requestID,
			r.RemoteAddr,
			r.Method,
			r.RequestURI,
			rw.statusCode,
			time.Since(start),
			userID,
		)
	})
}
