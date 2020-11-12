package algnhsa

import "strings"

type RequestType int

const (
	RequestTypeAuto RequestType = iota
	RequestTypeAPIGateway
	RequestTypeALB
	RequestTypeAPIGatewayWebsocket
)

// Options holds the optional parameters.
type Options struct {
	// RequestType sets the expected request type.
	// By default algnhsa deduces the request type from the lambda function payload.
	RequestType RequestType

	// BinaryContentTypes sets content types that should be treated as binary types.
	// The "*/* value makes algnhsa treat any content type as binary.
	BinaryContentTypes   []string
	binaryContentTypeMap map[string]bool

	// Use API Gateway PathParameters["proxy"] when constructing the request url.
	// Strips the base path mapping when using a custom domain with API Gateway.
	UseProxyPath bool

	// ActionPathOverrideMap allows you to provide a path and http method override
	// for API Gateway Websocket Actions.
	//
	// Example:
	//
	actionPathOverrideMap map[string]actionPathOverride
}

func (opts *Options) ActionPathOverride(action string, method string, path string) {
	if opts.actionPathOverrideMap == nil {
		opts.actionPathOverrideMap = map[string]actionPathOverride{}
	}
	opts.actionPathOverrideMap[strings.ToLower(action)] = actionPathOverride{
		HTTPMethod: method,
		Path:       path,
	}
}

type actionPathOverride struct {
	HTTPMethod string
	Path       string
}

func (opts *Options) setBinaryContentTypeMap() {
	types := map[string]bool{}
	for _, contentType := range opts.BinaryContentTypes {
		types[contentType] = true
	}
	opts.binaryContentTypeMap = types
}
