package proxy

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// We need to wrap the response writer to be able to log the status code https://gist.github.com/Boerworz/b683e46ae0761056a636
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	// WriteHeader(int) is not called if our response implicitly returns 200 OK, so
	// we default to that status code.
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Logger can be used as a middleware chain to log every request before proxying the request
func Logger(handler http.HandlerFunc, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lrw := NewLoggingResponseWriter(w)
		defer func(begin time.Time) {
			fields := []zap.Field{
				zap.String("method", r.Method),
				zap.String("url", r.URL.String()),
				zap.String("proto", r.Proto),
				zap.Int("status", lrw.statusCode),
				zap.String("ip", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
				zap.Duration("took", time.Since(begin)),
			}

			switch {
			case lrw.statusCode >= 500:
				logger.Error("Server error", fields...)
			case lrw.statusCode >= 400:
				logger.Warn("Client error", fields...)
			case lrw.statusCode >= 300:
				logger.Info("Redirection", fields...)
			default:
				logger.Info("Success", fields...)
			}
		}(time.Now())
		handler(lrw, r)
	}
}
