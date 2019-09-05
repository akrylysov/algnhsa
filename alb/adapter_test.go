package alb

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/akrylysov/algnhsa/config"
	"github.com/aws/aws-lambda-go/events"
)

func assertDeepEqual(t *testing.T, expected interface{}, actual interface{}, testCase interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\nexpected %+v\ngot      %+v\ntest case: %+v", expected, actual, testCase)
	}
}

// arbitrary value to insert as MultiValueHeaders in requests to trigger a multiValue response
var mvHeaders = map[string][]string{
	"Host": []string{"foo.bar.com"},
}

// an "empty" MultiValueHeaders response
var mvHeaderOutput = map[string][]string{"Content-Type": []string{"text/plain; charset=utf-8"}}

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
		if r.Header.Get("X-A") == "1" && r.Header.Get("X-B") == "2" {
			w.Header().Set("X-Bar", "baz")
			w.Header().Add("X-y", "1")
			w.Write([]byte("ok"))
		}
	})
	r.HandleFunc("/mvheaders", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-A") == "1" && reflect.DeepEqual(r.Header["X-B"], []string{"21", "22"}) {
			w.Header().Set("X-Bar", "baz")
			w.Header().Add("X-y", "1")
			w.Header().Add("X-Y", "2")
			w.Write([]byte("ok"))
		}
	})
	r.HandleFunc("/context", func(w http.ResponseWriter, r *http.Request) {
		expectedProxyReq := events.ALBTargetGroupRequest{
			Path: "/context",
			RequestContext: events.ALBTargetGroupRequestContext{
				ELB: events.ELBContext{
					TargetGroupArn: "foo::bar:baz",
				},
			},
		}
		proxyReq, ok := TargetGroupRequestFromContext(r.Context())
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
		req  events.ALBTargetGroupRequest
		opts *config.Options
		resp events.ALBTargetGroupResponse
	}{
		{
			req: events.ALBTargetGroupRequest{
				Path: "/html",
			},
			resp: events.ALBTargetGroupResponse{
				StatusCode: 200,
				Body:       "<html>foo</html>",
				Headers:    map[string]string{"Content-Type": "text/html; charset=utf-8"},
			},
		},
		{
			req: events.ALBTargetGroupRequest{
				Path: "/text",
			},
			resp: events.ALBTargetGroupResponse{
				StatusCode: 200,
				Body:       "ok",
			},
		},
		{
			req: events.ALBTargetGroupRequest{
				Path:              "/text",
				MultiValueHeaders: mvHeaders,
			},
			resp: events.ALBTargetGroupResponse{
				StatusCode:        200,
				Body:              "ok",
				MultiValueHeaders: mvHeaderOutput,
			},
		},
		{
			req: events.ALBTargetGroupRequest{
				Path: "/query-params",
				// ignored since it's a multiValue request
				QueryStringParameters: map[string]string{
					"a": "1",
				},
				MultiValueQueryStringParameters: map[string][]string{
					"b": {"2"},
					"c": {"31", "32", "33"},
				},
				MultiValueHeaders: mvHeaders,
			},
			resp: events.ALBTargetGroupResponse{
				StatusCode:        200,
				Body:              "a=[], b=[2], c=[31 32 33], unknown=[]",
				MultiValueHeaders: mvHeaderOutput,
			},
		},
		{
			req: events.ALBTargetGroupRequest{
				Path: "/query-params",
				QueryStringParameters: map[string]string{
					"a": "1",
					"b": "2",
				},
				// these should be ignored since it's not a multi-value request
				MultiValueQueryStringParameters: map[string][]string{
					"c": {"31", "32", "33"},
				},
			},
			resp: events.ALBTargetGroupResponse{
				StatusCode: 200,
				Body:       "a=[1], b=[2], c=[], unknown=[]",
			},
		},
		{
			req: events.ALBTargetGroupRequest{
				Path: "/path/encode%2Ftest%7C",
			},
			resp: events.ALBTargetGroupResponse{
				StatusCode: 200,
				Body:       "/path/encode/test|",
			},
		},
		{
			req: events.ALBTargetGroupRequest{
				HTTPMethod: "POST",
				Path:       "/post-body",
				Body:       "foobar",
			},
			resp: events.ALBTargetGroupResponse{
				StatusCode: 200,
				Body:       "foobar",
			},
		},
		{
			req: events.ALBTargetGroupRequest{
				HTTPMethod:      "POST",
				Path:            "/post-body",
				Body:            "Zm9vYmFy",
				IsBase64Encoded: true,
			},
			resp: events.ALBTargetGroupResponse{
				StatusCode: 200,
				Body:       "foobar",
			},
		},
		{
			req: events.ALBTargetGroupRequest{
				HTTPMethod: "POST",
				Path:       "/form",
				MultiValueHeaders: map[string][]string{
					"Content-Type":   {"application/x-www-form-urlencoded"},
					"Content-Length": {"19"},
				},
				Body: "f=foo&s=bar&xyz=123",
			},
			resp: events.ALBTargetGroupResponse{
				StatusCode: 200,
				Body:       "foobar",
			},
		},
		{
			req: events.ALBTargetGroupRequest{
				Path: "/status",
			},
			resp: events.ALBTargetGroupResponse{
				StatusCode: 204,
				Headers:    map[string]string{"Content-Type": "image/gif"},
			},
		},
		{
			req: events.ALBTargetGroupRequest{
				Path:              "/status",
				MultiValueHeaders: mvHeaders,
			},
			resp: events.ALBTargetGroupResponse{
				StatusCode:        204,
				MultiValueHeaders: map[string][]string{"Content-Type": {"image/gif"}},
			},
		},
		{
			req: events.ALBTargetGroupRequest{
				Path: "/headers",
				Headers: map[string]string{
					"X-a": "1",
					"x-b": "2",
				},
			},
			resp: events.ALBTargetGroupResponse{
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type": "text/plain; charset=utf-8",
					"X-Bar":        "baz",
					"X-Y":          "1",
				},
				Body: "ok",
			},
		},
		{
			req: events.ALBTargetGroupRequest{
				Path: "/mvheaders",
				MultiValueHeaders: map[string][]string{
					"X-a": {"1"},
					"x-b": {"2"},
					"x-B": {"21", "22"},
				},
			},
			resp: events.ALBTargetGroupResponse{
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
			req: events.ALBTargetGroupRequest{
				Path: "/text",
			},
			opts: &config.Options{
				BinaryContentTypes: []string{"text/plain; charset=utf-8"},
			},
			resp: events.ALBTargetGroupResponse{
				StatusCode:      200,
				Body:            "b2s=",
				IsBase64Encoded: true,
			},
		},
		{
			req: events.ALBTargetGroupRequest{
				Path: "/text",
			},
			opts: &config.Options{
				BinaryContentTypes: []string{"*/*"},
			},
			resp: events.ALBTargetGroupResponse{
				StatusCode:      200,
				Body:            "b2s=",
				IsBase64Encoded: true,
			},
		},
		{
			req: events.ALBTargetGroupRequest{
				Path: "/text",
			},
			opts: &config.Options{
				BinaryContentTypes: []string{"text/html; charset=utf-8"},
			},
			resp: events.ALBTargetGroupResponse{
				StatusCode: 200,
				Body:       "ok",
			},
		},
		{
			req: events.ALBTargetGroupRequest{
				Path: "/404",
			},
			resp: events.ALBTargetGroupResponse{
				StatusCode: 404,
				Body:       "404 page not found\n",
				Headers: map[string]string{
					"Content-Type":           "text/plain; charset=utf-8",
					"X-Content-Type-Options": "nosniff",
				},
			},
		},
		{
			req: events.ALBTargetGroupRequest{
				Path:              "/404",
				MultiValueHeaders: mvHeaders,
			},
			resp: events.ALBTargetGroupResponse{
				StatusCode: 404,
				Body:       "404 page not found\n",
				MultiValueHeaders: map[string][]string{
					"Content-Type":           {"text/plain; charset=utf-8"},
					"X-Content-Type-Options": {"nosniff"},
				},
			},
		},
		{
			req: events.ALBTargetGroupRequest{
				Path: "/context",
				RequestContext: events.ALBTargetGroupRequestContext{
					ELB: events.ELBContext{
						TargetGroupArn: "foo::bar:baz",
					},
				},
			},
			resp: events.ALBTargetGroupResponse{
				StatusCode: 200,
				Body:       "ok",
			},
		},
		{
			req: events.ALBTargetGroupRequest{
				Path: "/hostname",
				Headers: map[string]string{
					"Host": "bar",
				},
			},
			resp: events.ALBTargetGroupResponse{
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type": "text/plain; charset=utf-8",
				},
				Body: "bar",
			},
		},
		{
			req: events.ALBTargetGroupRequest{
				Path:              "/hostname",
				MultiValueHeaders: mvHeaders,
			},
			resp: events.ALBTargetGroupResponse{
				StatusCode: 200,
				MultiValueHeaders: map[string][]string{
					"Content-Type": {"text/plain; charset=utf-8"},
				},
				Body: "foo.bar.com",
			},
		},
		{
			req: events.ALBTargetGroupRequest{
				Path: "/requesturi",
				QueryStringParameters: map[string]string{
					"foo": "bar",
				},
			},
			resp: events.ALBTargetGroupResponse{
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

		if req.MultiValueHeaders == nil && expectedResp.Headers == nil {
			expectedResp.Headers = map[string]string{"Content-Type": "text/plain; charset=utf-8"}
		}

		if req.MultiValueHeaders != nil && expectedResp.MultiValueHeaders == nil {
			expectedResp.MultiValueHeaders = map[string][]string{"Content-Type": {"text/plain; charset=utf-8"}}
		}
		opts := testCase.opts
		if opts == nil {
			opts = &config.Options{}
		}
		opts.SetBinaryContentTypeMap()
		ctx := context.Background()
		resp, err := HandleEvent(ctx, testCase.req, r, opts)
		if err != nil {
			t.Fatal(err)
		}
		assertDeepEqual(t, expectedResp, resp, testCase)
	}
}
