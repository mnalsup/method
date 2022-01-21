package request

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

	"log"

	"github.com/Jeffail/gabs/v2"
	"github.com/drone/envsubst"
	"github.com/mnalsup/method/args"
	"gopkg.in/yaml.v2"
)

const (
	AUTH_TYPE_BEARER_TOKEN = "BearerToken"
)

var fileName string

type AuthenticationTrigger struct {
	OnHttpStatus []int        `yaml:"onHttpStatus"`
	OnJsonValue  *OnJsonValue `yaml:"onJsonValue"`
}

type BearerToken struct{}

type AuthHeader struct {
	Header       string `yaml:"header"`
	FormatString string `yaml:"formatString"`
}

type AuthEnvironmentVariable struct {
	Variable string `yaml:"variable"`
}

type AuthenticationHook struct {
	Triggers            []AuthenticationTrigger  `yaml:"triggers"`
	RequestPath         string                   `yaml:"requestPath"`
	JsonParseBodyPath   string                   `yaml:"jsonParseBodyPath"`
	BearerToken         *BearerToken             `yaml:"bearerToken"`
	AuthHeader          *AuthHeader              `yaml:"authHeader"`
	EnvironmentVariable *AuthEnvironmentVariable `yaml:"environmentVariable"`
}

type OnJsonValue struct {
	Path  string `yaml:"path"`
	Value string `yaml:"value"`
}

type RequestDefinition struct {
	Method             string              `yaml:"method"`
	URL                string              `yaml:"url"`
	Headers            map[string]string   `yaml:"headers"`
	AuthenticationHook *AuthenticationHook `yaml:"authenticationHook"`
	BodyStr            string              `yaml:"bodyStr"`
	Body               interface{}         `yaml:"body"`
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

func PrintRequestResult(result *RequestResult) {
	fmt.Println("--------------------Results--------------------")
	fmt.Printf("%s\n", result.Response.Status)

	for k, v := range result.Response.Header {
		fmt.Printf("%s: %s\n", k, v)
	}

	fmt.Println("")
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
		fmt.Println(fmt.Sprintf("Unable to decode content-type: %s printing raw output", contentType))
		print(string(result.Body))
	}

	fmt.Printf("Duration: %v\n", result.Elapsed)
	fmt.Println("-----------------------------------------------")
}

func ReadRequestDefinition(fileName string, request *RequestDefinition) error {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	subst, err := envsubst.EvalEnv(string(file))
	if err != nil {
		return err
	}
	err = yaml.Unmarshal([]byte(subst), &request)
	if err != nil {
		return err
	}
	return nil

}

func validateAuthenticationHook(result *RequestResult, definition *RequestDefinition) (bool, error) {
	if definition.AuthenticationHook == nil {
		return false, nil
	}
	if definition.AuthenticationHook.RequestPath == "" {
		return false, fmt.Errorf("unable to run authentication hook for empty request")
	}
	for _, v := range definition.AuthenticationHook.Triggers {
		// Check if configured

		// check if HTTP status mateches the hook
		if v.OnHttpStatus != nil &&
			contains(v.OnHttpStatus, result.Response.StatusCode) {
			return true, nil
		}

		// check if the JSON parsed value matches the hook
		if v.OnJsonValue != nil && result.Body != nil {
			body, err := gabs.ParseJSON(result.Body)
			if err != nil {
				panic(fmt.Errorf("Unable to parse json body for OnJsonValue: %s", err.Error()))
			}
			matchValue, ok := body.Path(v.OnJsonValue.Path).Data().(string)
			if ok {
				match := v.OnJsonValue.Value == matchValue
				if match {
					return true, nil
				}
			}
		}

	}
	return false, nil
}

/**
 * runAuthenticationHook
 *
 */
func runAuthenticationHook(definition *RequestDefinition) error {
	// Must initialize or ReadRequestDefinition will fail
	var authDefinition *RequestDefinition = &RequestDefinition{}
	fmt.Println("Running authentication hook...")
	authHook := definition.AuthenticationHook
	err := ReadRequestDefinition(authHook.RequestPath, authDefinition)
	if err != nil {
		return err
	}
	authResult, err := DoRequest(authDefinition)
	if err != nil {
		return err
	}
	PrintRequestResult(authResult)
	if authResult.Response.StatusCode > 399 {
		return fmt.Errorf("failed to retrieve credentials: %v", authResult.Response.Status)
	}
	token, err := getAuthToken(definition, authResult.Response.Header, authResult.Body)
	if err != nil {
		return fmt.Errorf("Unable to parse auth token from response: %s", err.Error())
	}
	switch {
	case authHook.BearerToken != nil:
		log.Println("Using BearerToken")
		err = decorateWithBearerToken(definition, token)
	case authHook.AuthHeader != nil:
		log.Println("Using AuthHeader")
		err = decorateWithAuthHeader(definition, token, authHook.AuthHeader.Header, authHook.AuthHeader.FormatString)
	case authHook.EnvironmentVariable != nil:
		log.Println("Using AuthEnvironmentVariable")
		os.Setenv(authHook.EnvironmentVariable.Variable, token)
		err := ReadRequestDefinition(args.ReadRequestFileName(), definition)
		if err != nil {
			return fmt.Errorf("unable to read request definition during auth environment variable hook")
		}
	default:
		return fmt.Errorf("unknown auth strategy should use one of [bearerToken, authHeader]")
	}
	return nil
}

func getAuthToken(definition *RequestDefinition, header http.Header, body []byte) (string, error) {
	authHook := definition.AuthenticationHook
	switch {
	case authHook.JsonParseBodyPath != "":
		if strings.Contains(header.Get("Content-Type"), "application/json") {
			return getJsonAuthToken(definition, body)
		} else {
			return "", fmt.Errorf("unable to parse json path from non-json auth response")
		}
	default:
		return "", fmt.Errorf("unable to get auth token, invalid parse strategy use: [JsonParseBodyPath]")
	}
}

func getJsonAuthToken(definition *RequestDefinition, body []byte) (string, error) {
	jsonBody, err := gabs.ParseJSON(body)
	if err != nil {
		return "", fmt.Errorf("unable to unmarshal auth hook body: %v", err.Error())
	}
	token, ok := jsonBody.Path(definition.AuthenticationHook.JsonParseBodyPath).Data().(string)
	if !ok {
		return "", fmt.Errorf("unable to retrieve token by path %s", definition.AuthenticationHook.JsonParseBodyPath)
	}
	return token, nil
}

func decorateWithBearerToken(definition *RequestDefinition, token string) error {
	return decorateWithAuthHeader(definition, token, "Authorization", "Bearer %s")
}

func decorateWithAuthHeader(definition *RequestDefinition, token string, header string, formatString string) error {
	if definition.Headers == nil {
		definition.Headers = make(map[string]string)
	}
	definition.Headers[header] = fmt.Sprintf(formatString, token)
	fmt.Printf("added header %s: %s", header, fmt.Sprintf(formatString, token))
	return nil
}

func DoMethod(definition *RequestDefinition) (*RequestResult, error) {
	result, err := DoRequest(definition)
	if err != nil {
		return nil, err
	}
	shouldTryAuthorize, err := validateAuthenticationHook(result, definition)
	if err != nil {
		return nil, fmt.Errorf("unable to validate authentication hook: %s", err.Error())
	}
	if shouldTryAuthorize {
		err := runAuthenticationHook(definition)
		if err != nil {
			return nil, fmt.Errorf("failed to authenticate: %s", err.Error())
		}
		result, err := DoRequest(definition)
		if err != nil {
			PrintRequestResult(result)
			return nil, err
		}
		return result, nil
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
		switch definition.Headers["Content-Type"] {
		case "application/json":
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
			req, err = http.NewRequest(
				definition.Method,
				reqUrl.String(),
				strings.NewReader(data.Encode()),
			)
		default:
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
