package main

import (
	internal "github.com/mnalsup/method/internal"
	"github.com/mnalsup/method/internal/authentication"
	"github.com/mnalsup/method/internal/output"
	"github.com/mnalsup/method/internal/yaml"
)

type AuthRequest struct{}

func (a *AuthRequest) Do(path string) (*authentication.AuthResult, error) {
	schema, err := yaml.ReadRequestSchema(path)
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
	schema *internal.RequestSchema
}

func (a *AuthDefinitionRefresher) RefreshDefinition() error {
	schema, err := yaml.ReadRequestSchema(a.schema.FileName)
	if err != nil {
		return err
	}
	*a.schema = *schema
	return nil
}
