package algnhsa

import (
	"context"
	"encoding/json"
	"errors"
	"path"

	"github.com/aws/aws-lambda-go/events"
)

var (
	errNonAPIGateway = errors.New("non APIGatewayProxyRequest event")
)

func newAPIGatewayRequest(ctx context.Context, payload []byte, opts *Options) (lambdaRequest, error) {
	var event events.APIGatewayProxyRequest
	if err := json.Unmarshal(payload, &event); err != nil {
		return lambdaRequest{}, err
	}
	if event.RequestContext.AccountID == "" {
		return lambdaRequest{}, errNonAPIGateway
	}

	req := lambdaRequest{}
	req.HTTPMethod = event.HTTPMethod
	if opts.UseProxyPath {
		req.Path = path.Join("/", event.PathParameters["proxy"])
	} else {
		req.Path = event.Path
	}

	req.QueryStringParameters = event.QueryStringParameters
	req.MultiValueQueryStringParameters = event.MultiValueQueryStringParameters
	req.Headers = event.Headers
	req.MultiValueHeaders = event.MultiValueHeaders
	req.Body = event.Body
	req.IsBase64Encoded = event.IsBase64Encoded
	req.SourceIP = event.RequestContext.Identity.SourceIP
	req.Context = newProxyRequestContext(ctx, event)

	return req, nil
}
