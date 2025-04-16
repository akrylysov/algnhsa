package algnhsa

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-lambda-go/events"
)

var apiGatewayV1TestEvent = `{
  "version": "1.0",
  "resource": "/my/path",
  "path": "/my/path",
  "httpMethod": "GET",
  "headers": {
    "header1": "value1",
    "header2": "value2"
  },
  "multiValueHeaders": {
    "header1": [
      "value1"
    ],
    "header2": [
      "value1",
      "value2"
    ],
	"cookie": [
		"cookie1",
		"cookie2"
	]
  },
  "queryStringParameters": {
    "parameter1": "value1",
    "parameter2": "value"
  },
  "multiValueQueryStringParameters": {
    "parameter1": [
      "value1",
      "value2"
    ],
    "parameter2": [
      "value"
    ]
  },
  "requestContext": {
    "accountId": "123456789012",
    "apiId": "id",
    "authorizer": {
      "claims": null,
      "scopes": null
    },
    "domainName": "id.execute-api.us-east-1.amazonaws.com",
    "domainPrefix": "id",
    "extendedRequestId": "request-id",
    "httpMethod": "GET",
    "identity": {
      "accessKey": null,
      "accountId": null,
      "caller": null,
      "cognitoAuthenticationProvider": null,
      "cognitoAuthenticationType": null,
      "cognitoIdentityId": null,
      "cognitoIdentityPoolId": null,
      "principalOrgId": null,
      "sourceIp": "192.0.2.1",
      "user": null,
      "userAgent": "user-agent",
      "userArn": null,
      "clientCert": {
        "clientCertPem": "CERT_CONTENT",
        "subjectDN": "www.example.com",
        "issuerDN": "Example issuer",
        "serialNumber": "a1:a1:a1:a1:a1:a1:a1:a1:a1:a1:a1:a1:a1:a1:a1:a1",
        "validity": {
          "notBefore": "May 28 12:30:02 2019 GMT",
          "notAfter": "Aug  5 09:36:04 2021 GMT"
        }
      }
    },
    "path": "/my/path",
    "protocol": "HTTP/1.1",
    "requestId": "id=",
    "requestTime": "04/Mar/2020:19:15:17 +0000",
    "requestTimeEpoch": 1583349317135,
    "resourceId": null,
    "resourcePath": "/my/path",
    "stage": "$default"
  },
  "pathParameters": {"proxy": "/my/path2"},
  "stageVariables": null,
  "body": "Hello from Lambda!",
  "isBase64Encoded": false
}
`

var expectedApiGatewayV1Dump = RequestDebugDump{
	Method: "GET",
	URL: struct {
		Path    string
		RawPath string
	}{
		Path:    "/my/path",
		RawPath: "",
	},
	RequestURI: "/my/path?parameter1=value1&parameter1=value2&parameter2=value",
	Host:       "",
	RemoteAddr: "192.0.2.1",
	Header: map[string][]string{
		"Header1": {"value1"},
		"Header2": {"value1", "value2"},
		"Cookie":  {"cookie1", "cookie2"},
	},
	Form: map[string][]string{
		"parameter1": {"value1", "value2"},
		"parameter2": {"value"},
	},
	Body: "Hello from Lambda!",
}

func dumpAPIGatewayV1(payload []byte, opts Options) (RequestDebugDump, error) {
	lh := lambdaHandler{
		httpHandler: http.HandlerFunc(RequestDebugDumpHandler),
		opts:        &opts,
	}
	responseBytes, err := lh.Invoke(context.Background(), payload)
	if err != nil {
		return RequestDebugDump{}, err
	}
	var r events.APIGatewayProxyResponse
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
	if dump.APIGatewayV1Request.RequestContext.HTTPMethod != "GET" {
		return RequestDebugDump{}, errors.New("expected method GET")
	}
	dump.APIGatewayV1Request = nil
	return dump, nil
}

func TestAPIGatewayV1Base(t *testing.T) {
	asrt := assert.New(t)

	dump, err := dumpAPIGatewayV1([]byte(apiGatewayV1TestEvent), Options{})
	asrt.NoError(err)

	asrt.Equal(expectedApiGatewayV1Dump, dump)
}

func TestAPIGatewayV1ProxyPath(t *testing.T) {
	asrt := assert.New(t)

	dump, err := dumpAPIGatewayV1([]byte(apiGatewayV1TestEvent), Options{UseProxyPath: true})
	asrt.NoError(err)

	expected := expectedApiGatewayV1Dump
	expected.RequestURI = "/my/path2?parameter1=value1&parameter1=value2&parameter2=value"
	expected.URL.Path = "/my/path2"
	asrt.Equal(expected, dump)
}

func TestAPIGatewayV1Base64BodyRequest(t *testing.T) {
	asrt := assert.New(t)

	event := events.APIGatewayProxyRequest{}
	asrt.NoError(json.Unmarshal([]byte(apiGatewayV1TestEvent), &event))
	event.IsBase64Encoded = true
	event.Body = "SGVsbG8gZnJvbSBMYW1iZGEh"
	encodedEvent, err := json.Marshal(event)
	asrt.NoError(err)

	dump, err := dumpAPIGatewayV1(encodedEvent, Options{})
	asrt.NoError(err)
	asrt.Equal(expectedApiGatewayV1Dump, dump)
}

func TestAPIGatewayV1URLEncoding(t *testing.T) {
	asrt := assert.New(t)

	event := events.APIGatewayProxyRequest{}
	asrt.NoError(json.Unmarshal([]byte(apiGatewayV1TestEvent), &event))
	event.Path = "/%D0%BF%D1%80%D0%B8%D0%B2%D0%B5%D1%82"
	event.MultiValueQueryStringParameters["parameter2"] = []string{"тест\""}
	encodedEvent, err := json.Marshal(event)
	asrt.NoError(err)

	dump, err := dumpAPIGatewayV1(encodedEvent, Options{})
	asrt.NoError(err)
	expected := expectedApiGatewayV1Dump
	expected.RequestURI = "/%D0%BF%D1%80%D0%B8%D0%B2%D0%B5%D1%82?parameter1=value1&parameter1=value2&parameter2=%D1%82%D0%B5%D1%81%D1%82%22"
	expected.URL.Path = "/привет"
	expected.Form = map[string][]string{
		"parameter1": {"value1", "value2"},
		"parameter2": {"тест\""},
	}
	asrt.Equal(expected, dump)
}

func TestAPIGatewayV1ResponseHeaders(t *testing.T) {
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
	responseBytes, err := lh.Invoke(context.Background(), []byte(apiGatewayV1TestEvent))
	asrt.NoError(err)

	var r events.APIGatewayProxyResponse
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

func testBodyResponseAll(t *testing.T, event string) {
	t.Helper()
	asrt := assert.New(t)

	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello from Lambda!")
	}
	lh := lambdaHandler{
		httpHandler: http.HandlerFunc(handler),
		opts:        &Options{BinaryContentTypes: []string{"*/*"}},
	}
	lh.opts.init()
	responseBytes, err := lh.Invoke(context.Background(), []byte(event))
	asrt.NoError(err)

	var r lambdaResponse
	err = json.Unmarshal(responseBytes, &r)
	asrt.NoError(err)
	asrt.Equal(200, r.StatusCode)
	asrt.Equal("SGVsbG8gZnJvbSBMYW1iZGEh", r.Body)
}

func testBase64BodyResponseNoMatch(t *testing.T, event string) {
	asrt := assert.New(t)

	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello from Lambda!")
	}
	lh := lambdaHandler{
		httpHandler: http.HandlerFunc(handler),
		opts:        &Options{BinaryContentTypes: []string{"image/png"}},
	}
	lh.opts.init()
	responseBytes, err := lh.Invoke(context.Background(), []byte(event))
	asrt.NoError(err)

	var r events.APIGatewayProxyResponse
	err = json.Unmarshal(responseBytes, &r)
	asrt.NoError(err)
	asrt.Equal(200, r.StatusCode)
	asrt.Equal("Hello from Lambda!", r.Body)
}

func testBase64BodyResponseMatch(t *testing.T, event string) {
	asrt := assert.New(t)

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		io.WriteString(w, "Hello from Lambda!")
	}
	lh := lambdaHandler{
		httpHandler: http.HandlerFunc(handler),
		opts:        &Options{BinaryContentTypes: []string{"image/png"}},
	}
	lh.opts.init()
	responseBytes, err := lh.Invoke(context.Background(), []byte(event))
	asrt.NoError(err)

	var r events.APIGatewayProxyResponse
	err = json.Unmarshal(responseBytes, &r)
	asrt.NoError(err)
	asrt.Equal(200, r.StatusCode)
	asrt.Equal("SGVsbG8gZnJvbSBMYW1iZGEh", r.Body)
}

func TestAPIGatewayV1Base64BodyResponseAll(t *testing.T) {
	testBodyResponseAll(t, apiGatewayV1TestEvent)
}

func TestAPIGatewayV1Base64BodyResponseNoMatch(t *testing.T) {
	testBase64BodyResponseNoMatch(t, apiGatewayV1TestEvent)
}

func TestAPIGatewayV1Base64BodyResponseMatch(t *testing.T) {
	testBase64BodyResponseMatch(t, apiGatewayV1TestEvent)
}

func testBodyContentEncodingResponseAll(t *testing.T, event string) {
	t.Helper()
	asrt := assert.New(t)

	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello from Lambda!")
	}
	lh := lambdaHandler{
		httpHandler: http.HandlerFunc(handler),
		opts:        &Options{BinaryContentEncodings: []string{"*"}},
	}
	lh.opts.init()
	responseBytes, err := lh.Invoke(context.Background(), []byte(event))
	asrt.NoError(err)

	var r lambdaResponse
	err = json.Unmarshal(responseBytes, &r)
	asrt.NoError(err)
	asrt.Equal(200, r.StatusCode)
	asrt.Equal("SGVsbG8gZnJvbSBMYW1iZGEh", r.Body)
}

func testBase64BodyContentEncodingResponseNoMatch(t *testing.T, event string) {
	asrt := assert.New(t)

	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello from Lambda!")
	}
	lh := lambdaHandler{
		httpHandler: http.HandlerFunc(handler),
		opts:        &Options{BinaryContentEncodings: []string{"gzip"}},
	}
	lh.opts.init()
	responseBytes, err := lh.Invoke(context.Background(), []byte(event))
	asrt.NoError(err)

	var r events.APIGatewayProxyResponse
	err = json.Unmarshal(responseBytes, &r)
	asrt.NoError(err)
	asrt.Equal(200, r.StatusCode)
	asrt.Equal("Hello from Lambda!", r.Body)
}

func testBase64BodyContentEncodingResponseMatch(t *testing.T, event string) {
	asrt := assert.New(t)

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		io.WriteString(w, "Hello from Lambda!")
	}
	lh := lambdaHandler{
		httpHandler: http.HandlerFunc(handler),
		opts:        &Options{BinaryContentEncodings: []string{"gzip"}},
	}
	lh.opts.init()
	responseBytes, err := lh.Invoke(context.Background(), []byte(event))
	asrt.NoError(err)

	var r events.APIGatewayProxyResponse
	err = json.Unmarshal(responseBytes, &r)
	asrt.NoError(err)
	asrt.Equal(200, r.StatusCode)
	asrt.Equal("SGVsbG8gZnJvbSBMYW1iZGEh", r.Body)
}

func TestAPIGatewayV1Base64BodyContentEncodingResponseAll(t *testing.T) {
	testBodyContentEncodingResponseAll(t, apiGatewayV1TestEvent)
}

func TestAPIGatewayV1Base64BodyContentEncodingResponseNoMatch(t *testing.T) {
	testBase64BodyContentEncodingResponseNoMatch(t, apiGatewayV1TestEvent)
}

func TestAPIGatewayV1Base64BodyContentEncodingResponseMatch(t *testing.T) {
	testBase64BodyContentEncodingResponseMatch(t, apiGatewayV1TestEvent)
}
