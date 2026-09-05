package internalhttp

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/logger"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func loggingMiddleware(log *logger.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		ctx := logger.WithContext(r.Context(), log.Slog())
		next.ServeHTTP(rec, r.WithContext(ctx))

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		log.Info(fmt.Sprintf(
			`%s [%s] %s %s %s %d %d "%s"`,
			ip,
			start.Format("02/Jan/2006:15:04:05 -0700"),
			r.Method,
			r.URL.RequestURI(),
			r.Proto,
			rec.status,
			time.Since(start).Milliseconds(),
			r.UserAgent(),
		))
	})
}
