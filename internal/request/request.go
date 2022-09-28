package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	client := &http.Client{}

	reqUrl, err = url.Parse(definition.URL)
	if err != nil {
		return nil, err
	}

	req = &http.Request{
		Method: definition.Method,
		URL:    reqUrl,
		Header: http.Header{},
	}

	if definition.Body != nil {
		switch definition.Headers["Content-Type"] {
		case "application/json":
			body := convert(definition.Body)
			rawBody, err := json.Marshal(body)
			if err != nil {
				panic(fmt.Errorf("unable to marshal definition.Body into json: %v", err))
			}
			req.Body = io.NopCloser(bytes.NewReader(rawBody))
		case "application/x-www-form-urlencoded":
			body := convert(definition.Body)
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
	for k, v := range definition.Headers {
		req.Header.Add(k, v)
	}

	var result *RequestResult

	log.Debugf("Making a request to url: %s", req.URL.String())

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)
	result = &RequestResult{
		Elapsed:  elapsed,
		Response: resp,
		Body:     nil,
	}
	if err != nil {
		return result, fmt.Errorf("@DoRequest unable to make request: %s", err.Error())
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}
	resp.Body.Close()
	resp.Body = nil
	result.Body = body

	return result, nil
}

// Used to take in an unstructured yaml body and convert it into a nested set of
// maps with string keys
func convert(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convert(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convert(v)
		}
	}
	return i
}
