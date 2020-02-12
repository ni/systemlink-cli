package unit_test

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ni/systemlink-cli/internal/model"
)

func TestHelp(t *testing.T) {
	models := []model.Data{}

	writer, _ := callCli([]string{"--help"}, models)

	helpOutput := "systemlink - Command-Line Interface for NI SystemLink Services"
	if !strings.Contains(writer.String(), helpOutput) {
		t.Errorf("Help output was wrong, got: %s, but expected to contain: %s.", writer.String(), helpOutput)
	}
}

func TestHelpOutputsAllModels(t *testing.T) {
	models := []model.Data{
		{Name: "mycommand", Content: []byte(`{}`)},
		{Name: "othercommand", Content: []byte(`{}`)},
	}

	writer, _ := callCli([]string{"--help"}, models)

	if !strings.Contains(writer.String(), "mycommand") {
		t.Errorf("Help output was wrong, got: %s, but expected to contain: %s.", writer.String(), "mycommand")
	}
	if !strings.Contains(writer.String(), "othercommand") {
		t.Errorf("Help output was wrong, got: %s, but expected to contain: %s.", writer.String(), "othercommand")
	}
}

func TestOutputsAllSubCommands(t *testing.T) {
	models := []model.Data{
		{Name: "messages", Content: []byte(`{}`)},
		{Name: "tags", Content: []byte(`
---
paths:
  "/tags":
    GET:
      operationId: get-tags
    POST:
      operationId: create-tag
  "/subscriptions":
    PUT:
      operationId: create-subscription
`)},
	}

	writer, _ := callCli([]string{"tags"}, models)

	if !strings.Contains(writer.String(), "get-tags") {
		t.Errorf("Output was wrong, got: %s, but expected to contain: %s.", writer.String(), "get-tags")
	}
	if !strings.Contains(writer.String(), "create-tag") {
		t.Errorf("Output was wrong, got: %s, but expected to contain: %s.", writer.String(), "create-tag")
	}
	if !strings.Contains(writer.String(), "create-subscription") {
		t.Errorf("Output was wrong, got: %s, but expected to contain: %s.", writer.String(), "create-subscription")
	}
}

func TestSupportsJson(t *testing.T) {
	models := []model.Data{
		{Name: "tags", Content: []byte(`
		  {
			"paths": {
			  "/tags": {
				"GET": {
				  "operationId": "get-tags"
				}
			  }
			}
		  }`)},
	}

	writer, _ := callCli([]string{"tags"}, models)

	if !strings.Contains(writer.String(), "get-tags") {
		t.Errorf("Output was wrong, got: %s, but expected to contain: %s.", writer.String(), "get-tags")
	}
}

var methodTests = []struct {
	method string
}{
	{"GET"},
	{"PUT"},
	{"POST"},
	{"DELETE"},
	{"OPTIONS"},
	{"HEAD"},
}

func TestCallsAllMethods(t *testing.T) {
	for _, tt := range methodTests {
		var actualMethod string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actualMethod = r.Method
		}))

		models := []model.Data{
			{
				Name: "messages",
				Content: []byte(`
---
paths:
  "/create-session":
    ` + strings.ToLower(tt.method) + `:
      operationId: create
`),
			},
		}

		callCli([]string{"messages", "create", "--url", server.URL}, models)

		if actualMethod != tt.method {
			t.Errorf("Expected %s request, but got %s", tt.method, actualMethod)
		}
	}
}

func TestOutputsResponseBody(t *testing.T) {
	server := successReponseStub(`{"token":"1234"}`)

	models := []model.Data{
		{
			Name: "messages",
			Content: []byte(`
---
paths:
  "/create-session":
    get:
      operationId: create
`),
		},
	}

	writer, errWriter := callCli([]string{"messages", "create", "--url", server.URL}, models)

	expectedOutput := `{
	"token": "1234"
}
`
	if writer.String() != expectedOutput {
		t.Errorf("Output was wrong, got: %s, but expected: %s", writer.String(), expectedOutput)
	}
	if errWriter.String() != "" {
		t.Errorf("Expected no error output but got: %s", errWriter.String())
	}
}

func TestVerboseOutputsFullRequestAndResponse(t *testing.T) {
	server := successReponseStub("")

	models := []model.Data{
		{
			Name: "messages",
			Content: []byte(`
---
paths:
  "/create-session":
    get:
      operationId: create
`),
		},
	}

	writer, _ := callCli([]string{"messages", "create", "--verbose", "--url", server.URL}, models)

	if !strings.Contains(writer.String(), "GET /create-session HTTP/1.1") {
		t.Errorf("Output was wrong, got: %s, but expected full request dump.", writer.String())
	}
	if !strings.Contains(writer.String(), "HTTP/1.1 200 OK") {
		t.Errorf("Output was wrong, got: %s, but expected full respopnse dump.", writer.String())
	}
}

func TestApiKeyIsAddedToHttpHeader(t *testing.T) {
	var apiKeyHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKeyHeader = r.Header.Get("x-ni-api-key")
	}))

	models := []model.Data{
		{
			Name: "messages",
			Content: []byte(`
---
paths:
  "/create-session":
    get:
      operationId: create
`),
		},
	}

	callCli([]string{"messages", "create", "--api-key", "my-api-key", "--url", server.URL}, models)

	if apiKeyHeader != "my-api-key" {
		t.Errorf("API key not found in HTTP header, got: %s, but expected %s", apiKeyHeader, "my-api-key")
	}
}

func TestBasicAuthIsAddedToHttpHeader(t *testing.T) {
	var basicAuthHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		basicAuthHeader = r.Header.Get("Authorization")
	}))

	models := []model.Data{
		{
			Name: "messages",
			Content: []byte(`
---
paths:
  "/create-session":
    get:
      operationId: create
`),
		},
	}

	callCli([]string{"messages", "create", "--username", "my-user", "--password", "my-password", "--url", server.URL}, models)

	encoded := "Basic " + base64.StdEncoding.EncodeToString([]byte("my-user:my-password"))
	if basicAuthHeader != encoded {
		t.Errorf("BasicAuth not found in HTTP header, got: %s, but expected %s", basicAuthHeader, encoded)
	}
}

func TestRequestIdIsAddedToHttpHeader(t *testing.T) {
	var requestId string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestId = r.Header.Get("x-request-id")
	}))

	models := []model.Data{
		{
			Name: "messages",
			Content: []byte(`
---
paths:
  "/create-session":
    get:
      operationId: create
`),
		},
	}

	callCli([]string{"messages", "create", "--url", server.URL}, models)

	if requestId == "" {
		t.Errorf("Request Id not found in HTTP header, got: %s", requestId)
	}
}

func TestCallsIncludeBodyParameter(t *testing.T) {
	var body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body = readerToString(r.Body)
	}))

	models := []model.Data{
		{
			Name: "messages",
			Content: []byte(`
---
paths:
  "/create-session":
    post:
      operationId: mymethod
      parameters:
      - name: paramA
        type: string
        in: body
`),
		},
	}

	callCli([]string{"messages", "mymethod", "--paramA", "1234", "--url", server.URL}, models)

	if body != `{"paramA":"1234"}` {
		t.Errorf("Expected body to contain paramA, but got %s", body)
	}
}

func TestCallsIncludePathParameter(t *testing.T) {
	var urlPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlPath = r.URL.Path
	}))

	models := []model.Data{
		{
			Name: "messages",
			Content: []byte(`
---
paths:
  "/delete-session/{id}":
    delete:
      operationId: delete-session
      parameters:
      - name: id
        type: string
        in: path
`),
		},
	}

	callCli([]string{"messages", "delete-session", "--id", "1234", "--url", server.URL}, models)

	if urlPath != "/delete-session/1234" {
		t.Errorf("Expected url path to contain id parameter, but got %s", urlPath)
	}
}

func TestCallsIncludeQueryStringParameter(t *testing.T) {
	var query string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
	}))

	models := []model.Data{
		{
			Name: "messages",
			Content: []byte(`
---
paths:
  "/delete-session":
    delete:
      operationId: delete-session
      parameters:
      - name: id
        type: string
        in: query
`),
		},
	}

	callCli([]string{"messages", "delete-session", "--id", "1234", "--url", server.URL}, models)

	if query != "id=1234" {
		t.Errorf("Expected url query to contain id parameter, but got %s", query)
	}
}

func TestCallsIncludeHeaderParameter(t *testing.T) {
	var headers http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers = r.Header
	}))

	models := []model.Data{
		{
			Name: "messages",
			Content: []byte(`
---
paths:
  "/delete-session":
    delete:
      operationId: delete-session
      parameters:
      - name: id
        type: string
        in: header
`),
		},
	}

	callCli([]string{"messages", "delete-session", "--id", "1234", "--url", server.URL}, models)

	if headers.Get("id") != "1234" {
		t.Errorf("Expected header to contain id parameter, but got %s", headers.Get("id"))
	}
}

func TestCallsIncludeParametersWithSameName(t *testing.T) {
	var urlPath string
	var query string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlPath = r.URL.Path
		query = r.URL.RawQuery
	}))

	models := []model.Data{
		{
			Name: "messages",
			Content: []byte(`
---
paths:
  "/list-sessions/{id}":
    delete:
      operationId: list-sessions
      parameters:
      - name: id
        type: string
        in: path
      - name: id
        type: string
        in: query
`),
		},
	}

	callCli([]string{"messages", "list-sessions", "--id", "9999", "--url", server.URL}, models)

	if urlPath != "/list-sessions/9999" {
		t.Errorf("Expected url path to contain id parameter, but got %s", urlPath)
	}
	if query != "id=9999" {
		t.Errorf("Expected query string to contain id parameter, but got %s", query)
	}
}

func TestShowsErrorResponseOnStdErr(t *testing.T) {
	server := reponseStub(500, `{"error":"an error occurred"}`)

	models := []model.Data{
		{
			Name: "messages",
			Content: []byte(`
---
paths:
  "/create-session":
    post:
      operationId: create
      parameters:
      - name: token
        type: string
        in: body
`),
		},
	}

	_, errWriter := callCli([]string{"messages", "create", "--url", server.URL}, models)

	expectedErrorOutput := `{
	"error": "an error occurred"
}
`
	if errWriter.String() != expectedErrorOutput {
		t.Errorf("Error output was wrong, got: %s, but expected: %s", errWriter.String(), expectedErrorOutput)
	}
}

func TestCallsIncludeSchemaRefParameters(t *testing.T) {
	var body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body = readerToString(r.Body)
	}))

	models := []model.Data{
		{
			Name: "messages",
			Content: []byte(`
---
paths:
  "/post-message":
    post:
      operationId: post-message
      parameters:
      - name: mydata
        in: body
        required: true
        schema:
          "$ref": "#/definitions/MyData"
definitions:
  MyData:
    properties:
      topic:
        type: string
      message:
        type: string
`),
		},
	}

	callCli([]string{"messages", "post-message", "--topic", "mytopic", "--message", "mymessage", "--url", server.URL}, models)

	if body != `{"message":"mymessage","topic":"mytopic"}` {
		t.Errorf("Expected body to contain message and topic, but got %s", body)
	}
}

func TestCallsIncludeRefParameters(t *testing.T) {
	var urlPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlPath = r.URL.Path
	}))

	models := []model.Data{
		{
			Name: "messages",
			Content: []byte(`
---
paths:
  "/v1/sessions/{token}":
    delete:
      operationId: delete-session
      parameters:
      - "$ref": "#/parameters/Token"
parameters:
  Token:
    in: path
    name: token
    description: Unique session ID
    type: string
    required: true
`),
		},
	}

	callCli([]string{"messages", "delete-session", "--token", "mytoken", "--url", server.URL}, models)

	if urlPath != "/v1/sessions/mytoken" {
		t.Errorf("Expected url to contain token, but got %s", urlPath)
	}
}

func TestCallsIncludeStringArrayParameters(t *testing.T) {
	var body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body = readerToString(r.Body)
	}))

	models := []model.Data{
		{
			Name: "messages",
			Content: []byte(`
---
paths:
  "/create-session":
    post:
      operationId: method-with-string-array
      parameters:
      - schema:
          type: array
          items:
            type: string
        name: names
        in: body
`),
		},
	}

	callCli([]string{"messages", "method-with-string-array", "--names", "name1,name2,name3", "--url", server.URL}, models)

	if body != `{"names":["name1","name2","name3"]}` {
		t.Errorf("Expected body to contain names array, but got %s", body)
	}
}

func TestMethodNameUsesDashCasedOperationId(t *testing.T) {
	models := []model.Data{
		{
			Name: "messages",
			Content: []byte(`
---
paths:
  "/create-session":
    post:
      operationId: myMethod
      parameters:
      - name: paramA
        type: string
        in: body
`),
		},
	}

	writer, _ := callCli([]string{"messages", "--help"}, models)

	if !strings.Contains(writer.String(), "my-method") {
		t.Errorf("Help output was wrong, got: %s, but expected to contain: %s.", writer.String(), "my-method")
	}
}

func TestMethodNameUsesLastPartOfOperationId(t *testing.T) {
	models := []model.Data{
		{
			Name: "messages",
			Content: []byte(`
---
paths:
  "/create-session":
    post:
      operationId: MyOperation.MyMethod1
      parameters:
      - name: paramA
        type: string
        in: body
`),
		},
	}

	writer, _ := callCli([]string{"messages", "--help"}, models)

	if !strings.Contains(writer.String(), "my-method") {
		t.Errorf("Help output was wrong, got: %s, but expected to contain: %s.", writer.String(), "my-method1")
	}
}

func TestMethodNameUsesPathWithoutOperationId(t *testing.T) {
	models := []model.Data{
		{
			Name: "messages",
			Content: []byte(`
---
paths:
  "/create-session":
    post:
      parameters:
      - name: paramA
        type: string
        in: body
`),
		},
	}

	writer, _ := callCli([]string{"messages", "--help"}, models)

	if !strings.Contains(writer.String(), "create-session") {
		t.Errorf("Help output was wrong, got: %s, but expected to contain: %s.", writer.String(), "create-session")
	}
}

func TestIgnoreWebSocketOperations(t *testing.T) {
	models := []model.Data{
		{
			Name: "messages",
			Content: []byte(`
---
paths:
  "/v1/websocket":
    post:
      parameters:
      - name: paramA
        type: string
        in: body
`),
		},
	}

	writer, _ := callCli([]string{"messages", "--help"}, models)

	if strings.Contains(writer.String(), "websocket") {
		t.Errorf("Help output was wrong, got: %s, but expected not to contain: %s.", writer.String(), "websocket")
	}
}

func TestUrlProvidedByModel(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	models := []model.Data{
		{
			Name: "messages",
			Content: []byte(`
---
host: "` + server.URL + `"
schemes:
- http
paths:
  "/create-session":
    get:
      operationId: create
`),
		},
	}

	callCli([]string{"messages", "create", "--url", server.URL}, models)

	if !called {
		t.Error("Expected url provided by the model to be called but it was not")
	}
}
