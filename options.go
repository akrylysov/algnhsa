package algnhsa

type RequestType int

const (
	RequestTypeAuto RequestType = iota
	RequestTypeAPIGatewayV1
	RequestTypeAPIGatewayV2
	RequestTypeALB
)

type set[T comparable] struct {
	items map[T]struct{}
}

func newSet[T comparable](items ...T) *set[T] {
	m := make(map[T]struct{}, len(items))
	for _, item := range items {
		m[item] = struct{}{}
	}
	return &set[T]{items: m}
}

func (s *set[T]) contains(item T) bool {
	if s == nil {
		return false
	}
	_, exists := s.items[item]
	return exists
}

// Options holds the optional parameters.
type Options struct {
	// RequestType sets the expected request type.
	// By default, algnhsa deduces the request type from the lambda function payload.
	RequestType RequestType

	// BinaryContentTypes sets content types that should be treated as binary types.
	// The "*/* value makes algnhsa treat any content type as binary.
	BinaryContentTypes []string
	binaryContentTypes *set[string]

	// BinaryContentEncodings sets content encodings that should be treated as binary.
	// The "*" value makes algnhsa treat any content encoding as binary.
	BinaryContentEncodings []string
	binaryContentEncodings *set[string]

	// Use API Gateway PathParameters["proxy"] when constructing the request url.
	// Strips the base path mapping when using a custom domain with API Gateway.
	UseProxyPath bool

	// DebugLog enables printing request and response objects to stdout.
	DebugLog bool
}

func (opts *Options) init() {
	opts.binaryContentTypes = newSet(opts.BinaryContentTypes...)
	opts.binaryContentEncodings = newSet(opts.BinaryContentEncodings...)
}
