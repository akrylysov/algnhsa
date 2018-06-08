package algnhsa

// Options holds the optional parameters.
type Options struct {
	// BinaryContentTypes sets content types that should be treated as binary types by API Gateway.
	// The "*/* value makes algnhsa treat any content type as binary.
	BinaryContentTypes   []string
	binaryContentTypeMap map[string]bool

	// Use API Gateway PathParameters["proxy"] when constructing the request url.
	// Strips the base path mapping when using a custom domain with API Gateway.
	UseProxyPath bool
}

func (opts *Options) setBinaryContentTypeMap() {
	types := map[string]bool{}
	for _, contentType := range opts.BinaryContentTypes {
		types[contentType] = true
	}
	opts.binaryContentTypeMap = types
}
