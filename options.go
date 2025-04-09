package algnhsa

type RequestType int

const (
	RequestTypeAuto RequestType = iota
	RequestTypeAPIGatewayV1
	RequestTypeAPIGatewayV2
	RequestTypeALB
)

// Options holds the optional parameters.
type Options struct {
	// RequestType sets the expected request type.
	// By default, algnhsa deduces the request type from the lambda function payload.
	RequestType RequestType

	// BinaryContentTypes sets content types that should be treated as binary types.
	// The "*/* value makes algnhsa treat any content type as binary.
	BinaryContentTypes   []string
	binaryContentTypeMap map[string]bool

	// BinaryContentEncoding sets the encoding for binary content.
	// The "*" value makes algnhsa treat any content encoding as binary.
	BinaryContentEncoding    []string
	binaryContentEncodingMap map[string]bool

	// Use API Gateway PathParameters["proxy"] when constructing the request url.
	// Strips the base path mapping when using a custom domain with API Gateway.
	UseProxyPath bool

	// DebugLog enables printing request and response objects to stdout.
	DebugLog bool
}

func (opts *Options) setBinaryContentTypeMap() {
	types := map[string]bool{}
	for _, contentType := range opts.BinaryContentTypes {
		types[contentType] = true
	}
	opts.binaryContentTypeMap = types

	encodings := map[string]bool{}
	for _, encoding := range opts.BinaryContentEncoding {
		encodings[encoding] = true
	}
	opts.binaryContentEncodingMap = encodings
}
