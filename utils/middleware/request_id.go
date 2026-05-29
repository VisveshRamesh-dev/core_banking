package middleware

import (
	"context"
	"fmt"

	"utils/snowflake"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

type contextKey string

// RequestIDKey is the context key under which the request ID is stored.
const RequestIDKey contextKey = "request_id"

const headerRequestID = "X-Request-Id"

// RequestID is a Kratos middleware that reads X-Request-Id from incoming headers,
// generating a snowflake ID if none is present, and echoes it back in the response.
func RequestID() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				rid := tr.RequestHeader().Get(headerRequestID)
				if rid == "" {
					rid = fmt.Sprintf("%d", snowflake.NextID())
				}
				tr.ReplyHeader().Set(headerRequestID, rid)
				ctx = context.WithValue(ctx, RequestIDKey, rid)
			}
			return handler(ctx, req)
		}
	}
}

// FromContext extracts the request ID stored by RequestID middleware.
func FromContext(ctx context.Context) string {
	rid, _ := ctx.Value(RequestIDKey).(string)
	return rid
}
