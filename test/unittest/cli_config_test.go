package unit_test

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ni/systemlink-cli/internal/model"
)

var configDefaultModels = []model.Data{
	{
		Name:    "messages",
		Content: []byte(`{ "paths": { "/create-session": { "get": { "operationId": "create" } } } }`),
	},
}

var defaultConfig = `
profiles:
  - name: default
    api-key: my-default-api-key`

func TestApiKeyFromConfigIsAddedToHttpHeader(t *testing.T) {
	var apiKeyHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKeyHeader = r.Header.Get("x-ni-api-key")
	}))

	callCliWithConfig([]string{"messages", "create", "--url", server.URL}, configDefaultModels, defaultConfig)

	if apiKeyHeader != "my-default-api-key" {
		t.Errorf("API key not found in HTTP header, got: %s, but expected %s", apiKeyHeader, "my-default-api-key")
	}
}

func TestBasicAuthFromConfigIsAddedToHttpHeader(t *testing.T) {
	var config = `
profiles:
  - name: default
    username: MyUser
    password: MyPassword`

	var basicAuthHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		basicAuthHeader = r.Header.Get("Authorization")
	}))

	callCliWithConfig([]string{"messages", "create", "--url", server.URL}, configDefaultModels, config)

	encoded := "Basic " + base64.StdEncoding.EncodeToString([]byte("MyUser:MyPassword"))
	if basicAuthHeader != encoded {
		t.Errorf("Basic Auth not found in HTTP header, got: %s, but expected %s", basicAuthHeader, encoded)
	}
}

func TestBasicAuthUsernameFromConfigIsAddedToHttpHeader(t *testing.T) {
	var config = `
profiles:
  - name: default
    username: MyUser`

	var basicAuthHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		basicAuthHeader = r.Header.Get("Authorization")
	}))

	callCliWithConfig([]string{"messages", "create", "--url", server.URL}, configDefaultModels, config)

	encoded := "Basic " + base64.StdEncoding.EncodeToString([]byte("MyUser:"))
	if basicAuthHeader != encoded {
		t.Errorf("Basic Auth not found in HTTP header, got: %s, but expected %s", basicAuthHeader, encoded)
	}
}

func TestBasicAuthPasswordFromConfigIsAddedToHttpHeader(t *testing.T) {
	var config = `
profiles:
  - name: default
    password: my-password`

	var basicAuthHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		basicAuthHeader = r.Header.Get("Authorization")
	}))

	callCliWithConfig([]string{"messages", "create", "--url", server.URL}, configDefaultModels, config)

	encoded := "Basic " + base64.StdEncoding.EncodeToString([]byte(":my-password"))
	if basicAuthHeader != encoded {
		t.Errorf("Basic Auth not found in HTTP header, got: %s, but expected %s", basicAuthHeader, encoded)
	}
}

func TestGetsConfigEntriesFromDefaultProfile(t *testing.T) {
	var apiKeyHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKeyHeader = r.Header.Get("x-ni-api-key")
	}))

	callCliWithConfig([]string{"messages", "create", "--url", server.URL}, configDefaultModels, defaultConfig)

	if apiKeyHeader != "my-default-api-key" {
		t.Errorf("Did not call URL from default profile, got: %s, but expected %s", apiKeyHeader, "my-default-api-key")
	}
}

func TestArgumentOverridesConfigValue(t *testing.T) {
	var apiKeyHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKeyHeader = r.Header.Get("x-ni-api-key")
	}))

	config := `
profiles:
  - name: my-profile
    api-key: my-profile-api-key`

	callCliWithConfig([]string{"messages", "create", "--api-key", "my-argument-api-key", "--profile", "my-profile", "--url", server.URL}, configDefaultModels, config)

	if apiKeyHeader != "my-argument-api-key" {
		t.Errorf("API key not found in HTTP header, got: %s, but expected %s", apiKeyHeader, "my-api-key")
	}
}

func TestSelectsSpecifiedProfile(t *testing.T) {
	var apiKeyHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKeyHeader = r.Header.Get("x-ni-api-key")
	}))

	config := `
profiles:
  - name: default
    api-key: default-key
  - name: my-profile
    api-key: key-from-profile`

	callCliWithConfig([]string{"messages", "create", "--profile", "my-profile", "--url", server.URL}, configDefaultModels, config)

	if apiKeyHeader != "key-from-profile" {
		t.Errorf("API key not found in HTTP header, got: %s, but expected %s", apiKeyHeader, "my-api-key")
	}
}

func TestIgnoresProfileIfNotExist(t *testing.T) {
	var apiKeyHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKeyHeader = r.Header.Get("x-ni-api-key")
	}))

	config := `
profiles:
  - name: my-profile
    api-key: key-from-profile`

	callCliWithConfig([]string{"messages", "create", "--profile", "other-profile", "--url", server.URL}, configDefaultModels, config)

	if apiKeyHeader != "" {
		t.Errorf("API key found in HTTP header, got: %s, but expected empty value", apiKeyHeader)
	}
}
