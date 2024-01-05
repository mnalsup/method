package authentication

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/mnalsup/method/logging"
)

type AuthHeader struct {
	Header       string `yaml:"header"`
	FormatString string `yaml:"formatString"`
}

type AuthEnvironmentVariable struct {
	Variable string `yaml:"variable"`
}

type BearerToken struct{}

type AuthenticationTrigger struct {
	OnHttpStatus []int        `yaml:"onHttpStatus"`
	OnJsonValue  *OnJsonValue `yaml:"onJsonValue"`
}

type OnJsonValue struct {
	Path  string `yaml:"path"`
	Value string `yaml:"value"`
}

type AuthenticationHook struct {
	Triggers            []AuthenticationTrigger  `yaml:"triggers"`
	RequestPath         string                   `yaml:"requestPath"`
	JsonParseBodyPath   string                   `yaml:"jsonParseBodyPath"`
	BearerToken         *BearerToken             `yaml:"bearerToken"`
	AuthHeader          *AuthHeader              `yaml:"authHeader"`
	EnvironmentVariable *AuthEnvironmentVariable `yaml:"environmentVariable"`
}

type AuthResult struct {
	Response *http.Response
	Elapsed  time.Duration
	Body     []byte
}

type AuthHeaders map[string]string

type RequestDoer interface {
	Do(requestPath string) (*AuthResult, error)
}

type ResultPrinter interface {
	PrintRequestResult(*AuthResult)
}

type DefinitionRefresher interface {
	// This is a janky bit of signalling that tells method to refresh the definition
	RefreshDefinition() error
}

/**
 * runAuthenticationHook
 *
 */
func RunAuthenticationHook(hook *AuthenticationHook, requestDoer RequestDoer, refresher DefinitionRefresher, resultPrinter ResultPrinter) (AuthHeaders, error) {
	log := logging.GetLogger()
	// Must initialize or ReadRequestDefinition will fail
	log.Debug("Running authentication hook...")
	authResult, err := requestDoer.Do(hook.RequestPath)
	if err != nil {
		return nil, err
	}
	resultPrinter.PrintRequestResult(authResult)
	if authResult.Response.StatusCode > 399 {
		return nil, fmt.Errorf("failed to retrieve credentials: %v", authResult.Response.Status)
	}
	token, err := getAuthToken(hook, authResult.Response.Header, authResult.Body)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse auth token from response: %s", err.Error())
	}
	switch {
	case hook.BearerToken != nil:
		log.Info("Using BearerToken")
		return decorateWithAuthHeader(token, "Authorization", "Bearer %s"), nil
	case hook.AuthHeader != nil:
		log.Info("Using AuthHeader")
		return decorateWithAuthHeader(token, hook.AuthHeader.Header, hook.AuthHeader.FormatString), nil
	case hook.EnvironmentVariable != nil:
		authHeaders := make(AuthHeaders, 0)
		log.Info("Using AuthEnvironmentVariable")
		os.Setenv(hook.EnvironmentVariable.Variable, token)
		err := refresher.RefreshDefinition()
		if err != nil {
			return nil, err
		}
		return authHeaders, nil
	default:
		return nil, fmt.Errorf("unknown auth strategy should use one of [bearerToken, authHeader]")
	}
}

func getAuthToken(authHook *AuthenticationHook, header http.Header, body []byte) (string, error) {
	switch {
	case authHook.JsonParseBodyPath != "":
		if strings.Contains(header.Get("Content-Type"), "application/json") {
			return getJsonAuthToken(authHook, body)
		} else {
			return "", fmt.Errorf("unable to parse json path from non-json auth response")
		}
	default:
		return "", fmt.Errorf("unable to get auth token, invalid parse strategy use: [JsonParseBodyPath]")
	}
}

func getJsonAuthToken(authHook *AuthenticationHook, body []byte) (string, error) {
	jsonBody, err := gabs.ParseJSON(body)
	if err != nil {
		return "", fmt.Errorf("unable to unmarshal auth hook body: %v", err.Error())
	}
	token, ok := jsonBody.Path(authHook.JsonParseBodyPath).Data().(string)
	if !ok {
		return "", fmt.Errorf("unable to retrieve token by path %s", authHook.JsonParseBodyPath)
	}
	return token, nil
}

func decorateWithAuthHeader(token string, header string, formatString string) AuthHeaders {
	headers := make(AuthHeaders)
	headers[header] = fmt.Sprintf(formatString, token)
	fmt.Printf("added header %s: %s", header, fmt.Sprintf(formatString, token))
	return headers
}
