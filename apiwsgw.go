package algnhsa

import (
	"context"
	"encoding/json"
	"errors"
	"path"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

var (
	errAPIGatewayWebsocketUnexpectedRequest = errors.New("expected APIGatewayWebsocketProxyRequest event")
)

func newAPIGatewayWebsocketRequest(ctx context.Context, payload []byte, opts *Options) (lambdaRequest, error) {
	var event events.APIGatewayWebsocketProxyRequest
	if err := json.Unmarshal(payload, &event); err != nil {
		return lambdaRequest{}, err
	}
	if event.RequestContext.APIID == "" || event.RequestContext.EventType == "" {
		return lambdaRequest{}, errAPIGatewayWebsocketUnexpectedRequest
	}

	var overriddenPath bool
	if opts != nil {
		if v, ok := opts.actionPathOverrideMap[strings.ToLower(event.RequestContext.EventType)]; ok {
			event.Path = v.Path
			event.HTTPMethod = v.HTTPMethod
			overriddenPath = true

			// This is behavior that has not been documented or defined anywhere...
			if event.Headers == nil {
				event.Headers = map[string]string{}
			}
			event.Headers["Connection-Id"] = event.RequestContext.ConnectionID
		}
	}

	req := lambdaRequest{
		HTTPMethod:                      event.HTTPMethod,
		Path:                            event.Path,
		QueryStringParameters:           event.QueryStringParameters,
		MultiValueQueryStringParameters: event.MultiValueQueryStringParameters,
		Headers:                         event.Headers,
		MultiValueHeaders:               event.MultiValueHeaders,
		Body:                            event.Body,
		IsBase64Encoded:                 event.IsBase64Encoded,
		SourceIP:                        event.RequestContext.Identity.SourceIP,
		Context:                         newWebsocketProxyRequestContext(ctx, event),
	}

	if opts.UseProxyPath && !overriddenPath {
		req.Path = path.Join("/", event.PathParameters["proxy"])
	}

	return req, nil
}
