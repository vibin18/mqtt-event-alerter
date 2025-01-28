package api

import (
	"log/slog"
	"net/http"
	"time"
)

// LoggingMiddleware logs the details of each HTTP request and response
func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := &statusWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(ww, r)

			logger.Info("HTTP Request",
				slog.String("method", r.Method),
				slog.String("url", r.URL.String()),
				slog.Int("status", ww.statusCode),
				slog.Duration("duration", time.Since(start)),
				slog.String("remote_addr", r.RemoteAddr),
			)
		})
	}
}

// statusWriter is a wrapper for http.ResponseWriter to capture the status code
type statusWriter struct {
	http.ResponseWriter
	statusCode int
}
