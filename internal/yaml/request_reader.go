package yaml

import (
	"fmt"
	"os"
	"strings"

	"github.com/drone/envsubst"
	internal "github.com/mnalsup/method/internal"
	"gopkg.in/yaml.v2"
)

func ReadRequestSchema(fileName string) (*internal.RequestSchema, error) {
	schema, err := readRequestSchema(fileName)
	if err != nil {
		panic(err.Error())
	}
	return schema, nil
}

func ReadCachedRequestSchema(fileName string) (*internal.RequestSchema, error) {
	schema, err := readRequestSchema(fileName)
	if err != nil {
		panic(err.Error())
	}
	cachedSchema, err := readRequestSchema(getTempFileName(fileName))
	if err == nil {
		mergeTempFileSchema(schema, cachedSchema)
	}
	return schema, nil
}

func WriteRequestSchema(fileName string, schema *internal.RequestSchema) error {
	out, err := yaml.Marshal(&schema)
	if err != nil {
		panic(err.Error())
	}

	return os.WriteFile(getTempFileName(fileName), out, 0666)
}

func readRequestSchema(fileName string) (*internal.RequestSchema, error) {
	var schema internal.RequestSchema
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
	schema.FileName = fileName
	return &schema, nil
}

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
