package niservice

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	uuid "github.com/nu7hatch/gouuid"

	"github.com/ni/systemlink-cli/internal/model"
	"github.com/ni/systemlink-cli/internal/ssh"
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

func (s NIService) prepareFormData(parameterValues []model.ParameterValue) (string, []byte, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	for _, paramValue := range parameterValues {
		file, err := os.Open(paramValue.Value.(string))
		if err != nil {
			return "", nil, err
		}
		fw, err := w.CreateFormFile(paramValue.Name, file.Name())
		if err != nil {
			return "", nil, err
		}
		_, err = io.Copy(fw, file)
		if err != nil {
			return "", nil, err
		}
	}
	w.Close()

	return "multipart/form-data; boundary=" + w.Boundary(), b.Bytes(), nil
}

func (s NIService) prepareJSON(parameterValues []model.ParameterValue) (string, []byte, error) {
	var body = map[string]interface{}{}

	for _, paramValue := range parameterValues {
		body[paramValue.Name] = paramValue.Value
	}

	json, err := json.Marshal(body)
	return "application/json", json, err
}

func (s NIService) prepareBody(parameterValues []model.ParameterValue) (string, []byte, error) {
	jsonParameterValues := s.filterParameterValues(model.BodyLocation, parameterValues)
	if len(jsonParameterValues) > 0 {
		return s.prepareJSON(jsonParameterValues)
	}

	formParamValues := s.filterParameterValues(model.FormDataLocation, parameterValues)
	if len(formParamValues) > 0 {
		return s.prepareFormData(formParamValues)
	}

	return "", []byte{}, nil
}

func (s NIService) prepareHeader(parameterValues []model.ParameterValue) map[string]string {
	var header = map[string]string{}

	var paramValues = s.filterParameterValues(model.HeaderLocation, parameterValues)
	for _, paramValue := range paramValues {
		header[paramValue.Name] = paramValue.Value.(string)
	}
	return header
}

func (s NIService) dumpRequest(req *http.Request) (string, error) {
	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return "", err
	}
	return string(dump), nil
}

func (s NIService) dumpResponse(resp *http.Response) (string, error) {
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return "", err
	}
	return string(dump), nil
}

func (s NIService) convertBytesToJSONString(value []byte) string {
	var prettyJSON bytes.Buffer
	error := json.Indent(&prettyJSON, value, "", "\t")
	if error != nil {
		return string(value)
	}
	return prettyJSON.String()
}

func (s NIService) startProxy(settings model.Settings) (*url.URL, error) {
	sshConfig, err := ssh.NewConfig(settings.SSHProxy, settings.SSHKey, settings.SSHKnownHost)
	if sshConfig == nil || err != nil {
		return nil, err
	}

	proxy := &ssh.HTTPOverSSHProxy{}
	proxyURL, err := proxy.Start(*sshConfig)
	if err != nil {
		return nil, err
	}
	return url.Parse("http://" + proxyURL)
}

func (s NIService) newHTTPCLient(insecure bool, proxyURL *url.URL) *http.Client {
	transport := &http.Transport{
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: insecure},
	}
	if proxyURL != nil {
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	return &http.Client{Transport: transport}
}

func (s NIService) send(client *http.Client, req *http.Request) (*http.Response, error) {
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
		return nil, err
	}
	return result.(*http.Response), nil
}

func (s NIService) newRequestID() string {
	u, err := uuid.NewV4()
	if err != nil {
		return strconv.Itoa(rand.Intn(math.MaxInt32))
	}
	return u.String()
}

func (s NIService) newRequest(
	client *http.Client,
	operation model.Operation,
	parameterValues []model.ParameterValue,
	settings model.Settings) (*http.Request, string, error) {
	serviceURL := s.prepareURL(settings.URL, operation, parameterValues)
	contentType, body, err := s.prepareBody(parameterValues)
	if err != nil {
		return nil, "", err
	}
	req, err := http.NewRequest(operation.Method, serviceURL, bytes.NewReader(body))
	if err != nil {
		return nil, "", err
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
	if contentType != "" {
		req.Header.Add("content-type", contentType)
	}

	output := ""
	if settings.Verbose {
		requestOutput, err := s.dumpRequest(req)
		if err != nil {
			return nil, "", err
		}
		output = requestOutput + "\n"
	}
	return req, output, nil
}

func (s NIService) readResponse(resp *http.Response, verbose bool) (int, string, error) {
	if verbose {
		responseOutput, err := s.dumpResponse(resp)
		return resp.StatusCode, responseOutput, err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	output := ""
	if err == nil {
		output = s.convertBytesToJSONString(bodyBytes)
	}
	return resp.StatusCode, output, err
}

// Call is instantiating a new HTTP client, prepares the request object
// and sends a message to the target service
// The response is parsed and returned to the caller.
func (s NIService) Call(
	operation model.Operation,
	parameterValues []model.ParameterValue,
	settings model.Settings) (int, string, error) {
	proxyURL, err := s.startProxy(settings)
	if err != nil {
		return 0, "", NewServiceError("Error starting proxy", err)
	}
	client := s.newHTTPCLient(settings.Insecure, proxyURL)

	req, output, err := s.newRequest(client, operation, parameterValues, settings)
	if err != nil {
		return 0, "", NewServiceError("Error creating request", err)
	}

	resp, err := s.send(client, req)
	if err != nil {
		return 0, "", NewServiceError("Error sending request", err)
	}

	statusCode, responseOutput, err := s.readResponse(resp, settings.Verbose)
	output = output + responseOutput
	if err != nil {
		return statusCode, output, NewServiceError("Error receiving response", err)
	}

	if statusCode >= 400 {
		err = errors.New(output)
	}
	return statusCode, output, err
}
