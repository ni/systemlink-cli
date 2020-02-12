package unit_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ni/systemlink-cli/internal/model"
)

func TestValidatesRequiredParameters(t *testing.T) {
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
        required: true
`),
		},
	}

	_, errWriter := callCli([]string{"messages", "create", "--url", "http://no-op"}, models)

	if errWriter.String() != "Missing argument: --token\n" {
		t.Errorf("Error output was wrong, got: %s, but expected: %s", errWriter.String(), "Missing argument: --token")
	}
}

func TestValidatesRequiredDefinitionParameters(t *testing.T) {
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
        in: body
        required: true
        schema:
          "$ref": "#/definitions/Topic"
definitions:
  Topic:
    type: object
    required:
    - topic
    properties:
      topic:
        type: string
`),
		},
	}

	_, errWriter := callCli([]string{"messages", "create", "--url", "http://no-op"}, models)

	if errWriter.String() != "Missing argument: --topic\n" {
		t.Errorf("Error output was wrong, got: %s, but expected: %s", errWriter.String(), "Missing argument: --topic")
	}
}

func TestValidatesInvalidJsonSchemaRefParameter(t *testing.T) {
	models := []model.Data{
		{
			Name: "messages",
			Content: []byte(`
---
paths:
  "/publish-message":
    post:
      operationId: publish
      parameters:
      - name: message
        in: body
        required: true
        schema:
          "$ref": "#/definitions/Message"
definitions:
  Message:
    type: object
    properties:
      topic:
        "$ref": "#/definitions/Topic"
  Topic:
    type: object
    properties:
      name:
        type: string
`),
		},
	}

	_, errWriter := callCli([]string{"messages", "publish", "--topic", "{ invalid json {", "--url", "http://no-op"}, models)

	if errWriter.String() != "Invalid value for argument 'topic'\n" {
		t.Errorf("Error output was wrong, got: %s, but expected: %s", errWriter.String(), "Invalid value for argument 'topic'")
	}
}

var types = []struct {
	typeInfo      string
	value         string
	valid         bool
	expectedValue string
}{
	// valid inputs
	{"string", "hello", true, `{"value":"hello"}`},
	{"integer", "98", true, `{"value":98}`},
	{"boolean", "true", true, `{"value":true}`},
	{"number", "99", true, `{"value":99}`},
	{"number", "123.0", true, `{"value":123}`},
	{"number", "456.99999", true, `{"value":456.99999}`},
	// invalid inputs
	{"integer", "1.090", false, "Invalid value for argument 'value'\n"},
	{"boolean", "fa l se", false, "Invalid value for argument 'value'\n"},
	{"number", "invalid", false, "Invalid value for argument 'value'\n"},
}

func TestValidatesParameterTypes(t *testing.T) {
	for _, tt := range types {
		var body string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body = readerToString(r.Body)
		}))

		models := []model.Data{
			{
				Name: "tag",
				Content: []byte(`
---
paths:
  "/":
    post:
      operationId: method-with-simple-type
      parameters:
      - type: "` + tt.typeInfo + `"
        name: value
        in: body
`),
			},
		}

		_, errWriter := callCli([]string{"tag", "method-with-simple-type", "--url", server.URL, "--value", tt.value}, models)

		if tt.valid && body != tt.expectedValue {
			t.Errorf("Request body was wrong, got: %s, but expected: %s", body, tt.expectedValue)
		}

		if !tt.valid && errWriter.String() != tt.expectedValue {
			t.Errorf("Error output was wrong, got: %s, but expected: %s", errWriter.String(), tt.expectedValue)
		}
	}
}

var arrayTypes = []struct {
	typeInfo      string
	value         string
	valid         bool
	expectedValue string
}{
	// valid inputs
	{"string", "1,hello,3", true, `{"ids":["1","hello","3"]}`},
	{"string", "yes", true, `{"ids":["yes"]}`},
	{"integer", "1,2,3", true, `{"ids":[1,2,3]}`},
	{"integer", "1", true, `{"ids":[1]}`},
	{"boolean", "true,false,true", true, `{"ids":[true,false,true]}`},
	{"boolean", "false", true, `{"ids":[false]}`},
	{"number", "1,5.0,999.99999", true, `{"ids":[1,5,999.99999]}`},
	{"number", "99", true, `{"ids":[99]}`},
	// invalid inputs
	{"integer", "1,invalid,3", false, "Invalid value for argument 'ids'\n"},
	{"boolean", "false,fa lse,true", false, "Invalid value for argument 'ids'\n"},
	{"number", "invalid", false, "Invalid value for argument 'ids'\n"},
}

func TestValidatesArrayParameterTypes(t *testing.T) {
	for _, tt := range arrayTypes {
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
  "/":
    post:
      operationId: method-with-array
      parameters:
      - schema:
          type: array
          items:
            type: "` + tt.typeInfo + `"
        name: ids
        in: body
`),
			},
		}

		_, errWriter := callCli([]string{"messages", "method-with-array", "--ids", tt.value, "--url", server.URL}, models)

		if tt.valid && body != tt.expectedValue {
			t.Errorf("Request body was wrong, got: %s, but expected: %s", body, tt.expectedValue)
		}

		if !tt.valid && errWriter.String() != tt.expectedValue {
			t.Errorf("Error output was wrong, got: %s, but expected: %s", errWriter.String(), tt.expectedValue)
		}
	}
}

func TestValidatesParametersWithSameNameHaveSameType(t *testing.T) {
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
        type: integer
        in: query
`),
		},
	}

	_, errWriter := callCli([]string{"messages", "list-sessions", "--id", "my-session"}, models)

	if errWriter.String() != "Invalid value for argument 'id'\n" {
		t.Errorf("Error output was wrong, got: %s, but expected: %s", errWriter.String(), "Invalid value for argument 'id'")
	}
}
