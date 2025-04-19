package algnhsa

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
)

const (
	acceptAllContentType     = "*/*"
	acceptAllContentEncoding = "*"
)

var canonicalSetCookieHeaderKey = http.CanonicalHeaderKey("Set-Cookie")

// lambdaResponse is a combined lambda response.
// It contains common fields from APIGatewayProxyResponse, APIGatewayV2HTTPResponse and ALBTargetGroupResponse.
type lambdaResponse struct {
	StatusCode        int                 `json:"statusCode"`
	Headers           map[string]string   `json:"headers,omitempty"`
	MultiValueHeaders map[string][]string `json:"multiValueHeaders,omitempty"`
	Cookies           []string            `json:"cookies,omitempty"`
	Body              string              `json:"body"`
	IsBase64Encoded   bool                `json:"isBase64Encoded,omitempty"`
}

func newLambdaResponse(w *httptest.ResponseRecorder, opts *Options, requestType RequestType) (lambdaResponse, error) {
	result := w.Result()

	var resp lambdaResponse
	var err error
	switch requestType {
	case RequestTypeAPIGatewayV1:
		resp, err = newAPIGatewayV1Response(result)
	case RequestTypeALB:
		resp, err = newALBResponse(result)
	case RequestTypeAPIGatewayV2:
		resp, err = newAPIGatewayV2Response(result)
	case RequestTypeSQS:
		resp, err = newSQSResponse(result)
	}
	if err != nil {
		return resp, err
	}

	resp.StatusCode = result.StatusCode

	// Set body.
	contentType := result.Header.Get("Content-Type")
	contentEncoding := result.Header.Get("Content-Encoding")
	if opts.binaryContentTypes.contains(acceptAllContentType) ||
		opts.binaryContentTypes.contains(contentType) ||
		opts.binaryContentEncodings.contains(acceptAllContentEncoding) ||
		opts.binaryContentEncodings.contains(contentEncoding) {
		resp.Body = base64.StdEncoding.EncodeToString(w.Body.Bytes())
		resp.IsBase64Encoded = true
	} else {
		resp.Body = w.Body.String()
	}

	return resp, nil
}
