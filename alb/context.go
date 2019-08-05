package alb

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

type key int

const requestContextKey key = 0

func newContext(ctx context.Context, event events.ALBTargetGroupRequest) context.Context {
	return context.WithValue(ctx, requestContextKey, event)
}

// TargetGroupRequestFromContext extracts the ALBTargetGroupRequest event from ctx.
func TargetGroupRequestFromContext(ctx context.Context) (events.ALBTargetGroupRequest, bool) {
	event, ok := ctx.Value(requestContextKey).(events.ALBTargetGroupRequest)
	return event, ok
}
