package algnhsa

import (
	"net/http"
	"net/http/httptest"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handleEvent(event events.APIGatewayProxyRequest, handler http.Handler, binaryContentTypes map[string]bool) (events.APIGatewayProxyResponse, error) {
	r, err := newHTTPRequest(event)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return newAPIGatewayResponse(w, binaryContentTypes)
}

// ListenAndServe starts the AWS Lambda runtime (aws-lambda-go lambda.Start) with a given handler.
// It accepts a slice of content types that should be treated as binary types by the API Gateway.
// The "*/* value makes algnhsa treat any content type as binary.
func ListenAndServe(handler http.Handler, binaryContentTypes []string) {
	if handler == nil {
		handler = http.DefaultServeMux
	}
	types := map[string]bool{}
	for _, contentType := range binaryContentTypes {
		types[contentType] = true
	}
	lambda.Start(func(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		return handleEvent(event, handler, types)
	})
}
