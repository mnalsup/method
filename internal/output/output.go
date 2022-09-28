package output

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mnalsup/method/logging"
)

func PrintRequestResult(
	body []byte,
	response *http.Response,
	elapsed time.Duration,
) {
	log := logging.GetLogger()
	fmt.Println("--------------------Results--------------------")
	fmt.Printf("%s\n", response.Status)

	for k, v := range response.Header {
		fmt.Printf("%s: %s\n", k, v)
	}

	fmt.Println("")
	contentType := response.Header.Get("Content-Type")
	switch true {
	case strings.Contains(contentType, "application/json"):
		var obj interface{}
		err := json.Unmarshal(body, &obj)
		if err != nil {
			log.Errorf("invalid JSON response, attempting other parsing schemes")
			break
		}
		pretty, err := json.MarshalIndent(obj, "", "  ")
		if err != nil {
			fmt.Println("unable to pretty print application/json")
			fmt.Println(string(body))
		}
		fmt.Println(string(pretty))
	case strings.Contains(contentType, "text/html"):
		fmt.Println(string(body))
	case strings.Contains(contentType, "text/plain"):
		fmt.Println(string(body))
	default:
		fmt.Printf("Unable to decode content-type: %s printing raw output\n", contentType)
		fmt.Println(string(body))
	}

	fmt.Printf("Duration: %v\n", elapsed)
	fmt.Println("-----------------------------------------------")
}
