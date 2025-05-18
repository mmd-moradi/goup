package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type ResponseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *ResponseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func RequestLogger(logger zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := &ResponseWriterWrapper{w, http.StatusOK}

			next.ServeHTTP(ww, r)

			duration := time.Since(start)

			event := logger.Info()

			if ww.statusCode >= 400 && ww.statusCode < 500 {
				event = logger.Warn()
			}
			if ww.statusCode >= 500 {
				event = logger.Error()
			}

			event.
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", ww.statusCode).
				Dur("duration", duration).
				Str("ip", r.RemoteAddr).
				Str("user-agent", r.UserAgent()).
				Msg("request processed")

		})
	}
}
