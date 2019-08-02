package alb

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

func NewHTTPRequest(ctx context.Context, event events.ALBTargetGroupRequest) (*http.Request, error) {

	// headers will always be set in one or the other depending
	// on whether or not the target group has use-multi enabled
	multiValue := len(event.MultiValueHeaders) > 0

	// Build request URL.
	params := url.Values{}

	if multiValue {
		for k, vals := range event.MultiValueQueryStringParameters {
			params[k] = vals
		}
	} else {
		for k, v := range event.QueryStringParameters {
			params.Set(k, v)
		}

	}

	u := url.URL{
		Host:     event.Headers["Host"],
		RawPath:  event.Path,
		RawQuery: params.Encode(),
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
	if multiValue {
		for k, vals := range event.MultiValueHeaders {
			r.Header[http.CanonicalHeaderKey(k)] = vals
		}
	} else {
		for k, v := range event.Headers {
			r.Header.Set(k, v)
		}
	}

	// There doesn't seem to be a way to get source IP from an ALB request
	//r.RemoteAddr = event.RequestContext.Identity.SourceIP

	// Set request URI
	r.RequestURI = u.RequestURI()

	return r.WithContext(newContext(ctx, event)), nil
}
