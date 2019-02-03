package algnhsa

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

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
		expectedProxyReq := events.APIGatewayProxyRequest{
			Path: "/context",
			RequestContext: events.APIGatewayProxyRequestContext{
				AccountID: "foo",
			},
		}
		proxyReq, ok := ProxyRequestFromContext(r.Context())
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
		req  events.APIGatewayProxyRequest
		opts *Options
		resp events.APIGatewayProxyResponse
	}{
		{
			req: events.APIGatewayProxyRequest{
				Path: "/html",
			},
			resp: events.APIGatewayProxyResponse{
				StatusCode:        200,
				Body:              "<html>foo</html>",
				MultiValueHeaders: map[string][]string{"Content-Type": {"text/html; charset=utf-8"}},
			},
		},
		{
			req: events.APIGatewayProxyRequest{
				Path: "/text",
			},
			resp: events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       "ok",
			},
		},
		{
			req: events.APIGatewayProxyRequest{
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
			resp: events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       "a=[1], b=[2], c=[31 32 33], unknown=[]",
			},
		},
		{
			req: events.APIGatewayProxyRequest{
				Path: "/path/encode%2Ftest%7C",
			},
			resp: events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       "/path/encode/test|",
			},
		},
		{
			req: events.APIGatewayProxyRequest{
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
			req: events.APIGatewayProxyRequest{
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
			req: events.APIGatewayProxyRequest{
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
			req: events.APIGatewayProxyRequest{
				Path: "/status",
			},
			resp: events.APIGatewayProxyResponse{
				StatusCode:        204,
				MultiValueHeaders: map[string][]string{"Content-Type": {"image/gif"}},
			},
		},
		{
			req: events.APIGatewayProxyRequest{
				Path: "/headers",
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
			req: events.APIGatewayProxyRequest{
				Path: "/text",
			},
			opts: &Options{
				BinaryContentTypes: []string{"text/plain; charset=utf-8"},
			},
			resp: events.APIGatewayProxyResponse{
				StatusCode:      200,
				Body:            "b2s=",
				IsBase64Encoded: true,
			},
		},
		{
			req: events.APIGatewayProxyRequest{
				Path: "/text",
			},
			opts: &Options{
				BinaryContentTypes: []string{"*/*"},
			},
			resp: events.APIGatewayProxyResponse{
				StatusCode:      200,
				Body:            "b2s=",
				IsBase64Encoded: true,
			},
		},
		{
			req: events.APIGatewayProxyRequest{
				Path: "/text",
			},
			opts: &Options{
				BinaryContentTypes: []string{"text/html; charset=utf-8"},
			},
			resp: events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       "ok",
			},
		},
		{
			req: events.APIGatewayProxyRequest{
				Path: "/404",
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
			req: events.APIGatewayProxyRequest{
				Path: "/context",
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
			req: events.APIGatewayProxyRequest{
				Path: "/hostname",
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
			req: events.APIGatewayProxyRequest{
				Path: "/stage/text",
				PathParameters: map[string]string{
					"proxy": "text",
				},
			},
			opts: &Options{
				UseProxyPath: true,
			},
			resp: events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       "ok",
			},
		},
		{
			req: events.APIGatewayProxyRequest{
				Path: "/requesturi",
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
		opts.setBinaryContentTypeMap()
		ctx := context.Background()
		resp, err := handleEvent(ctx, testCase.req, r, opts)
		if err != nil {
			t.Fatal(err)
		}
		assertDeepEqual(t, expectedResp, resp, testCase)
	}
}
