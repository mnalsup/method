package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/mnalsup/method/args"
	"github.com/mnalsup/method/internal/output"
	"github.com/mnalsup/method/logging"
)

func main() {
	log := logging.GetLogger()
	defer log.Sync()
	log.Debugf("Initiated logger, starting request")
	fileName := args.ReadRequestFileName()

	schema, err := ReadRequestSchema(fileName)
	if err != nil {
		panic(err.Error())
	}
	cachedSchema, err := ReadRequestSchema(GetTempFileName(fileName))
	if err == nil {
		MergeTempFileSchema(schema, cachedSchema)
	}

	result, err := DoMethod(schema)
	if err != nil {
		panic(err.Error())
	}

	out, err := yaml.Marshal(&schema)
	if err != nil {
		panic(err.Error())
	}

	err = os.WriteFile(GetTempFileName(fileName), out, 0666)
	if err != nil {
		panic(err.Error())
	}

	output.PrintRequestResult(result.Body, result.Response, result.Elapsed)
}

func MergeTempFileSchema(orig *RequestSchema, temp *RequestSchema) {
	// merge any cached headers into the original request
	if temp.Headers != nil {
		if orig.Headers == nil {
			orig.Headers = make(map[string]string)
		}
		for k, v := range temp.Headers {
			if orig.Headers[k] == "" {
				orig.Headers[k] = v
			}
		}
	}
}

func GetTempFileName(fileName string) string {
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
