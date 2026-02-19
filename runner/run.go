package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
    "strings"
    "html/template"


	"github.com/google/uuid"
	"github.com/itchyny/gojq"
)

const (
	reset = "\x1b[0m"
	bold  = "\x1b[1m"

	cyan   = "\x1b[38;2;122;162;247m" // #7AA2F7
	purple = "\x1b[38;2;187;154;247m" // #BB9AF7
)

func run(requests []Request, rawFlag bool, quiet bool) error {
	token, err := getToken()
	if err != nil {
		return fmt.Errorf("authentication error: %s", err)
	}

	// prepare stack map with values for the template
	stack := make(map[string]string)

	for _, request := range requests {
		stack["uuid"] = uuid.NewString()

		// TODO: fix parse() panic
		request.url = parse(request.url, stack)

        // body validation
        var requestBody io.Reader
        if request.body != nil {
            bodyBytes, _ := json.Marshal(request.body)
            bodyString := parse(string(bodyBytes), stack) // to be able to use saved variables
            requestBody = strings.NewReader(bodyString)
        }

		req, err := http.NewRequest(request.method, request.url, requestBody)
		if err != nil {
			return fmt.Errorf("client: could not create request: %s", err)
		}

		authHeader := fmt.Sprintf("Bearer %s", token)
		req.Header = http.Header{
			"Content-Type":  {"application/json"},
			"Authorization": {authHeader},
		}

		if request.headers != nil {
			for _, header := range request.headers {
                value := parse(header.value, stack)
				req.Header.Add(header.key, value)
			}
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("client: error making http request: %s", err)
		}

		defer res.Body.Close()

		// Response body
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("client: could not read response body: %s", err)
		}

        if !quiet {
            // print status code
            fmt.Printf("%s%sHTTP%s %v\n", bold, cyan, reset, res.Status)

            // print header
            traceIDs := res.Header["X-B3-Traceid"]
            if len(traceIDs) > 0 {
                fmt.Printf("%s%sX-B3-Traceid:%s %v\n", bold, purple, reset, traceIDs[0])
            }
        }

		// print response data
		fmt.Println(formatString(body, rawFlag))
        fmt.Println()

		if request.statusCode > 0 {
			respCode := res.StatusCode
			if respCode != request.statusCode {
				return fmt.Errorf("\033[0;31merror\033[0m: expected %v but got %v", request.statusCode, res.StatusCode)
			}
		}

		// TODO: simplify this and remove from function
		if request.captures != nil {
			for _, capture := range request.captures {
				// key = variable name from the capture: customer_id from "customer_id: .id" in the yaml file
				// val = the string to query the response json: .id from "customer_id: .id" in the yaml file
				// do the jq query on the response body string
				query, err := gojq.Parse(capture.value)
				if err != nil {
					fmt.Println(err)
				}

				// convert body byte to map[string]any to be able to run the query
				bodyAny := make(map[string]any)

				err_ := json.Unmarshal(body, &bodyAny)
				if err_ != nil {
					fmt.Printf("unmarshal: %v\n", err)
				}

				iter := query.Run(bodyAny)
				for {
					v, ok := iter.Next()
					if !ok {
						break
					}
					if err, ok := v.(error); ok {
						if err, ok := err.(*gojq.HaltError); ok && err.Value() == nil {
							break
						}
						fmt.Println(err)
					}
					stack[capture.key] = fmt.Sprintf("%v", v)
				}
			}
		}

		if request.wait > 0 {
			chars := []string{"|", "/", "-", "\\"}
			duration := time.Duration(request.wait) * time.Millisecond // Total duration of the animation

			startTime := time.Now()

			for time.Since(startTime) < duration {
				for _, char := range chars {
					fmt.Printf("waiting... %s\r", char) // \r to overwrite the previous character
					time.Sleep(100 * time.Millisecond)  // Adjust delay for speed
				}
			}
			fmt.Println("\rwaiting...  ")
		}
	}
	return nil
}

func parse(s string, stack map[string]string) string {
	urlTemplate, err := template.New("urlTemplate").Parse(s)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	err = urlTemplate.Execute(&buf, stack)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

func formatString(body []byte, rawFlag bool) string {
	// if requested, try to return the body string formated as json
	if rawFlag {
		return string(body)
	}

	var bodyJsonOutput bytes.Buffer
	err := json.Indent(&bodyJsonOutput, body, "", "  ")
	if err != nil {
		return string(body)
	}

	return bodyJsonOutput.String()
}
