package yaml

import (
	"fmt"
	"os"
	"strings"

	"github.com/drone/envsubst"
	"github.com/mnalsup/method/internal"
	"gopkg.in/yaml.v2"
)

// SchemaLoader implements the internal.SchemaLoader interface
type SchemaLoader struct{}

// NewSchemaLoader creates a new schema loader
func NewSchemaLoader() *SchemaLoader {
	return &SchemaLoader{}
}

// LoadSchema loads a request schema from a YAML file
func (l *SchemaLoader) LoadSchema(fileName string, cached bool) (*internal.RequestSchema, error) {
	if cached {
		return l.readCachedRequestSchema(fileName)
	}
	return l.readRequestSchema(fileName)
}

// readRequestSchema reads and parses a YAML request schema file
func (l *SchemaLoader) readRequestSchema(fileName string) (*internal.RequestSchema, error) {
	var schema internal.RequestSchema
	
	file, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", fileName, err)
	}

	// Substitute environment variables
	subst, err := envsubst.EvalEnv(string(file))
	if err != nil {
		return nil, fmt.Errorf("failed to substitute environment variables: %w", err)
	}

	// Parse YAML
	err = yaml.Unmarshal([]byte(subst), &schema)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	schema.FileName = fileName
	return &schema, nil
}

// readCachedRequestSchema reads a schema and merges it with cached data
func (l *SchemaLoader) readCachedRequestSchema(fileName string) (*internal.RequestSchema, error) {
	schema, err := l.readRequestSchema(fileName)
	if err != nil {
		return nil, err
	}

	// Try to read cached schema
	cachedSchema, err := l.readRequestSchema(getTempFileName(fileName))
	if err == nil {
		mergeTempFileSchema(schema, cachedSchema)
	}
	// Ignore cache read errors - cache is optional

	return schema, nil
}

// WriteTempFile writes a schema to a temporary cache file
func WriteTempFile(fileName string, schema *internal.RequestSchema) error {
	out, err := yaml.Marshal(&schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	tempFileName := getTempFileName(fileName)
	err = os.WriteFile(tempFileName, out, 0666)
	if err != nil {
		return fmt.Errorf("failed to write temp file %s: %w", tempFileName, err)
	}

	return nil
}

// mergeTempFileSchema merges cached headers into the original request
func mergeTempFileSchema(orig *internal.RequestSchema, temp *internal.RequestSchema) {
	// merge any cached headers into the original request
	if temp.RequestDefinition.Headers != nil {
		if orig.RequestDefinition.Headers == nil {
			orig.RequestDefinition.Headers = make(map[string]string)
		}
		for k, v := range temp.RequestDefinition.Headers {
			if orig.RequestDefinition.Headers[k] == "" {
				orig.RequestDefinition.Headers[k] = v
			}
		}
	}
}

// getTempFileName generates a temporary file name for caching
func getTempFileName(fileName string) string {
	pathParts := strings.Split(fileName, "/")
	dir := strings.Join(pathParts[:len(pathParts)-1], "/")
	file := pathParts[len(pathParts)-1]

	fileParts := strings.Split(file, ".")
	name := strings.Join(fileParts[:len(fileParts)-1], ".")
	ext := fileParts[len(fileParts)-1]
	tmpFileName := fmt.Sprintf(".%s.%s.%s", name, "tmp", ext)

	var tmpFilePath string
	if dir != "" {
		tmpFilePath = strings.Join([]string{dir, tmpFileName}, "/")
	} else {
		tmpFilePath = tmpFileName
	}

	return tmpFilePath
}

// Legacy functions for backwards compatibility
func ReadRequestSchema(fileName string) (*internal.RequestSchema, error) {
	loader := NewSchemaLoader()
	return loader.LoadSchema(fileName, false)
}

func ReadCachedRequestSchema(fileName string) (*internal.RequestSchema, error) {
	loader := NewSchemaLoader()
	return loader.LoadSchema(fileName, true)
}
