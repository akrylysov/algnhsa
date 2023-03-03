package algnhsa

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

var apiGatewayV2TestEvent = `{
  "version": "2.0",
  "routeKey": "$default",
  "rawPath": "/my/path",
  "rawQueryString": "parameter1=value1&parameter1=value2&parameter2=value",
  "cookies": [
    "cookie1",
    "cookie2"
  ],
  "headers": {
    "header1": "value1",
    "header2": "value1,value2"
  },
  "queryStringParameters": {
    "parameter1": "value1,value2",
    "parameter2": "value"
  },
  "requestContext": {
    "accountId": "123456789012",
    "apiId": "api-id",
    "authentication": {
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
    "authorizer": {
      "jwt": {
        "claims": {
          "claim1": "value1",
          "claim2": "value2"
        },
        "scopes": [
          "scope1",
          "scope2"
        ]
      }
    },
    "domainName": "id.execute-api.us-east-1.amazonaws.com",
    "domainPrefix": "id",
    "http": {
      "method": "POST",
      "path": "/my/path",
      "protocol": "HTTP/1.1",
      "sourceIp": "IP",
      "userAgent": "agent"
    },
    "requestId": "id",
    "routeKey": "$default",
    "stage": "$default",
    "time": "12/Mar/2020:19:03:58 +0000",
    "timeEpoch": 1583348638390
  },
  "body": "Hello from Lambda",
  "pathParameters": {
    "parameter1": "value1",
	"proxy": "/my/path2"
  },
  "isBase64Encoded": false,
  "stageVariables": {
    "stageVariable1": "value1",
    "stageVariable2": "value2"
  }
}
`

var expectedApiGatewayV2Dump = RequestDebugDump{
	Method: "POST",
	URL: struct {
		Path    string
		RawPath string
	}{
		Path:    "/my/path",
		RawPath: "",
	},
	RequestURI: "/my/path?parameter1=value1&parameter1=value2&parameter2=value",
	Host:       "",
	RemoteAddr: "IP",
	Header: map[string][]string{
		"Header1": {"value1"},
		"Header2": {"value1,value2"},
		"Cookie":  {"cookie1", "cookie2"},
	},
	Form: map[string][]string{
		"parameter1": {"value1", "value2"},
		"parameter2": {"value"},
	},
	Body: "Hello from Lambda",
}

func dumpAPIGatewayV2(payload []byte, opts Options) (RequestDebugDump, error) {
	lh := lambdaHandler{
		httpHandler: http.HandlerFunc(RequestDebugDumpHandler),
		opts:        &opts,
	}
	responseBytes, err := lh.Invoke(context.Background(), payload)
	if err != nil {
		return RequestDebugDump{}, err
	}
	var r events.APIGatewayV2HTTPResponse
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
	if dump.APIGatewayV2Request.RequestContext.HTTP.Method != "POST" {
		fmt.Printf("%+v\n", dump)
		return RequestDebugDump{}, errors.New("expected method POST")
	}
	dump.APIGatewayV2Request = nil
	return dump, nil
}

func TestAPIGatewayV2Base(t *testing.T) {
	asrt := assert.New(t)

	dump, err := dumpAPIGatewayV2([]byte(apiGatewayV2TestEvent), Options{})
	asrt.NoError(err)

	asrt.Equal(expectedApiGatewayV2Dump, dump)
}

func TestAPIGatewayV2ProxyPath(t *testing.T) {
	asrt := assert.New(t)

	dump, err := dumpAPIGatewayV2([]byte(apiGatewayV2TestEvent), Options{UseProxyPath: true})
	asrt.NoError(err)

	expected := expectedApiGatewayV2Dump
	expected.RequestURI = "/my/path2?parameter1=value1&parameter1=value2&parameter2=value"
	expected.URL.Path = "/my/path2"
	asrt.Equal(expected, dump)
}

func TestAPIGatewayV2Base64BodyRequest(t *testing.T) {
	asrt := assert.New(t)

	event := events.APIGatewayV2HTTPRequest{}
	asrt.NoError(json.Unmarshal([]byte(apiGatewayV2TestEvent), &event))
	event.IsBase64Encoded = true
	event.Body = "SGVsbG8gZnJvbSBMYW1iZGE="
	encodedEvent, err := json.Marshal(event)
	asrt.NoError(err)

	dump, err := dumpAPIGatewayV2(encodedEvent, Options{})
	asrt.NoError(err)
	asrt.Equal(expectedApiGatewayV2Dump, dump)
}

func TestAPIGatewayV2URLEncoding(t *testing.T) {
	asrt := assert.New(t)

	event := events.APIGatewayV2HTTPRequest{}
	asrt.NoError(json.Unmarshal([]byte(apiGatewayV2TestEvent), &event))
	event.RawPath = "/%D0%BF%D1%80%D0%B8%D0%B2%D0%B5%D1%82"
	event.RawQueryString = "parameter1=value1&parameter1=value2&parameter2=%D1%82%D0%B5%D1%81%D1%82%22"
	event.QueryStringParameters["parameter2"] = "тест\""
	encodedEvent, err := json.Marshal(event)
	asrt.NoError(err)

	dump, err := dumpAPIGatewayV2(encodedEvent, Options{})
	asrt.NoError(err)
	expected := expectedApiGatewayV2Dump
	expected.RequestURI = "/%D0%BF%D1%80%D0%B8%D0%B2%D0%B5%D1%82?parameter1=value1&parameter1=value2&parameter2=%D1%82%D0%B5%D1%81%D1%82%22"
	expected.URL.Path = "/привет"
	expected.Form = map[string][]string{
		"parameter1": {"value1", "value2"},
		"parameter2": {"тест\""},
	}
	asrt.Equal(expected, dump)
}

func TestAPIGatewayV2ResponseHeaders(t *testing.T) {
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
	responseBytes, err := lh.Invoke(context.Background(), []byte(apiGatewayV2TestEvent))
	asrt.NoError(err)

	var r events.APIGatewayV2HTTPResponse
	err = json.Unmarshal(responseBytes, &r)
	asrt.NoError(err)
	asrt.Equal(404, r.StatusCode)
	asrt.Equal("FOO", r.Body)
	expectedHeaders := map[string]string{
		"X-Foo": "1",
		"X-Bar": "2,3",
	}
	asrt.Equal(expectedHeaders, r.Headers)
	asrt.Equal([]string{"cookie1", "cookie2"}, r.Cookies)
}

func TestAPIGatewayV2Base64BodyResponseAll(t *testing.T) {
	testBodyResponseAll(t, apiGatewayV2TestEvent)
}

func TestAPIGatewayV2Base64BodyResponseNoMatch(t *testing.T) {
	testBase64BodyResponseNoMatch(t, apiGatewayV2TestEvent)
}

func TestAPIGatewayV2Base64BodyResponseMatch(t *testing.T) {
	testBase64BodyResponseMatch(t, apiGatewayV2TestEvent)
}
