package niservice

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	uuid "github.com/nu7hatch/gouuid"

	"github.com/ni/systemlink-cli/internal/model"
)

// NIService struct is taking the parsed model, all input parameters and settings
// and creates a new HTTP request with all HTTP headers, url and body parameters set
// and sends it to SystemLink web service
type NIService struct{}

func (s NIService) prepareQueryString(parameterValues []model.ParameterValue) string {
	var queryString []string

	var paramValues = s.filterParameterValues(model.QueryLocation, parameterValues)
	for _, paramValue := range paramValues {
		queryString = append(queryString, paramValue.Name+"="+paramValue.Value.(string))
	}
	if len(queryString) > 0 {
		return "?" + strings.Join(queryString, "&")
	}
	return ""
}

func (s NIService) prepareURL(baseURL string, operation model.Operation, parameterValues []model.ParameterValue) string {
	url := baseURL + operation.Path
	var paramValues = s.filterParameterValues(model.PathLocation, parameterValues)
	for _, paramValue := range paramValues {
		url = strings.Replace(url, "{"+paramValue.Name+"}", paramValue.Value.(string), -1)
	}
	queryString := s.prepareQueryString(parameterValues)
	return url + queryString
}

func (s NIService) filterParameterValues(location model.ParameterLocation, parameterValues []model.ParameterValue) []model.ParameterValue {
	var result []model.ParameterValue

	for _, paramValue := range parameterValues {
		if paramValue.Location == location {
			result = append(result, paramValue)
		}
	}

	return result
}

func (s NIService) prepareBody(parameterValues []model.ParameterValue) ([]byte, error) {
	var body = map[string]interface{}{}

	var paramValues = s.filterParameterValues(model.BodyLocation, parameterValues)
	for _, paramValue := range paramValues {
		body[paramValue.Name] = paramValue.Value
	}
	if len(body) > 0 {
		var json, err = json.Marshal(body)
		return json, err
	}
	return []byte{}, nil
}

func (s NIService) prepareHeader(parameterValues []model.ParameterValue) map[string]string {
	var header = map[string]string{}

	var paramValues = s.filterParameterValues(model.HeaderLocation, parameterValues)
	for _, paramValue := range paramValues {
		header[paramValue.Name] = paramValue.Value.(string)
	}
	return header
}

func (s NIService) dumpRequest(req *http.Request) string {
	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		panic(err)
	}
	return string(dump)
}

func (s NIService) dumpResponse(resp *http.Response) string {
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		panic(err)
	}
	return string(dump)
}

func (s NIService) convertBytesToJSONString(value []byte) string {
	var prettyJSON bytes.Buffer
	error := json.Indent(&prettyJSON, value, "", "\t")
	if error != nil {
		return string(value)
	}
	return prettyJSON.String()
}

func (s NIService) startProxy(sshConfig SSHConfig) string {
	proxy := &HTTPOverSSHProxy{}
	localProxyURL, err := proxy.Start(sshConfig)
	if err != nil {
		panic(err)
	}
	return localProxyURL
}

func (s NIService) newHTTPCLient(insecure bool, sshConfig *SSHConfig) *http.Client {
	transport := &http.Transport{
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: insecure},
	}

	if sshConfig != nil {
		localProxyURL, _ := url.Parse("http://" + s.startProxy(*sshConfig))
		transport.Proxy = http.ProxyURL(localProxyURL)
	}

	return &http.Client{Transport: transport}
}

func (s NIService) send(client *http.Client, req *http.Request) *http.Response {
	result, err := retry(3, time.Second, func() (interface{}, error) {
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode >= 500 {
			return resp, fmt.Errorf("Server Error: %s", resp.Status)
		}
		return resp, nil
	})
	if result == nil && err != nil {
		panic(err)
	}
	return result.(*http.Response)
}

func (s NIService) createSSHConfig(settings model.Settings) *SSHConfig {
	if settings.SSHProxy == "" {
		return nil
	}

	url, err := url.Parse("//" + settings.SSHProxy)
	if err != nil {
		panic(err)
	}
	username := "ubuntu"
	if url.User != nil {
		username = url.User.Username()
	}

	return &SSHConfig{
		HostName:  url.Host,
		KeyFile:   settings.SSHKey,
		KnownHost: settings.SSHKnownHost,
		UserName:  username,
	}
}

func (s NIService) newRequestID() string {
	u, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return u.String()
}

// Call is instantiating a new HTTP client, prepares the request object
// and sends a message to the target service
// The response is parsed and returned to the caller.
func (s NIService) Call(
	operation model.Operation,
	parameterValues []model.ParameterValue,
	settings model.Settings) (int, string, error) {
	sshConfig := s.createSSHConfig(settings)
	client := s.newHTTPCLient(settings.Insecure, sshConfig)

	serviceURL := s.prepareURL(settings.URL, operation, parameterValues)
	body, err := s.prepareBody(parameterValues)
	if err != nil {
		return 0, "", err
	}
	req, err := http.NewRequest(operation.Method, serviceURL, bytes.NewReader(body))
	if err != nil {
		panic(err)
	}

	headers := s.prepareHeader(parameterValues)
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	if settings.APIKey != "" {
		req.Header.Add("x-ni-api-key", settings.APIKey)
	}
	if settings.Username != "" || settings.Password != "" {
		req.SetBasicAuth(settings.Username, settings.Password)
	}
	req.Header.Add("x-request-id", s.newRequestID())
	if len(body) > 0 {
		req.Header.Add("content-type", "application/json")
	}

	output := ""
	if settings.Verbose {
		requestOutput := s.dumpRequest(req)
		output = output + requestOutput + "\n"
	}

	resp := s.send(client, req)

	if settings.Verbose {
		responseOutput := s.dumpResponse(resp)
		output = output + responseOutput
	} else {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		output = s.convertBytesToJSONString(bodyBytes)
	}

	if resp.StatusCode >= 400 {
		err = errors.New(output)
	}
	return resp.StatusCode, output, err
}
