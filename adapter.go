package algnhsa

import (
	"context"
	"net/http"

	"github.com/akrylysov/algnhsa/alb"
	"github.com/akrylysov/algnhsa/apigw"
	"github.com/akrylysov/algnhsa/config"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mitchellh/mapstructure"
)

func handleEvent(ctx context.Context, event map[string]interface{}, handler http.Handler, opts *config.Options) (response interface{}, err error) {

	_, isAPIGatewayRequest := event["resource"]

	if isAPIGatewayRequest {
		var request events.APIGatewayProxyRequest
		mapstructure.Decode(event, &request)
		response, err = apigw.HandleEvent(ctx, request, handler, opts)
	} else {
		var request events.ALBTargetGroupRequest
		mapstructure.Decode(event, &request)
		response, err = alb.HandleEvent(ctx, request, handler, opts)
	}

	return
}

var defaultOptions = &config.Options{}

// ListenAndServe starts the AWS Lambda runtime (aws-lambda-go lambda.Start) with a given handler.
func ListenAndServe(handler http.Handler, opts *config.Options) {
	if handler == nil {
		handler = http.DefaultServeMux
	}
	if opts == nil {
		opts = defaultOptions
	}
	opts.SetBinaryContentTypeMap()
	lambda.Start(func(ctx context.Context, event map[string]interface{}) (interface{}, error) {
		return handleEvent(ctx, event, handler, opts)
	})
}
