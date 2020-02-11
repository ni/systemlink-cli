package unit_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/ni/systemlink-cli/internal/commandline"
	"github.com/ni/systemlink-cli/internal/model"
	"github.com/ni/systemlink-cli/internal/parser"
)

type fakeService struct {
	operation       model.Operation
	parameterValues []model.ParameterValue
	settings        model.Settings
}

func (s *fakeService) Call(operation model.Operation, parameterValues []model.ParameterValue, settings model.Settings) (int, string, error) {
	s.operation = operation
	s.parameterValues = parameterValues
	s.settings = settings
	return 200, "", nil
}

func createCliWithFakeService(config string) (commandline.CLI, *bytes.Buffer, *bytes.Buffer, *fakeService) {
	writer := new(bytes.Buffer)
	errWriter := new(bytes.Buffer)
	service := fakeService{}
	c := commandline.CLI{
		Parser:    parser.SwaggerParser{},
		Service:   &service,
		Writer:    writer,
		ErrWriter: errWriter,
		Config:    commandline.NewConfig([]byte(config), "/home"),
	}
	return c, writer, errWriter, &service
}

func callCliWithFakeService(args []string, models []model.Data, config string) (*bytes.Buffer, *bytes.Buffer, *fakeService) {
	args = append([]string{"systemlink"}, args...)
	c, writer, errWriter, service := createCliWithFakeService(config)
	c.Exec(args, models)
	return writer, errWriter, service
}

var callDefaultModels = []model.Data{
	{
		Name:    "messages",
		Content: []byte(`{ "paths": { "/create-session": { "get": { "operationId": "create" } } } }`),
	},
}

func TestCLICallsServiceWithDefaultConfigParameters(t *testing.T) {
	config := `
profiles:
  - name: default
    api-key: my-default-api-key
    url: http://localhost:1234
    ssh-proxy: ubuntu@1.2.3.4:22
    ssh-key: /home/user/key.pem
    ssh-known-host: "my-host-key"
    insecure: true
    verbose: true`

	_, _, service := callCliWithFakeService([]string{"messages", "create"}, callDefaultModels, config)

	expectedSettings := model.Settings{
		APIKey:       "my-default-api-key",
		Verbose:      true,
		URL:          "http://localhost:1234",
		Insecure:     true,
		SSHProxy:     "ubuntu@1.2.3.4:22",
		SSHKey:       `/home/user/key.pem`,
		SSHKnownHost: "my-host-key",
	}
	if !reflect.DeepEqual(service.settings, expectedSettings) {
		t.Errorf("Different settings than expected in service call, got: %v, but expected %v", service.settings, expectedSettings)
	}
}

func TestCLICallsServiceWithNamedProfileConfigParameters(t *testing.T) {
	config := `
profiles:
  - name: other-profile
    api-key: other-api-key
  - name: my-profile
    api-key: my-profile-api-key
    url: http://localhost:1234
    ssh-proxy: ubuntu@1.2.3.4:22
    ssh-key: /home/user/key.pem
    ssh-known-host: "my-host-key"
    insecure: true
    verbose: true`

	_, _, service := callCliWithFakeService([]string{"messages", "create", "--profile", "my-profile"}, callDefaultModels, config)

	expectedSettings := model.Settings{
		APIKey:       "my-profile-api-key",
		Verbose:      true,
		URL:          "http://localhost:1234",
		Insecure:     true,
		SSHProxy:     "ubuntu@1.2.3.4:22",
		SSHKey:       `/home/user/key.pem`,
		SSHKnownHost: "my-host-key",
	}
	if !reflect.DeepEqual(service.settings, expectedSettings) {
		t.Errorf("Different settings than expected in service call, got: %v, but expected %v", service.settings, expectedSettings)
	}
}

var resolvedPathsTest = []struct {
	path         string
	resolvedPath string
}{
	{"key.pem", "/home/key.pem"},
	{"../test/key.pem", "/test/key.pem"},
	{"../../test/key.pem", "/test/key.pem"},
	{"/other/key.pem", "/other/key.pem"},
}

func TestCLICallsServiceWithResolvedPaths(t *testing.T) {
	for _, tt := range resolvedPathsTest {
		config := `
profiles:
  - name: default
    ssh-key: ` + tt.path

		_, _, service := callCliWithFakeService([]string{"messages", "create"}, callDefaultModels, config)

		expectedSettings := model.Settings{
			URL:    "https://api.systemlinkcloud.com",
			SSHKey: tt.resolvedPath,
		}
		if !reflect.DeepEqual(service.settings, expectedSettings) {
			t.Errorf("Different settings than expected in service call, got: %v, but expected %v", service.settings, expectedSettings)
		}
	}
}
