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

	body, resp, err := DoMethod(definition)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("%s", resp.Status)
	contentType := resp.Header.Get("Content-Type")

	if strings.Contains(contentType, "application/json") {
		var obj map[string]interface{}
		err := json.Unmarshal(body, &obj)
		if err != nil {
			panic(err.Error())
		}
		pretty, err := json.MarshalIndent(obj, "", "  ")
		fmt.Println(string(pretty))
		return
	}
	if strings.Contains(contentType, "text/html") {
		fmt.Println(string(body))
		return
	}
	if strings.Contains(contentType, "text/plain") {
		fmt.Println(string(body))
		return
	}
	panic(fmt.Sprintf("Unable to decode content-type: %s", contentType))
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
	fmt.Println(responseStatus)
	if definition.AuthorizationHook != nil &&
		definition.AuthorizationHook.RequestPath != "" &&
		contains(definition.AuthorizationHook.OnHttpStatus, responseStatus) {
		return true
	}
	return false
}

func runAuthorizationHook(definition *RequestDefinition) ([]byte, *http.Response, error) {
	authHook := definition.AuthorizationHook
	authDefinition, err := readRequestDefinition(authHook.RequestPath)
	if err != nil {
		return nil, nil, err
	}
	authBody, authResp, err := DoRequest(authDefinition)
	if err != nil {
		return nil, nil, err
	}
	switch authHook.AuthType {
	case AUTH_TYPE_BEARER_TOKEN:
		if authHook.JsonParseBodyPath != "" && strings.Contains(authResp.Header.Get("Content-Type"), "application/json") {
			err = decorateWithBearerTokenFromJson(definition, authBody, authHook.JsonParseBodyPath)
			if err != nil {
				return nil, nil, err
			}
		} else {
			return nil, nil, fmt.Errorf("unknown authorization hook request parsing strategy")
		}
	default:
		return nil, nil, fmt.Errorf("unknown authtype: %s", authHook.AuthType)
	}
	for k, v := range definition.Headers {
		fmt.Printf("%s: %s\n", k, v)
	}
	body, resp, err := DoRequest(definition)
	if err != nil {
		return nil, nil, err
	}
	return body, resp, nil
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

func DoMethod(definition *RequestDefinition) ([]byte, *http.Response, error) {
	body, resp, err := DoRequest(definition)
	if err != nil {
		return nil, nil, err
	}
	if resp != nil {
		if validateAuthorizationHook(resp.StatusCode, definition) {
			fmt.Println("Validated Authorization Hook")
			return runAuthorizationHook(definition)
		}
	}

	return body, resp, nil

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

func DoRequest(definition *RequestDefinition) ([]byte, *http.Response, error) {
	var req *http.Request

	client := &http.Client{}
	reqUrl, err := url.Parse(definition.URL)
	if err != nil {
		return nil, nil, err
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
		return nil, nil, err
	}
	for k, v := range definition.Headers {
		fmt.Printf("adding header %s: %s\n", k, v)
		req.Header.Add(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, resp, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	resp.Body.Close()
	resp.Body = nil

	return body, resp, nil
}
