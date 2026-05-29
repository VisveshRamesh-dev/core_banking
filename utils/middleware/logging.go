package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

// Logging is a Kratos middleware that logs each request's kind, operation,
// request ID, latency, and any error.
func Logging(logger log.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var operation, kind string
			if tr, ok := transport.FromServerContext(ctx); ok {
				operation = tr.Operation()
				kind = string(tr.Kind())
			}

			start := time.Now()
			reply, err := handler(ctx, req)
			latency := time.Since(start)

			rid := FromContext(ctx)
			level := log.LevelInfo
			if err != nil {
				level = log.LevelError
			}

			_ = log.WithContext(ctx, logger).Log(level,
				"kind", kind,
				"operation", operation,
				"request_id", rid,
				"latency", fmt.Sprintf("%dms", latency.Milliseconds()),
				"error", err,
			)
			return reply, err
		}
	}
}
