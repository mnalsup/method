package http

import (
	"fmt"
	"os"

	"github.com/Jeffail/gabs"
	"github.com/mnalsup/method/internal"
	"github.com/mnalsup/method/logging"
)

// Handler handles HTTP requests using dependency injection
type Handler struct {
	schemaLoader    internal.SchemaLoader
	outputFormatter internal.OutputFormatter
	requestRunner   internal.RequestRunner
}

// NewHandler creates a new HTTP handler with injected dependencies
func NewHandler(schemaLoader internal.SchemaLoader, outputFormatter internal.OutputFormatter, requestRunner internal.RequestRunner) *Handler {
	return &Handler{
		schemaLoader:    schemaLoader,
		outputFormatter: outputFormatter,
		requestRunner:   requestRunner,
	}
}

// ExecuteRequest executes a request schema and returns formatted output
func (h *Handler) ExecuteRequest(fileName string, cached bool) (string, error) {
	log := logging.GetLogger()
	
	schema, err := h.schemaLoader.LoadSchema(fileName, cached)
	if err != nil {
		return "", fmt.Errorf("failed to load schema: %w", err)
	}

	// Initialize headers if nil
	if schema.RequestDefinition.Headers == nil {
		schema.RequestDefinition.Headers = make(map[string]string)
	}

	// Resolve any values first
	err = h.resolveValues(schema)
	if err != nil {
		return "", fmt.Errorf("failed to resolve values: %w", err)
	}

	// Execute the main request
	result, err := doRequest(&schema.RequestDefinition)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}

	log.Debugf("Request completed with status: %s", result.Response.Status)
	
	// Format and return the output
	return h.outputFormatter.FormatResponse(*result), nil
}

// resolveValues processes all value resolvers and applies their results
func (h *Handler) resolveValues(schema *internal.RequestSchema) error {
	log := logging.GetLogger()
	
	if len(schema.ValueResolvers) == 0 {
		log.Debug("No value resolvers configured")
		return nil
	}
	
	log.Infof("✨ Found %d value resolver(s) - using new ValueResolver system", len(schema.ValueResolvers))
	
	for _, resolver := range schema.ValueResolvers {
		log.Debugf("Processing value resolver: %s", resolver.Name)
		
		// Execute the resolver request
		result, err := h.requestRunner.RunRequest(resolver.RequestPath)
		if err != nil {
			return fmt.Errorf("failed to execute resolver request for %s: %w", resolver.Name, err)
		}

		if result.Response.StatusCode >= 400 {
			return fmt.Errorf("resolver request failed for %s with status: %s", resolver.Name, result.Response.Status)
		}

		// Parse the value from the response
		value, err := h.parseValueFromResponse(resolver, result)
		if err != nil {
			return fmt.Errorf("failed to parse value from resolver %s: %w", resolver.Name, err)
		}

		// Apply the value to its target
		err = h.applyResolvedValue(schema, resolver, value)
		if err != nil {
			return fmt.Errorf("failed to apply resolved value for %s: %w", resolver.Name, err)
		}

		log.Debugf("Successfully resolved value for %s", resolver.Name)
	}

	return nil
}

// parseValueFromResponse extracts the desired value from a resolver response
func (h *Handler) parseValueFromResponse(resolver internal.ValueResolver, result internal.RequestResult) (string, error) {
	if resolver.JsonParseBodyPath == "" {
		return "", fmt.Errorf("no parse path specified for resolver %s", resolver.Name)
	}

	// Parse JSON response
	jsonBody, err := gabs.ParseJSON(result.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Extract value using the specified path
	value, ok := jsonBody.Path(resolver.JsonParseBodyPath).Data().(string)
	if !ok {
		return "", fmt.Errorf("failed to extract string value at path %s", resolver.JsonParseBodyPath)
	}

	return value, nil
}

// applyResolvedValue applies a resolved value to its configured target
func (h *Handler) applyResolvedValue(schema *internal.RequestSchema, resolver internal.ValueResolver, value string) error {
	log := logging.GetLogger()
	
	target := resolver.Target

	switch {
	case target.Header != nil:
		headerValue := value
		if target.Header.FormatString != "" {
			headerValue = fmt.Sprintf(target.Header.FormatString, value)
		}
		schema.RequestDefinition.Headers[target.Header.Name] = headerValue
		log.Debugf("Set header %s: %s", target.Header.Name, headerValue)
		
	case target.EnvironmentVariable != nil:
		err := os.Setenv(target.EnvironmentVariable.Name, value)
		if err != nil {
			return fmt.Errorf("failed to set environment variable %s: %w", target.EnvironmentVariable.Name, err)
		}
		log.Debugf("Set environment variable %s", target.EnvironmentVariable.Name)
		
	default:
		return fmt.Errorf("no valid target specified for resolver %s", resolver.Name)
	}

	return nil
} 