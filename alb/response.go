package alb

import (
	"encoding/base64"
	"net/http/httptest"

	"github.com/aws/aws-lambda-go/events"
)

const acceptAllContentType = "*/*"

func NewALBResponse(w *httptest.ResponseRecorder, binaryContentTypes map[string]bool, multiValue bool) (events.ALBTargetGroupResponse, error) {
	event := events.ALBTargetGroupResponse{}

	// Set status code.
	event.StatusCode = w.Code

	// Per AWS docs: You must use multiValueHeaders if you have enabled multi-value headers and headers otherwise
	// In practice - leaving headers null when multi-value is enabled (and vice versa) result in the ALB
	// returning a 502 Bad Gateway
	if multiValue {
		event.MultiValueHeaders = w.Result().Header
	} else {

		singleValueHeaders := make(map[string]string)

		// we can only return one header for each key, so use the first
		for key := range w.Result().Header {
			singleValueHeaders[key] = w.Result().Header.Get(key)
		}

		event.Headers = singleValueHeaders
	}

	// Set body.
	contentType := w.Header().Get("Content-Type")
	if binaryContentTypes[acceptAllContentType] || binaryContentTypes[contentType] {
		event.Body = base64.StdEncoding.EncodeToString(w.Body.Bytes())
		event.IsBase64Encoded = true
	} else {
		event.Body = w.Body.String()
	}

	return event, nil
}
