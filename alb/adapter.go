package alb

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/akrylysov/algnhsa/config"
	"github.com/aws/aws-lambda-go/events"
)

// HandleEvent translates an ALBTargetGroupRequest for use by an http.Handler
func HandleEvent(ctx context.Context, request events.ALBTargetGroupRequest, handler http.Handler, opts *config.Options) (interface{}, error) {
	r, err := newHTTPRequest(ctx, request)
	if err != nil {
		return events.ALBTargetGroupResponse{}, err
	}

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	multiValue := len(request.MultiValueHeaders) > 0
	return newALBResponse(w, opts.BinaryContentTypeMap, multiValue)
}
