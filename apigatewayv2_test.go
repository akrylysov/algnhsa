package algnhsa

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

var testV2Request = `{
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
    "parameter1": "value1"
  },
  "isBase64Encoded": false,
  "stageVariables": {
    "stageVariable1": "value1",
    "stageVariable2": "value2"
  }
}
`

func TestAPIGatewayV2(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/my/path" {
			t.Errorf("expected path %q, got %q", "/my/path", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected method %q, got %q", "POST", r.Method)
		}
		io.WriteString(w, "test")
	})
	lh := lambdaHandler{
		httpHandler: h,
		opts:        &Options{},
	}
	responseBytes, err := lh.Invoke(context.Background(), []byte(testV2Request))
	if err != nil {
		t.Errorf("failed to invoke handler: %v", err)
	}
	var r events.APIGatewayV2HTTPResponse
	err = json.Unmarshal(responseBytes, &r)
	if err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}
	if r.Body != "test" {
		t.Errorf("unexpected body: %q", r.Body)
	}
}
