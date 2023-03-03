package algnhsa

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

var albTestEvent = `{
    "requestContext": {
        "elb": {
            "targetGroupArn": "arn:aws:elasticloadbalancing:us-east-2:123456789012:targetgroup/lambda-279XGJDqGZ5rsrHC2Fjr/49e9d65c45c6791a"
        }
    },
    "httpMethod": "GET",
    "path": "/lambda",
	"multiValueQueryStringParameters": { "myKey": ["val1", "val2"] },
    "multiValueHeaders": {
        "accept": ["text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8"],
        "accept-encoding": ["gzip"],
        "accept-language": ["en-US,en;q=0.9"],
        "connection": ["keep-alive"],
        "host": ["lambda-alb-123578498.us-east-2.elb.amazonaws.com"],
        "upgrade-insecure-requests": ["1"],
        "user-agent": ["Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36"],
        "x-amzn-trace-id": ["Root=1-5c536348-3d683b8b04734faae651f476"],
        "x-forwarded-for": ["72.12.164.125"],
        "x-forwarded-port": ["80"],
        "x-forwarded-proto": ["http"],
        "x-imforwards": ["20"],
		"cookie": ["cookie-name=cookie-value;Domain=myweb.com;Secure;HttpOnly","cookie-name=cookie-value;Expires=May 8, 2019"]
    },
    "body": "",
    "isBase64Encoded": false
}
`

var expectedALBDump = RequestDebugDump{
	Method: "GET",
	URL: struct {
		Path    string
		RawPath string
	}{
		Path:    "/lambda",
		RawPath: "",
	},
	RequestURI: "/lambda?myKey=val1&myKey=val2",
	Host:       "lambda-alb-123578498.us-east-2.elb.amazonaws.com",
	RemoteAddr: "72.12.164.125",
	Header: map[string][]string{
		"Accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8"},
		"Accept-Encoding":           {"gzip"},
		"Accept-Language":           {"en-US,en;q=0.9"},
		"Connection":                {"keep-alive"},
		"Host":                      {"lambda-alb-123578498.us-east-2.elb.amazonaws.com"},
		"Cookie":                    {"cookie-name=cookie-value;Domain=myweb.com;Secure;HttpOnly", "cookie-name=cookie-value;Expires=May 8, 2019"},
		"Upgrade-Insecure-Requests": {"1"},
		"User-Agent":                {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36"},
		"X-Amzn-Trace-Id":           {"Root=1-5c536348-3d683b8b04734faae651f476"},
		"X-Forwarded-For":           {"72.12.164.125"},
		"X-Forwarded-Port":          {"80"},
		"X-Forwarded-Proto":         {"http"},
		"X-Imforwards":              {"20"},
	},
	Form: map[string][]string{
		"myKey": {"val1", "val2"},
	},
	Body: "",
}

func dumpALB(payload []byte, opts Options) (RequestDebugDump, error) {
	lh := lambdaHandler{
		httpHandler: http.HandlerFunc(RequestDebugDumpHandler),
		opts:        &opts,
	}
	responseBytes, err := lh.Invoke(context.Background(), payload)
	if err != nil {
		return RequestDebugDump{}, err
	}
	var r events.ALBTargetGroupResponse
	if err := json.Unmarshal(responseBytes, &r); err != nil {
		return RequestDebugDump{}, err
	}
	if r.StatusCode != 200 {
		return RequestDebugDump{}, errors.New("expected status code 200")
	}
	var dump RequestDebugDump
	if err := json.Unmarshal([]byte(r.Body), &dump); err != nil {
		return RequestDebugDump{}, err
	}
	if dump.ALBRequest.HTTPMethod != "GET" {
		return RequestDebugDump{}, errors.New("expected method GET")
	}
	dump.ALBRequest = nil
	return dump, nil
}

func TestALBBase(t *testing.T) {
	asrt := assert.New(t)

	dump, err := dumpALB([]byte(albTestEvent), Options{})
	asrt.NoError(err)

	asrt.Equal(expectedALBDump, dump)
}

func TestALBBase64BodyRequest(t *testing.T) {
	asrt := assert.New(t)

	event := events.ALBTargetGroupRequest{}
	asrt.NoError(json.Unmarshal([]byte(albTestEvent), &event))
	event.IsBase64Encoded = true
	event.Body = "SGVsbG8gZnJvbSBMYW1iZGEh"
	encodedEvent, err := json.Marshal(event)
	asrt.NoError(err)

	dump, err := dumpALB(encodedEvent, Options{})
	asrt.NoError(err)
	expected := expectedALBDump
	expected.Body = "Hello from Lambda!"
	asrt.Equal(expected, dump)
}

func TestALBURLEncoding(t *testing.T) {
	asrt := assert.New(t)

	event := events.ALBTargetGroupRequest{}
	asrt.NoError(json.Unmarshal([]byte(albTestEvent), &event))
	event.Path = "/%D0%BF%D1%80%D0%B8%D0%B2%D0%B5%D1%82"
	event.MultiValueQueryStringParameters["parameter2"] = []string{"тест\""}
	encodedEvent, err := json.Marshal(event)
	asrt.NoError(err)

	dump, err := dumpALB(encodedEvent, Options{})
	asrt.NoError(err)
	expected := expectedALBDump
	expected.RequestURI = "/%D0%BF%D1%80%D0%B8%D0%B2%D0%B5%D1%82?myKey=val1&myKey=val2&parameter2=%D1%82%D0%B5%D1%81%D1%82%22"
	expected.URL.Path = "/привет"
	expected.Form = map[string][]string{
		"myKey":      {"val1", "val2"},
		"parameter2": {"тест\""},
	}
	asrt.Equal(expected, dump)
}

func TestALBResponseHeaders(t *testing.T) {
	asrt := assert.New(t)

	handler := func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		header.Add("X-Foo", "1")
		header.Add("X-Bar", "2")
		header.Add("X-Bar", "3")
		header.Add("Set-Cookie", "cookie1")
		header.Add("Set-Cookie", "cookie2")
		w.WriteHeader(404)
		io.WriteString(w, "FOO")
	}
	lh := lambdaHandler{
		httpHandler: http.HandlerFunc(handler),
		opts:        &Options{},
	}
	responseBytes, err := lh.Invoke(context.Background(), []byte(albTestEvent))
	asrt.NoError(err)

	var r events.ALBTargetGroupResponse
	err = json.Unmarshal(responseBytes, &r)
	asrt.NoError(err)
	asrt.Equal(404, r.StatusCode)
	asrt.Equal("FOO", r.Body)
	expectedHeaders := map[string][]string{
		"X-Foo":      {"1"},
		"X-Bar":      {"2", "3"},
		"Set-Cookie": {"cookie1", "cookie2"},
	}
	asrt.Equal(expectedHeaders, r.MultiValueHeaders)
}

func TestALBBase64BodyResponseAll(t *testing.T) {
	testBodyResponseAll(t, albTestEvent)
}

func TestALBBase64BodyResponseNoMatch(t *testing.T) {
	testBase64BodyResponseNoMatch(t, albTestEvent)
}

func TestALBBase64BodyResponseMatch(t *testing.T) {
	testBase64BodyResponseMatch(t, albTestEvent)
}
