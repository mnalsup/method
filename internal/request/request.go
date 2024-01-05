package request

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

	"github.com/mnalsup/method/logging"
)

type RequestDefinition struct {
	Method  string            `yaml:"method"`
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
	BodyStr string            `yaml:"bodyStr"`
	Body    interface{}       `yaml:"body"`
	Files   []FileDefinition  `yaml:"files"`
}

type FileDefinition struct {
	RequestBodyPath string `yaml:"requestBodyPath"`
	FilePath        string `yaml:"filePath"`
}

type RequestResult struct {
	Body     []byte
	Response *http.Response
	Elapsed  time.Duration
}

func DoRequest(definition *RequestDefinition) (*RequestResult, error) {
	var req *http.Request
	log := logging.GetLogger()

	var reqUrl *url.URL
	var err error
	client := &http.Client{Timeout: 10 * time.Second}

	reqUrl, err = url.Parse(definition.URL)
	if err != nil {
		return nil, err
	}

	req = &http.Request{
		Method: definition.Method,
		URL:    reqUrl,
		Header: http.Header{},
	}

	for k, v := range definition.Headers {
		req.Header.Add(k, v)
	}

	if definition.Body != nil {
		switch definition.Headers["Content-Type"] {
		case "application/json":
			body, err := convert(definition.Body)
			if err != nil {
				return nil, err
			}
			rawBody, err := json.Marshal(body)
			if err != nil {
				panic(fmt.Errorf("unable to marshal definition.Body into json: %v", err))
			}
			req.Body = io.NopCloser(bytes.NewReader(rawBody))
		case "application/x-www-form-urlencoded":
			body, err := convert(definition.Body)
			if err != nil {
				return nil, err
			}
			data := url.Values{}
			switch bt := body.(type) {
			case map[string]interface{}:
				for k, v := range body.(map[string]interface{}) {
					data.Add(k, fmt.Sprintf("%v", v))
				}
			default:
				panic(fmt.Sprintf("unable to convert type %v to application/x-www-form-urlencoded data", bt))
			}
			req.Body = io.NopCloser(strings.NewReader(data.Encode()))
		// Write files
		case "multipart/form-data":
			payloadMap, err := convertToMap(definition.Body)
			if err != nil {
				return nil, err
			}
			payload := &bytes.Buffer{}
			writer := multipart.NewWriter(payload)
			// Write Fields
			for key, value := range payloadMap {
				err = writer.WriteField(key, value.(string))
				if err != nil {
					return nil, err
				}
			}
			// Write Files
			for _, fileDef := range definition.Files {
				fmt.Println(fileDef)
				file, err := os.Open(fileDef.FilePath)
				if err != nil {
					return nil, err
				}
				fileWriter, err := writer.CreateFormFile(fileDef.RequestBodyPath, filepath.Base(fileDef.FilePath))
				if err != nil {
					return nil, err
				}
				_, err = io.Copy(fileWriter, file)
				if err != nil {
					return nil, err
				}
				file.Close()
				writer.Close()
			}
			req.Body = io.NopCloser(bytes.NewReader(payload.Bytes()))
			req.Header.Set("Content-Type", writer.FormDataContentType())
		default:
			panic(fmt.Sprintf("No request body parser available for %s", definition.Headers["Content-Type"]))
		}
	} else {
		if definition.BodyStr == "" {
			req.Body = nil
		} else {
			req.Body = io.NopCloser(strings.NewReader(definition.BodyStr))
		}
	}
	if err != nil {
		return nil, err
	}

	var result *RequestResult

	log.Debugf("Making a request to url: %s", req.URL.String())

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return result, fmt.Errorf("@DoRequest unable to make request: %s", err.Error())
	}
	defer func() {
		if resp.Body != nil {
			resp.Body.Close()
		}
	}()
	elapsed := time.Since(start)
	result = &RequestResult{
		Elapsed:  elapsed,
		Response: resp,
		Body:     nil,
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}
	result.Body = body

	return result, nil
}

// Used to take in an unstructured yaml body and convert it into a nested set of
// maps with string keys
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

func convertToMap(i interface{}) (map[string]interface{}, error) {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			converted, err := convert(v)
			if err != nil {
				return nil, err
			}
			if err != nil {
				return nil, err
			}
			m2[k.(string)] = converted
		}
		return m2, nil
	default:
		return nil, fmt.Errorf("unable to convert unknown to map[string]interface{}")
	}
}
