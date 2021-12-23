package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Jeffail/gabs/v2"
	"gopkg.in/yaml.v2"
)

const (
	AUTH_TYPE_BEARER_TOKEN = "BearerToken"
)

type AuthorizationHook struct {
	OnHttpStatus      []int  `yaml:"onHttpStatus"`
	RequestPath       string `yaml:"requestPath"`
	JsonParseBodyPath string `yaml:"jsonParseBodyPath"`
	AuthType          string `yaml:"authType"`
}

type RequestDefinition struct {
	Method            string             `yaml:"method"`
	URL               string             `yaml:"url"`
	Headers           map[string]string  `yaml:"headers"`
	AuthorizationHook *AuthorizationHook `yaml:"authorizationHook"`
	BodyStr           string             `yaml:"bodyStr"`
	Body              interface{}        `yaml:"body"`
}

type RequestResult struct {
	Body     []byte
	Response *http.Response
	Elapsed  time.Duration
}

func contains(elems []int, v int) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}

func main() {
	args := os.Args[1:]

	fileName := args[0]
	definition, err := readRequestDefinition(fileName)
	if err != nil {
		panic(err.Error())
	}

	result, err := DoMethod(definition)
	if err != nil {
		panic(err.Error())
	}

	printRequestResult(result)
}

func printRequestResult(result *RequestResult) {
	fmt.Printf("%s", result.Response.Status)
	contentType := result.Response.Header.Get("Content-Type")
	switch true {
	case strings.Contains(contentType, "application/json"):
		var obj map[string]interface{}
		err := json.Unmarshal(result.Body, &obj)
		if err != nil {
			panic(err.Error())
		}
		pretty, err := json.MarshalIndent(obj, "", "  ")
		fmt.Println(string(pretty))
	case strings.Contains(contentType, "text/html"):
		fmt.Println(string(result.Body))
	case strings.Contains(contentType, "text/plain"):
		fmt.Println(string(result.Body))
	default:
		panic(fmt.Sprintf("Unable to decode content-type: %s", contentType))
	}

	fmt.Printf("Duration: %v\n", result.Elapsed)
}

func readRequestDefinition(fileName string) (*RequestDefinition, error) {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	var request *RequestDefinition
	err = yaml.Unmarshal(file, &request)
	if err != nil {
		return nil, err
	}
	return request, nil

}

func validateAuthorizationHook(responseStatus int, definition *RequestDefinition) bool {
	if definition.AuthorizationHook != nil &&
		definition.AuthorizationHook.RequestPath != "" &&
		contains(definition.AuthorizationHook.OnHttpStatus, responseStatus) {
		return true
	}
	return false
}

func runAuthorizationHook(definition *RequestDefinition) (*RequestResult, error) {
	authHook := definition.AuthorizationHook
	authDefinition, err := readRequestDefinition(authHook.RequestPath)
	if err != nil {
		return nil, err
	}
	authResult, err := DoRequest(authDefinition)
	if err != nil {
		return nil, err
	}
	switch authHook.AuthType {
	case AUTH_TYPE_BEARER_TOKEN:
		if authHook.JsonParseBodyPath != "" && strings.Contains(authResult.Response.Header.Get("Content-Type"), "application/json") {
			err = decorateWithBearerTokenFromJson(definition, authResult.Body, authHook.JsonParseBodyPath)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("unknown authorization hook request parsing strategy")
		}
	default:
		return nil, fmt.Errorf("unknown authtype: %s", authHook.AuthType)
	}
	result, err := DoRequest(definition)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func decorateWithBearerTokenFromJson(definition *RequestDefinition, body []byte, jsonPath string) error {
	if definition.Headers == nil {
		definition.Headers = make(map[string]string)
	}
	jsonBody, err := gabs.ParseJSON(body)
	if err != nil {
		return fmt.Errorf("unable to unmarshal auth hook body: %v", err.Error())
	}
	token, ok := jsonBody.Path(definition.AuthorizationHook.JsonParseBodyPath).Data().(string)
	if !ok {
		return fmt.Errorf("unable to retrieve token by path %s", definition.AuthorizationHook.JsonParseBodyPath)
	}
	definition.Headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
	return nil
}

func DoMethod(definition *RequestDefinition) (*RequestResult, error) {
	result, err := DoRequest(definition)
	if err != nil {
		return nil, err
	}
	if result.Response != nil {
		if validateAuthorizationHook(result.Response.StatusCode, definition) {
			return runAuthorizationHook(definition)
		}
	}

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

func DoRequest(definition *RequestDefinition) (*RequestResult, error) {
	var req *http.Request

	client := &http.Client{}
	reqUrl, err := url.Parse(definition.URL)
	if err != nil {
		return nil, err
	}

	if definition.Body != nil {
		if definition.Headers["Content-Type"] == "application/json" {
			body := convert(definition.Body)
			rawBody, err := json.Marshal(body)
			if err != nil {
				panic(fmt.Errorf("unable to marshal definition.Body into json: %v", err))
			}
			req, err = http.NewRequest(
				definition.Method,
				reqUrl.String(),
				bytes.NewReader(rawBody),
			)
			if err != nil {
				panic(fmt.Sprintf("unable to create new reququest: %v", err))
			}
		} else {
			panic(fmt.Sprintf("No request body parser available for %s", definition.Headers["Content-Type"]))
		}
	} else {
		if definition.BodyStr == "" {
			req, err = http.NewRequest(
				definition.Method,
				reqUrl.String(),
				nil,
			)
		} else {
			req, err = http.NewRequest(
				definition.Method,
				reqUrl.String(),
				strings.NewReader(definition.BodyStr),
			)
		}
	}
	if err != nil {
		return nil, err
	}
	for k, v := range definition.Headers {
		req.Header.Add(k, v)
	}

	var result *RequestResult

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)
	result = &RequestResult{
		Elapsed:  elapsed,
		Response: resp,
		Body:     nil,
	}
	if err != nil {
		return result, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}
	resp.Body.Close()
	resp.Body = nil
	result.Body = body

	return result, nil
}
