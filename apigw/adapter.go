package apigw

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/akrylysov/algnhsa/config"
	"github.com/aws/aws-lambda-go/events"
)

func HandleEvent(ctx context.Context, request events.APIGatewayProxyRequest, handler http.Handler, opts *config.Options) (events.APIGatewayProxyResponse, error) {

	r, err := newHTTPRequest(ctx, request, opts.UseProxyPath)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return newAPIGatewayResponse(w, opts.BinaryContentTypeMap)

}
