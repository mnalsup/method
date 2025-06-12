package internal

import (
	// Normally imports in top-level types are not good, but accepting carve-outs since
	// the domain _is_ http to some extent
	"net/http"
	"time"
)

// RequestSchema represents the complete configuration for a method request
type RequestSchema struct {
	FileName         string
	RequestDefinition `yaml:",inline"`
	ValueResolvers   []ValueResolver `yaml:"valueResolvers"`
}

// RequestDefinition defines the HTTP request to be made
type RequestDefinition struct {
	Method  string            `yaml:"method"`
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
	BodyStr string            `yaml:"bodyStr"`
	Body    interface{}       `yaml:"body"`
	Files   []FileDefinition  `yaml:"files"`
}

// FileDefinition specifies file uploads in multipart requests
type FileDefinition struct {
	RequestBodyPath string `yaml:"requestBodyPath"`
	FilePath        string `yaml:"filePath"`
}

// ValueResolver runs another request to resolve values that are injected into the main request
type ValueResolver struct {
	Name              string                   `yaml:"name"`
	RequestPath       string                   `yaml:"requestPath"`
	JsonParseBodyPath string                   `yaml:"jsonParseBodyPath"`
	Target            ValueResolverTarget      `yaml:"target"`
}

// ValueResolverTarget specifies where the resolved value should be placed
type ValueResolverTarget struct {
	Header              *HeaderTarget              `yaml:"header"`
	EnvironmentVariable *EnvironmentVariableTarget `yaml:"environmentVariable"`
}

// HeaderTarget specifies placing the resolved value in an HTTP header
type HeaderTarget struct {
	Name         string `yaml:"name"`
	FormatString string `yaml:"formatString"`
}

// EnvironmentVariableTarget specifies placing the resolved value in an environment variable
type EnvironmentVariableTarget struct {
	Name string `yaml:"name"`
}

// RequestResult holds the result of an HTTP request
type RequestResult struct {
	Body     []byte
	Response *http.Response
	Elapsed  time.Duration
}

// RequestRunner defines the interface for executing HTTP requests
type RequestRunner interface {
	RunRequest(path string) (RequestResult, error)
}

// SchemaLoader defines the interface for loading request schemas
type SchemaLoader interface {
	LoadSchema(fileName string, cached bool) (*RequestSchema, error)
}

// OutputFormatter defines the interface for formatting request results
type OutputFormatter interface {
	FormatResponse(result RequestResult) string
}
