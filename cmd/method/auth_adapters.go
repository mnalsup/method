package main

import (
	"github.com/mnalsup/method/internal/authentication"
	"github.com/mnalsup/method/internal/output"
)

type AuthRequest struct{}

func (a *AuthRequest) Do(path string) (*authentication.AuthResult, error) {
	schema, err := ReadRequestSchema(path)
	if err != nil {
		return nil, err
	}
	result, err := DoMethod(schema)
	if err != nil {
		return nil, err
	}
	return &authentication.AuthResult{
		Response: result.Response,
		Elapsed:  result.Elapsed,
		Body:     result.Body,
	}, nil
}

type AuthPrinter struct{}

func (p *AuthPrinter) PrintRequestResult(res *authentication.AuthResult) {
	output.PrintRequestResult(res.Body, res.Response, res.Elapsed)
}

type AuthDefinitionRefresher struct {
	schema *RequestSchema
}

func (a *AuthDefinitionRefresher) RefreshDefinition() error {
	schema, err := ReadRequestSchema(a.schema.fileName)
	if err != nil {
		return err
	}
	*a.schema = *schema
	return nil
}
