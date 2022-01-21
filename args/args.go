package args

import (
	"os"
)

func ReadRequestFileName() string {
	args := os.Args[1:]
	fileName := args[0]
	return fileName
}
