package algnhsa

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

// RequestTypeSQS
// https://github.com/aws/aws-lambda-go/blob/v1.47.0/events/sqs.go

var (
	errSQSUnexpectedRequest = errors.New("expected SQSRequest event")
)

func newSQSRequest(ctx context.Context, payload []byte, opts *Options) (lambdaRequest, error) {
	var event events.SQSEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return lambdaRequest{}, err
	}
	body, err := json.Marshal(event.Records)
	req := lambdaRequest{
		HTTPMethod:  "POST",
		Path:        "/sqs",
		Body:        string(body),
		Context:     context.WithValue(ctx, RequestTypeSQS, event),
		requestType: RequestTypeSQS,
	}
	return req, err
}

func newSQSResponse(r *http.Response) (lambdaResponse, error) {
	resp := lambdaResponse{}
	return resp, nil
}

// SQSRequestFromContext extracts the SQSEvent from ctx.
func SQSRequestFromContext(ctx context.Context) (events.SQSEvent, bool) {
	val := ctx.Value(RequestTypeSQS)
	if val == nil {
		return events.SQSEvent{}, false
	}
	event, ok := val.(events.SQSEvent)
	return event, ok
}
