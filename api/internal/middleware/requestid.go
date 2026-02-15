package middleware

import (
	"context"
	"net/http"

	"github.com/gocql/gocql"
)

const ContextKeyRequestID contextKey = "request_id"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = gocql.TimeUUID().String()
		}
		w.Header().Set("X-Request-ID", requestID)
		ctx := context.WithValue(r.Context(), ContextKeyRequestID, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
