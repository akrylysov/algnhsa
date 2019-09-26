package algnhsa

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

type adapterTestCase struct {
	req         lambdaRequest
	opts        *Options
	resp        lambdaResponse
	apigwReq    events.APIGatewayProxyRequest
	albReq      events.ALBTargetGroupRequest
	expectedErr error
}

var commonAdapterTestCases = []adapterTestCase{
	{
		req: lambdaRequest{
			Path: "/html",
		},
		resp: lambdaResponse{
			StatusCode:        200,
			Body:              "<html>foo</html>",
			MultiValueHeaders: map[string][]string{"Content-Type": {"text/html; charset=utf-8"}},
		},
	},
	{
		req: lambdaRequest{
			Path: "/text",
		},
		resp: lambdaResponse{
			StatusCode: 200,
			Body:       "ok",
		},
	},
	{
		req: lambdaRequest{
			Path: "/query-params",
			QueryStringParameters: map[string]string{
				"a": "1",
				"b": "",
			},
			MultiValueQueryStringParameters: map[string][]string{
				"b": {"2"},
				"c": {"31", "32", "33"},
			},
		},
		resp: lambdaResponse{
			StatusCode: 200,
			Body:       "a=[1], b=[2], c=[31 32 33], unknown=[]",
		},
	},
	{
		req: lambdaRequest{
			Path: "/path/encode%2Ftest%7C",
		},
		resp: lambdaResponse{
			StatusCode: 200,
			Body:       "/path/encode/test|",
		},
	},
	{
		req: lambdaRequest{
			HTTPMethod: "POST",
			Path:       "/post-body",
			Body:       "foobar",
		},
		resp: lambdaResponse{
			StatusCode: 200,
			Body:       "foobar",
		},
	},
	{
		req: lambdaRequest{
			HTTPMethod:      "POST",
			Path:            "/post-body",
			Body:            "Zm9vYmFy",
			IsBase64Encoded: true,
		},
		resp: lambdaResponse{
			StatusCode: 200,
			Body:       "foobar",
		},
	},
	{
		req: lambdaRequest{
			HTTPMethod: "POST",
			Path:       "/form",
			MultiValueHeaders: map[string][]string{
				"Content-Type":   {"application/x-www-form-urlencoded"},
				"Content-Length": {"19"},
			},
			Body: "f=foo&s=bar&xyz=123",
		},
		resp: lambdaResponse{
			StatusCode: 200,
			Body:       "foobar",
		},
	},
	{
		req: lambdaRequest{
			Path: "/status",
		},
		resp: lambdaResponse{
			StatusCode:        204,
			MultiValueHeaders: map[string][]string{"Content-Type": {"image/gif"}},
		},
	},
	{
		req: lambdaRequest{
			Path: "/headers",
			Headers: map[string]string{
				"X-a": "1",
				"x-b": "2",
			},
			MultiValueHeaders: map[string][]string{
				"x-B": {"21", "22"},
			},
		},
		resp: lambdaResponse{
			StatusCode: 200,
			MultiValueHeaders: map[string][]string{
				"Content-Type": {"text/plain; charset=utf-8"},
				"X-Bar":        {"baz"},
				"X-Y":          {"1", "2"},
			},
			Body: "ok",
		},
	},
	{
		req: lambdaRequest{
			Path: "/text",
		},
		opts: &Options{
			BinaryContentTypes: []string{"text/plain; charset=utf-8"},
		},
		resp: lambdaResponse{
			StatusCode:      200,
			Body:            "b2s=",
			IsBase64Encoded: true,
		},
	},
	{
		req: lambdaRequest{
			Path: "/text",
		},
		opts: &Options{
			BinaryContentTypes: []string{"*/*"},
		},
		resp: lambdaResponse{
			StatusCode:      200,
			Body:            "b2s=",
			IsBase64Encoded: true,
		},
	},
	{
		req: lambdaRequest{
			Path: "/text",
		},
		opts: &Options{
			BinaryContentTypes: []string{"text/html; charset=utf-8"},
		},
		resp: lambdaResponse{
			StatusCode: 200,
			Body:       "ok",
		},
	},
	{
		req: lambdaRequest{
			Path: "/404",
		},
		resp: lambdaResponse{
			StatusCode: 404,
			Body:       "404 page not found\n",
			MultiValueHeaders: map[string][]string{
				"Content-Type":           {"text/plain; charset=utf-8"},
				"X-Content-Type-Options": {"nosniff"},
			},
		},
	},
	{
		req: lambdaRequest{
			Path: "/hostname",
			Headers: map[string]string{
				"Host": "bar",
			},
		},
		resp: lambdaResponse{
			StatusCode: 200,
			MultiValueHeaders: map[string][]string{
				"Content-Type": {"text/plain; charset=utf-8"},
			},
			Body: "bar",
		},
	},
	{
		req: lambdaRequest{
			Path: "/requesturi",
			QueryStringParameters: map[string]string{
				"foo": "bar",
			},
		},
		resp: lambdaResponse{
			StatusCode: 200,
			Body:       "/requesturi?foo=bar",
		},
	},
}

var apigwAdapterTestCases = []adapterTestCase{
	{
		req: lambdaRequest{
			Path: "/apigw/text",
		},
		apigwReq: events.APIGatewayProxyRequest{
			PathParameters: map[string]string{
				"proxy": "text",
			},
		},
		opts: &Options{
			UseProxyPath: true,
		},
		resp: lambdaResponse{
			StatusCode: 200,
			Body:       "ok",
		},
	},
	{
		req: lambdaRequest{
			Path: "/apigw/context",
		},
		apigwReq: events.APIGatewayProxyRequest{
			RequestContext: events.APIGatewayProxyRequestContext{
				AccountID: "foo",
			},
		},
		resp: lambdaResponse{
			StatusCode: 200,
			Body:       "ok",
		},
	},
	{
		req: lambdaRequest{
			Path: "/apigw/wrongtype",
		},
		opts: &Options{
			RequestType: RequestTypeALB,
		},
		expectedErr: errALBUnexpectedRequest,
	},
}

var albAdapterTestCases = []adapterTestCase{
	{
		req: lambdaRequest{
			Path: "/alb/context",
		},
		albReq: events.ALBTargetGroupRequest{
			RequestContext: events.ALBTargetGroupRequestContext{
				ELB: events.ELBContext{
					TargetGroupArn: "foo",
				},
			},
		},
		resp: lambdaResponse{
			StatusCode: 200,
			Body:       "ok",
		},
	},
	{
		req: lambdaRequest{
			Path: "/alb/wrongtype",
		},
		opts: &Options{
			RequestType: RequestTypeAPIGateway,
		},
		expectedErr: errAPIGatewayUnexpectedRequest,
	},
}

func testHandle(t *testing.T, testCases []adapterTestCase, testMode RequestType, requestType RequestType) {
	t.Helper()
	asrt := assert.New(t)

	r := http.NewServeMux()

	// Common handlers
	r.HandleFunc("/html", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html>foo</html>"))
	})
	r.HandleFunc("/text", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	r.HandleFunc("/query-params", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		fmt.Fprintf(w, "a=%s, b=%s, c=%s, unknown=%v", r.Form["a"], r.Form["b"], r.Form["c"], r.Form["unknown"])
	})
	r.HandleFunc("/path/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.URL.Path))
	})
	r.HandleFunc("/post-body", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				fmt.Fprintf(w, "%v", err)
			} else {
				w.Write(body)
			}
		}
	})
	r.HandleFunc("/form", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.Write([]byte(r.FormValue("f") + r.FormValue("s") + r.FormValue("unknown")))
		}
	})
	r.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/gif")
		w.WriteHeader(204)
	})
	r.HandleFunc("/headers", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-A") == "1" && reflect.DeepEqual(r.Header["X-B"], []string{"21", "22"}) {
			w.Header().Set("X-Bar", "baz")
			w.Header().Add("X-y", "1")
			w.Header().Add("X-Y", "2")
			w.Write([]byte("ok"))
		}
	})
	r.HandleFunc("/hostname", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.Host))
	})
	r.HandleFunc("/requesturi", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.RequestURI))
	})

	// APIGateway specific handlers
	r.HandleFunc("/apigw/context", func(w http.ResponseWriter, r *http.Request) {
		expectedProxyReq := events.APIGatewayProxyRequest{
			HTTPMethod: "GET",
			Path:       "/apigw/context",
			RequestContext: events.APIGatewayProxyRequestContext{
				AccountID: "foo",
			},
		}
		proxyReq, ok := ProxyRequestFromContext(r.Context())
		if ok && reflect.DeepEqual(expectedProxyReq, proxyReq) {
			w.Write([]byte("ok"))
		}
	})

	// ALB specific handlers
	r.HandleFunc("/alb/context", func(w http.ResponseWriter, r *http.Request) {
		expectedProxyReq := events.ALBTargetGroupRequest{
			HTTPMethod: "GET",
			Path:       "/alb/context",
			MultiValueHeaders: map[string][]string{
				"X-Test": {"1"},
			},
			RequestContext: events.ALBTargetGroupRequestContext{
				ELB: events.ELBContext{
					TargetGroupArn: "foo",
				},
			},
		}
		targetReq, ok := TargetGroupRequestFromContext(r.Context())
		if ok && reflect.DeepEqual(expectedProxyReq, targetReq) {
			w.Write([]byte("ok"))
		}
	})

	for _, testCase := range testCases {
		lambdaReq := testCase.req
		if lambdaReq.HTTPMethod == "" {
			lambdaReq.HTTPMethod = "GET"
		}

		expectedResp := testCase.resp
		if expectedResp.MultiValueHeaders == nil {
			expectedResp.MultiValueHeaders = map[string][]string{"Content-Type": {"text/plain; charset=utf-8"}}
		}

		lambdaPayload, err := json.Marshal(lambdaReq)
		asrt.NoError(err)

		var payload []byte
		if testMode == RequestTypeAPIGateway {
			req := testCase.apigwReq
			err = json.Unmarshal(lambdaPayload, &req)
			asrt.NoError(err)
			if req.RequestContext.AccountID == "" {
				req.RequestContext.AccountID = "test"
			}
			payload, err = json.Marshal(req)
			asrt.NoError(err)
		} else {
			req := testCase.albReq
			err := json.Unmarshal(lambdaPayload, &req)
			asrt.NoError(err)
			if req.RequestContext.ELB.TargetGroupArn == "" {
				req.RequestContext.ELB.TargetGroupArn = "test"
			}
			if req.MultiValueHeaders == nil {
				req.MultiValueHeaders = map[string][]string{
					"X-Test": {"1"},
				}
			}
			payload, err = json.Marshal(req)
			asrt.NoError(err)
		}

		opts := testCase.opts
		if opts == nil {
			opts = defaultOptions
			opts.RequestType = requestType
		}
		opts.setBinaryContentTypeMap()
		handler := lambdaHandler{httpHandler: r, opts: opts}
		resp, err := handler.handleEvent(context.Background(), payload)
		if testCase.expectedErr == nil {
			asrt.NoError(err)
			asrt.EqualValues(expectedResp, resp, testCase)
		} else {
			asrt.Equal(testCase.expectedErr, err)
		}
	}
}

func TestHandleAPIGatewayAuto(t *testing.T) {
	var testCases []adapterTestCase
	testCases = append(testCases, commonAdapterTestCases...)
	testCases = append(testCases, apigwAdapterTestCases...)
	testHandle(t, testCases, RequestTypeAPIGateway, RequestTypeAuto)
}

func TestHandleAPIGatewayForced(t *testing.T) {
	var testCases []adapterTestCase
	testCases = append(testCases, commonAdapterTestCases...)
	testCases = append(testCases, apigwAdapterTestCases...)
	testHandle(t, testCases, RequestTypeAPIGateway, RequestTypeAPIGateway)
}

func TestHandleALBAuto(t *testing.T) {
	var testCases []adapterTestCase
	testCases = append(testCases, commonAdapterTestCases...)
	testCases = append(testCases, albAdapterTestCases...)
	testHandle(t, testCases, RequestTypeALB, RequestTypeAuto)
}

func TestHandleALBForced(t *testing.T) {
	var testCases []adapterTestCase
	testCases = append(testCases, commonAdapterTestCases...)
	testCases = append(testCases, albAdapterTestCases...)
	testHandle(t, testCases, RequestTypeALB, RequestTypeALB)
}
