package algnhsa

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

type key int

const (
	requestContextKey key = iota
	requestIDContextKey
)

func newContext(ctx context.Context, event events.APIGatewayProxyRequest) context.Context {

	// if the lambdacontext exists extract the the requestID into something
	// less lambda specific
	if lc, ok := lambdacontext.FromContext(ctx); ok {
		ctx = context.WithValue(ctx, requestIDContextKey, lc.AwsRequestID)
	}

	return context.WithValue(ctx, requestContextKey, event)
}

// ProxyRequestFromContext extracts the APIGatewayProxyRequest event from ctx.
func ProxyRequestFromContext(ctx context.Context) (events.APIGatewayProxyRequest, bool) {
	event, ok := ctx.Value(requestContextKey).(events.APIGatewayProxyRequest)
	return event, ok
}

// RequestIDFromContext extracts the APIGatewayProxyRequest event from ctx
func RequestIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDContextKey).(string)
	return id, ok
}
