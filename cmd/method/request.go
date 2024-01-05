package main

import (
	"fmt"
	"os"

	"github.com/drone/envsubst"
	"github.com/mnalsup/method"
	"github.com/mnalsup/method/authentication"
	"github.com/mnalsup/method/output"
	"github.com/mnalsup/method/request"
	"gopkg.in/yaml.v2"
)

type RequestSchema struct {
	fileName                          string
	method.RequestDefinition          `yaml:",inline"`
	authentication.AuthenticationHook `yaml:"authenticationHook"`
}

func contains(elems []int, v int) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}

func ReadRequestSchema(fileName string) (*RequestSchema, error) {
	var schema RequestSchema
	file, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	subst, err := envsubst.EvalEnv(string(file))
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal([]byte(subst), &schema)
	if err != nil {
		return nil, err
	}
	schema.fileName = fileName
	return &schema, nil
}

func DoMethod(schema *RequestSchema) (*request.RequestResult, error) {
	if schema.RequestDefinition.Headers == nil {
		schema.RequestDefinition.Headers = make(map[string]string)
	}
	result, err := request.DoRequest(&schema.RequestDefinition)
	if err != nil {
		return nil, err
	}
	shouldTryAuthorize, err := validateAuthenticationHook(result, &schema.AuthenticationHook)
	if err != nil {
		return nil, fmt.Errorf("unable to validate authentication hook: %s", err.Error())
	}
	if shouldTryAuthorize {
		headers, err := authentication.RunAuthenticationHook(&schema.AuthenticationHook, &AuthRequest{}, &AuthDefinitionRefresher{schema: schema}, &AuthPrinter{})
		if err != nil {
			return nil, fmt.Errorf("failed to authenticate: %s", err.Error())
		}
		for header, value := range headers {
			schema.RequestDefinition.Headers[header] = value
		}
		result, err := request.DoRequest(&schema.RequestDefinition)
		if err != nil {
			output.PrintRequestResult(result.Body, result.Response, result.Elapsed)
			return nil, err
		}
		return result, nil
	}

	return result, nil
}
