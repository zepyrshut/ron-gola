package ron

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func (e *Engine) TimeOutMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), e.Config.Timeout)
			defer cancel()

			r = r.WithContext(ctx)
			done := make(chan struct{})

			go func() {
				next.ServeHTTP(w, r)
				close(done)
			}()

			select {
			case <-ctx.Done():
				if errors.Is(ctx.Err(), context.DeadlineExceeded) {
					slog.Debug("timeout reached")
					http.Error(w, "Request timed out", http.StatusGatewayTimeout)
				}
			case <-done:
			}
		})
	}
}

func (e *Engine) RequestIdMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			id := r.Header.Get("X-Request-ID")
			if id == "" {
				id = fmt.Sprintf("%d", time.Now().UnixNano())
			}
			ctx = context.WithValue(ctx, RequestID, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
