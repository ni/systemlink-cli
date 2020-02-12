package unit_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/ni/systemlink-cli/internal/commandline"
	"github.com/ni/systemlink-cli/internal/model"
	"github.com/ni/systemlink-cli/internal/niservice"
	"github.com/ni/systemlink-cli/internal/parser"
)

func createCli(configData string) (commandline.CLI, *bytes.Buffer, *bytes.Buffer) {
	writer := new(bytes.Buffer)
	errWriter := new(bytes.Buffer)
	config, err := commandline.NewConfig([]byte(configData), "/home")
	if err != nil {
		fmt.Fprintln(errWriter, "Error reading config:", err)
	}
	c := commandline.CLI{
		Parser:    parser.SwaggerParser{},
		Service:   niservice.NIService{},
		Writer:    writer,
		ErrWriter: errWriter,
		Config:    config,
	}
	return c, writer, errWriter
}

func callCliWithConfig(args []string, models []model.Data, config string) (*bytes.Buffer, *bytes.Buffer) {
	args = append([]string{"systemlink"}, args...)
	c, writer, errWriter := createCli(config)
	c.Exec(args, models)
	return writer, errWriter
}

func callCli(args []string, models []model.Data) (*bytes.Buffer, *bytes.Buffer) {
	return callCliWithConfig(args, models, "")
}

func reponseStub(statusCode int, content string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte(content))
	}))
}

func successReponseStub(content string) *httptest.Server {
	return reponseStub(http.StatusOK, content)
}

func readerToString(reader io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	return buf.String()
}
