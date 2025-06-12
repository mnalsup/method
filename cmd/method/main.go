package main

import (
	"fmt"
	"os"

	"github.com/mnalsup/method/internal/http"
	"github.com/mnalsup/method/internal/yaml"
	"github.com/mnalsup/method/logging"
)

func ReadRequestFileName() (string, error) {
	if len(os.Args) < 2 {
		return "", fmt.Errorf("you must supply at least one argument to method")
	}
	args := os.Args[1:]
	fileName := args[0]
	return fileName, nil
}

func main() {
	// Always show version info to stderr regardless of log level
	fmt.Fprintf(os.Stderr, "🚀 METHOD v2.0 - Using new ValueResolver architecture (refactored)\n")
	
	log := logging.GetLogger()
	defer log.Sync()
	log.Debugf("Initiated logger, starting request")
	
	fileName, err := ReadRequestFileName()
	if err != nil {
		log.Fatalf("Failed to get request filename: %v", err)
	}
	log.Debugf("reading: %s", fileName)

	// Create dependencies using dependency injection
	schemaLoader := yaml.NewSchemaLoader()
	outputFormatter := http.NewFormatter()
	requestRunner := http.NewRequestRunner(schemaLoader)
	
	// Create the handler with injected dependencies
	handler := http.NewHandler(schemaLoader, outputFormatter, requestRunner)

	// Execute the request
	result, err := handler.ExecuteRequest(fileName, false)
	if err != nil {
		log.Fatalf("Failed to execute request: %v", err)
	}

	fmt.Print(result)
}
