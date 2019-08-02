package alb

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/akrylysov/algnhsa/config"
	"github.com/aws/aws-lambda-go/events"
)

func HandleEvent(ctx context.Context, request events.ALBTargetGroupRequest, handler http.Handler, opts *config.Options) (interface{}, error) {

	multiValue := len(request.MultiValueHeaders) > 0

	r, err := NewHTTPRequest(ctx, request)
	if err != nil {
		return events.ALBTargetGroupResponse{}, err
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return NewALBResponse(w, opts.BinaryContentTypeMap, multiValue)

}
