package http

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mnalsup/method/internal"
	"github.com/mnalsup/method/logging"
	"go.uber.org/zap"
)

// Formatter implements the internal.OutputFormatter interface
type Formatter struct{}

// NewFormatter creates a new output formatter
func NewFormatter() *Formatter {
	return &Formatter{}
}

// FormatResponse formats a request result into a readable string
func (f *Formatter) FormatResponse(result internal.RequestResult) string {
	log := logging.GetLogger()
	output := ""

	output += fmt.Sprintln("--------------------Results--------------------")
	output += fmt.Sprintf("%s\n", result.Response.Status)

	// Print response headers
	for k, v := range result.Response.Header {
		output += fmt.Sprintf("%s: %s\n", k, v)
	}

	output += fmt.Sprintln("")

	// Format response body based on content type
	contentType := result.Response.Header.Get("Content-Type")
	switch {
	case strings.Contains(contentType, "application/json"):
		output += f.formatJSON(result.Body, log)
	case strings.Contains(contentType, "text/html"):
		output += fmt.Sprintln(string(result.Body))
	case strings.Contains(contentType, "text/plain"):
		output += fmt.Sprintln(string(result.Body))
	case strings.Contains(contentType, "application/octet-stream"):
		output += fmt.Sprintln(string(result.Body))
	default:
		output += fmt.Sprintf("Unable to decode content-type: %s printing raw output\n", contentType)
		output += fmt.Sprintln(string(result.Body))
	}

	output += fmt.Sprintf("Duration: %v\n", result.Elapsed)
	output += fmt.Sprintln("-----------------------------------------------")

	return output
}

// formatJSON attempts to pretty-print JSON response bodies  
func (f *Formatter) formatJSON(body []byte, log *zap.SugaredLogger) string {
	var obj interface{}
	err := json.Unmarshal(body, &obj)
	if err != nil {
		log.Errorf("invalid JSON response, using raw output: %v", err)
		return fmt.Sprintln(string(body))
	}

	pretty, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		log.Errorf("unable to pretty print JSON: %v", err)
		return fmt.Sprintln(string(body))
	}

	return fmt.Sprintln(string(pretty))
} 