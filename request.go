package algnhsa

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

func newHTTPRequest(ctx context.Context, event events.APIGatewayProxyRequest, useProxyPath bool) (*http.Request, error) {
	// Build request URL.
	params := url.Values{}
	for k, v := range event.QueryStringParameters {
		params.Set(k, v)
	}
	for k, vals := range event.MultiValueQueryStringParameters {
		params[k] = vals
	}

	u := url.URL{
		Host:     event.Headers["Host"],
		RawPath:  event.Path,
		RawQuery: params.Encode(),
	}
	if useProxyPath {
		u.RawPath = path.Join("/", event.PathParameters["proxy"])
	}

	// Unescape request path
	p, err := url.PathUnescape(u.RawPath)
	if err != nil {
		return nil, err
	}
	u.Path = p

	if u.Path == u.RawPath {
		u.RawPath = ""
	}

	// Handle base64 encoded body.
	var body io.Reader = strings.NewReader(event.Body)
	if event.IsBase64Encoded {
		body = base64.NewDecoder(base64.StdEncoding, body)
	}

	// Create a new request.
	r, err := http.NewRequest(event.HTTPMethod, u.String(), body)
	if err != nil {
		return nil, err
	}

	// Set headers.
	// https://docs.aws.amazon.com/apigateway/latest/developerguide/set-up-lambda-proxy-integrations.html
	// If you specify values for both headers and multiValueHeaders, API Gateway merges them into a single list.
	// If the same key-value pair is specified in both, only the values from multiValueHeaders will appear
	// the merged list.
	for k, v := range event.Headers {
		r.Header.Set(k, v)
	}
	for k, vals := range event.MultiValueHeaders {
		r.Header[http.CanonicalHeaderKey(k)] = vals
	}

	// Set remote IP address.
	r.RemoteAddr = event.RequestContext.Identity.SourceIP

	// Set request URI
	r.RequestURI = u.RequestURI()

	return r.WithContext(newContext(ctx, event)), nil
}
