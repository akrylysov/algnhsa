package algnhsa

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/akrylysov/algnhsa/apigw"
	"github.com/akrylysov/algnhsa/config"
	"github.com/aws/aws-lambda-go/events"
)

// TestRequest abstracts both an APIGatewayProxyRequest and an ALBTargetGroupRequest
// the adapter should handle both identically
type TestRequest struct {
	Resource                        string              `json:"resource"`
	Path                            string              `json:"path"`
	HTTPMethod                      string              `json:"httpMethod"`
	Headers                         map[string]string   `json:"headers"`
	MultiValueHeaders               map[string][]string `json:"multiValueHeaders"`
	QueryStringParameters           map[string]string   `json:"queryStringParameters"`
	MultiValueQueryStringParameters map[string][]string `json:"multiValueQueryStringParameters"`
	PathParameters                  map[string]string   `json:"pathParameters"`
	StageVariables                  map[string]string   `json:"stageVariables"`
	RequestContext                  interface{}         `json:"requestContext"`
	Body                            string              `json:"body"`
	IsBase64Encoded                 bool                `json:"isBase64Encoded,omitempty"`
}

func assertDeepEqual(t *testing.T, expected interface{}, actual interface{}, testCase interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\nexpected %+v\ngot      %+v\ntest case: %+v", expected, actual, testCase)
	}
}

func TestHandleEvent(t *testing.T) {
	r := http.NewServeMux()
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
	r.HandleFunc("/context", func(w http.ResponseWriter, r *http.Request) {
		expectedProxyReq := TestRequest{
			Resource: "foo",
			Path:     "/context",
			RequestContext: events.APIGatewayProxyRequestContext{
				AccountID: "foo",
			},
		}
		proxyReq, ok := apigw.ProxyRequestFromContext(r.Context())
		if ok && reflect.DeepEqual(expectedProxyReq, proxyReq) {
			w.Write([]byte("ok"))
		}
	})
	r.HandleFunc("/hostname", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.Host))
	})
	r.HandleFunc("/requesturi", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.RequestURI))
	})
	testCases := []struct {
		req  TestRequest
		opts *config.Options
		resp events.APIGatewayProxyResponse
	}{
		{
			req: TestRequest{
				Resource: "foo",
				Path:     "/html",
			},
			resp: events.APIGatewayProxyResponse{
				StatusCode:        200,
				Body:              "<html>foo</html>",
				MultiValueHeaders: map[string][]string{"Content-Type": {"text/html; charset=utf-8"}},
			},
		},
		{
			req: TestRequest{
				Resource: "foo",
				Path:     "/text",
			},
			resp: events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       "ok",
			},
		},
		{
			req: TestRequest{
				Resource: "foo",
				Path:     "/query-params",
				QueryStringParameters: map[string]string{
					"a": "1",
					"b": "",
				},
				MultiValueQueryStringParameters: map[string][]string{
					"b": {"2"},
					"c": {"31", "32", "33"},
				},
			},
			resp: events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       "a=[1], b=[2], c=[31 32 33], unknown=[]",
			},
		},
		{
			req: TestRequest{
				Resource: "foo",
				Path:     "/path/encode%2Ftest%7C",
			},
			resp: events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       "/path/encode/test|",
			},
		},
		{
			req: TestRequest{
				Resource:   "foo",
				HTTPMethod: "POST",
				Path:       "/post-body",
				Body:       "foobar",
			},
			resp: events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       "foobar",
			},
		},
		{
			req: TestRequest{
				Resource:        "foo",
				HTTPMethod:      "POST",
				Path:            "/post-body",
				Body:            "Zm9vYmFy",
				IsBase64Encoded: true,
			},
			resp: events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       "foobar",
			},
		},
		{
			req: TestRequest{
				Resource:   "foo",
				HTTPMethod: "POST",
				Path:       "/form",
				MultiValueHeaders: map[string][]string{
					"Content-Type":   {"application/x-www-form-urlencoded"},
					"Content-Length": {"19"},
				},
				Body: "f=foo&s=bar&xyz=123",
			},
			resp: events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       "foobar",
			},
		},
		{
			req: TestRequest{
				Resource: "foo",
				Path:     "/status",
			},
			resp: events.APIGatewayProxyResponse{
				StatusCode:        204,
				MultiValueHeaders: map[string][]string{"Content-Type": {"image/gif"}},
			},
		},
		{
			req: TestRequest{
				Resource: "foo",
				Path:     "/headers",
				Headers: map[string]string{
					"X-a": "1",
					"x-b": "2",
				},
				MultiValueHeaders: map[string][]string{
					"x-B": {"21", "22"},
				},
			},
			resp: events.APIGatewayProxyResponse{
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
			req: TestRequest{
				Resource: "foo",
				Path:     "/text",
			},
			opts: &config.Options{
				BinaryContentTypes: []string{"text/plain; charset=utf-8"},
			},
			resp: events.APIGatewayProxyResponse{
				StatusCode:      200,
				Body:            "b2s=",
				IsBase64Encoded: true,
			},
		},
		{
			req: TestRequest{
				Resource: "foo",
				Path:     "/text",
			},
			opts: &config.Options{
				BinaryContentTypes: []string{"*/*"},
			},
			resp: events.APIGatewayProxyResponse{
				StatusCode:      200,
				Body:            "b2s=",
				IsBase64Encoded: true,
			},
		},
		{
			req: TestRequest{
				Resource: "foo",
				Path:     "/text",
			},
			opts: &config.Options{
				BinaryContentTypes: []string{"text/html; charset=utf-8"},
			},
			resp: events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       "ok",
			},
		},
		{
			req: TestRequest{
				Resource: "foo",
				Path:     "/404",
			},
			resp: events.APIGatewayProxyResponse{
				StatusCode: 404,
				Body:       "404 page not found\n",
				MultiValueHeaders: map[string][]string{
					"Content-Type":           {"text/plain; charset=utf-8"},
					"X-Content-Type-Options": {"nosniff"},
				},
			},
		},
		{
			req: TestRequest{
				Resource: "foo",
				Path:     "/context",
				RequestContext: events.APIGatewayProxyRequestContext{
					AccountID: "foo",
				},
			},
			resp: events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       "ok",
			},
		},
		{
			req: TestRequest{
				Resource: "foo",
				Path:     "/hostname",
				Headers: map[string]string{
					"Host": "bar",
				},
			},
			resp: events.APIGatewayProxyResponse{
				StatusCode: 200,
				MultiValueHeaders: map[string][]string{
					"Content-Type": {"text/plain; charset=utf-8"},
				},
				Body: "bar",
			},
		},
		{
			req: TestRequest{
				Resource: "foo",
				Path:     "/stage/text",
				PathParameters: map[string]string{
					"proxy": "text",
				},
			},
			opts: &config.Options{
				UseProxyPath: true,
			},
			resp: events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       "ok",
			},
		},
		{
			req: TestRequest{
				Resource: "foo",
				Path:     "/requesturi",
				QueryStringParameters: map[string]string{
					"foo": "bar",
				},
			},
			resp: events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       "/requesturi?foo=bar",
			},
		},
	}
	for _, testCase := range testCases {
		req := testCase.req
		if req.HTTPMethod == "" {
			req.HTTPMethod = "GET"
		}
		expectedResp := testCase.resp
		if expectedResp.MultiValueHeaders == nil {
			expectedResp.MultiValueHeaders = map[string][]string{"Content-Type": {"text/plain; charset=utf-8"}}
		}
		opts := testCase.opts
		if opts == nil {
			opts = defaultOptions
		}
		opts.SetBinaryContentTypeMap()
		ctx := context.Background()

		serialized, err := json.Marshal(testCase.req)
		if err != nil {
			t.Fatal(err)
		}

		var mapReq map[string]interface{}
		err = json.Unmarshal(serialized, &mapReq)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := handleEvent(ctx, mapReq, r, opts)
		if err != nil {
			t.Fatal(err)
		}
		assertDeepEqual(t, expectedResp, resp, testCase)
	}
}
