package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mnalsup/method/internal"
	"github.com/mnalsup/method/logging"
)

// RequestRunner implements the internal.RequestRunner interface
type RequestRunner struct {
	schemaLoader internal.SchemaLoader
}

// NewRequestRunner creates a new RequestRunner
func NewRequestRunner(schemaLoader internal.SchemaLoader) *RequestRunner {
	return &RequestRunner{schemaLoader: schemaLoader}
}

// RunRequest executes a request file and returns the result
func (r *RequestRunner) RunRequest(path string) (internal.RequestResult, error) {
	schema, err := r.schemaLoader.LoadSchema(path, false)
	if err != nil {
		return internal.RequestResult{}, fmt.Errorf("failed to load schema %s: %w", path, err)
	}

	// Initialize headers if nil
	if schema.RequestDefinition.Headers == nil {
		schema.RequestDefinition.Headers = make(map[string]string)
	}

	// Execute the request directly without value resolution
	// Value resolution should be handled at a higher level
	result, err := doRequest(&schema.RequestDefinition)
	if err != nil {
		return internal.RequestResult{}, fmt.Errorf("failed to execute request: %w", err)
	}

	return *result, nil
}

// doRequest performs the actual HTTP request
func doRequest(definition *internal.RequestDefinition) (*internal.RequestResult, error) {
	var req *http.Request
	log := logging.GetLogger()

	// Set timeout - default to 10 seconds if not specified
	timeout := 10 * time.Second
	if definition.TimeoutSeconds > 0 {
		timeout = time.Duration(definition.TimeoutSeconds) * time.Second
	}

	client := &http.Client{Timeout: timeout}

	reqUrl, err := url.Parse(definition.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	req = &http.Request{
		Method: definition.Method,
		URL:    reqUrl,
		Header: http.Header{},
	}

	// Add headers
	for k, v := range definition.Headers {
		req.Header.Add(k, v)
	}

	// Handle request body
	err = setRequestBody(req, definition)
	if err != nil {
		return nil, fmt.Errorf("failed to set request body: %w", err)
	}

	log.Debugf("Making request to: %s", req.URL.String())

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if resp.Body != nil {
			resp.Body.Close()
		}
	}()

	elapsed := time.Since(start)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &internal.RequestResult{
			Elapsed:  elapsed,
			Response: resp,
			Body:     nil,
		}, fmt.Errorf("failed to read response body: %w", err)
	}

	return &internal.RequestResult{
		Body:     body,
		Response: resp,
		Elapsed:  elapsed,
	}, nil
}

// setRequestBody configures the request body based on content type
func setRequestBody(req *http.Request, definition *internal.RequestDefinition) error {
	if definition.Body != nil {
		contentType := definition.Headers["Content-Type"]
		switch contentType {
		case "application/json":
			return setJSONBody(req, definition.Body)
		case "application/x-www-form-urlencoded":
			return setFormBody(req, definition.Body)
		case "multipart/form-data":
			return setMultipartBody(req, definition)
		default:
			return fmt.Errorf("unsupported content type: %s", contentType)
		}
	} else if definition.BodyStr != "" {
		req.Body = io.NopCloser(strings.NewReader(definition.BodyStr))
	}

	return nil
}

// setJSONBody sets JSON request body
func setJSONBody(req *http.Request, body interface{}) error {
	converted, err := convert(body)
	if err != nil {
		return fmt.Errorf("failed to convert body: %w", err)
	}

	rawBody, err := json.Marshal(converted)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	req.Body = io.NopCloser(bytes.NewReader(rawBody))
	return nil
}

// setFormBody sets form-encoded request body
func setFormBody(req *http.Request, body interface{}) error {
	converted, err := convert(body)
	if err != nil {
		return fmt.Errorf("failed to convert body: %w", err)
	}

	data := url.Values{}
	bodyMap, ok := converted.(map[string]interface{})
	if !ok {
		return fmt.Errorf("body must be a map for form encoding")
	}

	for k, v := range bodyMap {
		data.Add(k, fmt.Sprintf("%v", v))
	}

	req.Body = io.NopCloser(strings.NewReader(data.Encode()))
	return nil
}

// setMultipartBody sets multipart form data request body
func setMultipartBody(req *http.Request, definition *internal.RequestDefinition) error {
	payloadMap, err := convertToMap(definition.Body)
	if err != nil {
		return fmt.Errorf("failed to convert body to map: %w", err)
	}

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	// Write form fields
	for key, value := range payloadMap {
		err = writer.WriteField(key, fmt.Sprintf("%v", value))
		if err != nil {
			return fmt.Errorf("failed to write field %s: %w", key, err)
		}
	}

	// Write files
	for _, fileDef := range definition.Files {
		file, err := os.Open(fileDef.FilePath)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", fileDef.FilePath, err)
		}

		fileWriter, err := writer.CreateFormFile(fileDef.RequestBodyPath, filepath.Base(fileDef.FilePath))
		if err != nil {
			file.Close()
			return fmt.Errorf("failed to create form file: %w", err)
		}

		_, err = io.Copy(fileWriter, file)
		file.Close()
		if err != nil {
			return fmt.Errorf("failed to copy file content: %w", err)
		}
	}

	writer.Close()
	req.Body = io.NopCloser(bytes.NewReader(payload.Bytes()))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return nil
}

// convert transforms YAML-parsed data into proper Go types
func convert(i interface{}) (interface{}, error) {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		return convertToMap(x)
	case []interface{}:
		for i, v := range x {
			converted, err := convert(v)
			if err != nil {
				return nil, err
			}
			x[i] = converted
		}
		return x, nil
	default:
		return i, nil
	}
}

// convertToMap converts map[interface{}]interface{} to map[string]interface{}
func convertToMap(i interface{}) (map[string]interface{}, error) {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			converted, err := convert(v)
			if err != nil {
				return nil, err
			}
			keyStr, ok := k.(string)
			if !ok {
				return nil, fmt.Errorf("non-string key found in map: %v", k)
			}
			m2[keyStr] = converted
		}
		return m2, nil
	default:
		return nil, fmt.Errorf("unable to convert %T to map[string]interface{}", i)
	}
}
