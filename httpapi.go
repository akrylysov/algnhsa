package algnhsa

import (
	"context"
	"encoding/json"
	"errors"
	"path"

	"github.com/aws/aws-lambda-go/events"
)

var (
	errAPIGatewayV2UnexpectedRequest = errors.New("expected APIGatewayV2HTTPRequest event")
)

func newAPIGatewayV2HTTPRequest(ctx context.Context, payload []byte, opts *Options) (lambdaRequest, error) {
	var event events.APIGatewayV2HTTPRequest
	if err := json.Unmarshal(payload, &event); err != nil {
		return lambdaRequest{}, err
	}
	if event.Version != "2.0" {
		return lambdaRequest{}, errAPIGatewayV2UnexpectedRequest
	}

	req := lambdaRequest{
		HTTPMethod:            event.RequestContext.HTTP.Method,
		Path:                  event.RequestContext.HTTP.Path,
		QueryStringParameters: event.QueryStringParameters,
		Headers:               event.Headers,
		Body:                  event.Body,
		IsBase64Encoded:       event.IsBase64Encoded,
		SourceIP:              event.RequestContext.HTTP.SourceIP,
		Context:               newAPIGatewayV2HTTPRequestContext(ctx, event),
	}

	if opts.UseProxyPath {
		req.Path = path.Join("/", event.PathParameters["proxy"])
	}

	return req, nil
}
