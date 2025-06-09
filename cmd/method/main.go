package main

import (
	"fmt"
	"os"

	internal "github.com/mnalsup/method/internal"
	"github.com/mnalsup/method/internal/authentication"
	"github.com/mnalsup/method/internal/output"
	"github.com/mnalsup/method/internal/request"
	"github.com/mnalsup/method/internal/yaml"
	"github.com/mnalsup/method/logging"
)

func ReadRequestFileName() (string, error) {
	fmt.Println(os.Args)
	fmt.Println(len(os.Args))
	if len(os.Args) < 2 {
		return "", fmt.Errorf("you must supply at least one argument to method")
	}
	args := os.Args[1:]
	fileName := args[0]
	return fileName, nil
}

func main() {
	log := logging.GetLogger()
	defer log.Sync()
	log.Debugf("Initiated logger, starting request")
	fileName, err := ReadRequestFileName()
	if err != nil {
		panic(err.Error())
	}
	log.Debugf("reading: %s", fileName)

	schema, err := yaml.ReadCachedRequestSchema(fileName)
	if err != nil {
		panic(err.Error())
	}

	result, err := DoMethod(schema)
	if err != nil {
		panic(err.Error())
	}

	err = yaml.WriteRequestSchema(fileName, schema)
	if err != nil {
		panic(err.Error())
	}

	output.PrintRequestResult(result.Body, result.Response, result.Elapsed)
}

func DoMethod(schema *internal.RequestSchema) (*request.RequestResult, error) {
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

func contains(elems []int, v int) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}
