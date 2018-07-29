package algnhsa

import (
	"encoding/base64"
	"net/http/httptest"
	"unicode"

	"github.com/aws/aws-lambda-go/events"
)

const acceptAllContentType = "*/*"

// binaryCase performs binary case switching on alpha characters.
// Ported from https://github.com/Gi60s/binary-case/blob/b100ba0d63075c28485fd1724d94746f74742107/index.js#L86.
func binaryCase(s string, n int) string {
	if n == 0 {
		return s
	}
	inp := []rune(s)
	for i, c := range inp {
		if c <= unicode.MaxASCII && unicode.IsUpper(c) {
			if n&1 > 0 {
				c += 32
			}
			n >>= 1
		} else if c <= unicode.MaxASCII && unicode.IsLower(c) {
			if n&1 > 0 {
				c -= 32
			}
			n >>= 1
		}
		inp[i] = c
	}
	return string(inp)
}

func newAPIGatewayResponse(w *httptest.ResponseRecorder, binaryContentTypes map[string]bool) (events.APIGatewayProxyResponse, error) {
	event := events.APIGatewayProxyResponse{}

	// Set status code.
	event.StatusCode = w.Code

	// Set headers.
	respHeaders := map[string]string{}
	for k, v := range w.HeaderMap {
		// Workaround for API Gateway not supporting headers with multiple values.
		// https://forums.aws.amazon.com/thread.jspa?threadID=205782
		for i, val := range v {
			respHeaders[binaryCase(k, i)] = val
		}
	}
	event.Headers = respHeaders

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
