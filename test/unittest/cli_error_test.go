package unit_test

import (
	"strings"
	"testing"

	"github.com/ni/systemlink-cli/internal/model"
)

func TestInvalidModel(t *testing.T) {
	models := []model.Data{
		{
			Name: "messages",
			Content: []byte(
				`=== INVALID ===`),
		},
	}

	_, errWriter := callCli([]string{"messages"}, models)

	errorOutput := "Error parsing model 'messages'"
	if !strings.Contains(errWriter.String(), errorOutput) {
		t.Errorf("Error output was wrong, got: %s, but expected to contain: %s.", errWriter.String(), errorOutput)
	}
}

func TestInvalidModelExpansion(t *testing.T) {
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
          "$ref": "#/definitions/INVALID"
`),
		},
	}

	_, errWriter := callCli([]string{"messages"}, models)

	errorOutput := "Error parsing model 'messages'"
	if !strings.Contains(errWriter.String(), errorOutput) {
		t.Errorf("Error output was wrong, got: %s, but expected to contain: %s.", errWriter.String(), errorOutput)
	}
}

func TestInvalidConfigFile(t *testing.T) {
	config := `INVALID YAML`

	_, errWriter := callCliWithConfig([]string{"messages", "create"}, configDefaultModels, config)

	errorOutput := "Error reading yaml"
	if !strings.Contains(errWriter.String(), errorOutput) {
		t.Errorf("Error output was wrong, got: %s, but expected to contain: %s.", errWriter.String(), errorOutput)
	}
}

func TestInvalidParameterType(t *testing.T) {
	models := []model.Data{
		{
			Name: "messages",
			Content: []byte(`
---
paths:
  "/create":
    post:
      operationId: create
      parameters:
      - name: id
        type: INVALID
        in: body
`),
		},
	}

	_, errWriter := callCli([]string{"messages", "create"}, models)

	errorOutput := "Invalid type 'INVALID'"
	if !strings.Contains(errWriter.String(), errorOutput) {
		t.Errorf("Error output was wrong, got: %s, but expected to contain: %s.", errWriter.String(), errorOutput)
	}
}

func TestInvalidParameterArrayType(t *testing.T) {
	models := []model.Data{
		{
			Name: "messages",
			Content: []byte(`
---
paths:
  "/create":
    post:
      operationId: create
      parameters:
      - name: id
        schema:
          type: array
          items:
            type: INVALID_ARR
        in: body
`),
		},
	}

	_, errWriter := callCli([]string{"messages", "create"}, models)

	errorOutput := "Invalid array type 'INVALID_ARR'"
	if !strings.Contains(errWriter.String(), errorOutput) {
		t.Errorf("Error output was wrong, got: %s, but expected to contain: %s.", errWriter.String(), errorOutput)
	}
}

func TestInvalidParameterLocation(t *testing.T) {
	models := []model.Data{
		{
			Name: "messages",
			Content: []byte(`
---
paths:
  "/create":
    post:
      operationId: create
      parameters:
      - name: id
        type: string
        in: INVALID
`),
		},
	}

	_, errWriter := callCli([]string{"messages", "create"}, models)

	errorOutput := "Invalid location 'INVALID'"
	if !strings.Contains(errWriter.String(), errorOutput) {
		t.Errorf("Error output was wrong, got: %s, but expected to contain: %s.", errWriter.String(), errorOutput)
	}
}

func TestInvalidTypeInDefinition(t *testing.T) {
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
      message:
        type: INVALID
`),
		},
	}

	_, errWriter := callCli([]string{"messages", "create"}, models)

	errorOutput := "Invalid type 'INVALID'"
	if !strings.Contains(errWriter.String(), errorOutput) {
		t.Errorf("Error output was wrong, got: %s, but expected to contain: %s.", errWriter.String(), errorOutput)
	}
}
