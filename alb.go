package algnhsa

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

var (
	errNonALBEvent                  = errors.New("non ALBTargetGroupRequest event")
	errALBExpectedMultiValueHeaders = errors.New("expected multi value headers; enable Multi value headers in target group settings")
)

func getALBSourceIP(event events.ALBTargetGroupRequest) string {
	if xff, ok := event.MultiValueHeaders["x-forwarded-for"]; ok && len(xff) > 0 {
		ips := strings.SplitN(xff[0], ",", 2)
		if len(ips) > 0 {
			return ips[0]
		}
	}
	return ""
}

func newALBRequest(ctx context.Context, payload []byte, opts *Options) (lambdaRequest, error) {
	var event events.ALBTargetGroupRequest
	if err := json.Unmarshal(payload, &event); err != nil {
		return lambdaRequest{}, err
	}
	if event.RequestContext.ELB.TargetGroupArn == "" {
		return lambdaRequest{}, errNonALBEvent
	}
	if len(event.MultiValueHeaders) == 0 {
		return lambdaRequest{}, errALBExpectedMultiValueHeaders
	}

	req := lambdaRequest{}
	req.HTTPMethod = event.HTTPMethod
	req.Path = event.Path
	req.QueryStringParameters = event.QueryStringParameters
	req.MultiValueQueryStringParameters = event.MultiValueQueryStringParameters
	req.Headers = event.Headers
	req.MultiValueHeaders = event.MultiValueHeaders
	req.Body = event.Body
	req.IsBase64Encoded = event.IsBase64Encoded
	req.SourceIP = getALBSourceIP(event)
	req.Context = newTargetGroupRequestContext(ctx, event)

	return req, nil
}
