package algnhsa

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

func TestNewContext(t *testing.T) {

	ctx := context.Background()

	apigwReq := events.APIGatewayProxyRequest{}

	// create a generic empty request context
	reqCtx := newContext(ctx, apigwReq)

	val, ok := RequestIDFromContext(reqCtx)
	if val != "" || ok {
		t.Fatalf("Expected the requestIDContextKey to be empty but got %v", val)
	}

	lc := lambdacontext.NewContext(ctx, &lambdacontext.LambdaContext{
		AwsRequestID: "test",
	})

	// now create a context from a lambda context
	lambdaReqCtx := newContext(lc, apigwReq)

	val, _ = RequestIDFromContext(lambdaReqCtx)
	if val != "test" {
		t.Fatalf("Expected %v but got %v when extracting requestIDContextKey", "test", val)
	}

}

func TestRequestIDFromContext(t *testing.T) {

	ctx := context.Background()

	val, ok := RequestIDFromContext(ctx)
	if ok {
		t.Fatal("Expected ok to be false but got true")
	}
	if val != "" {
		t.Fatalf("Expected val to be empty but got %v", val)
	}

	ctx = context.WithValue(ctx, requestIDContextKey, "test")

	val, ok = RequestIDFromContext(ctx)
	if !ok {
		t.Fatal("Expected ok to be true but got false")
	}
	if val != "test" {
		t.Fatalf("Expected %v but got %v", "test", val)
	}
}
