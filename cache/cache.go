package cache

import (
	"fmt"
	"github.com/mnalsup/method/request"
	"strings"
)

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

func MergeTempFileDefinition(orig *request.RequestDefinition, temp *request.RequestDefinition) {
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
