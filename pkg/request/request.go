package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/itchyny/gojq"
	"gopkg.in/yaml.v3"
)

type Step struct {
	Name    string            `yaml:"name"`
	Method  string            `yaml:"method"`
	Url     string            `yaml:"url"`
	Http    int               `yaml:"http"`
	Header  map[string]string `yaml:"header"`
	Body    string            `json:"body"`
	Capture map[string]string `yaml:"capture"`
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

func run(prog []Step, rawFlag bool, headerFlag bool, quiet bool) error {
	token, err := getToken()
	if err != nil {
		return fmt.Errorf("authentication error: %s", err)
	}

	// prepare stack map with values for the template
	stack := make(map[string]string)

	for _, step := range prog {
		stack["uuid"] = uuid.NewString()

		// TODO: fix parse() panic
		step.Url = parse(step.Url, stack)

		reqBody := strings.NewReader("")
		if step.Body != "" {
			step.Body = parse(step.Body, stack)
			reqBody = strings.NewReader(step.Body)
		}

		req, err := http.NewRequest(step.Method, step.Url, reqBody)
		if err != nil {
			return fmt.Errorf("client: could not create request: %s", err)
		}

		authHeader := fmt.Sprintf("Bearer %s", token)
		req.Header = http.Header{
			"Content-Type":  {"application/json"},
			"Authorization": {authHeader},
		}

		if step.Header != nil {
			for k, v := range step.Header {
				v = parse(v, stack)
				req.Header.Add(k, v)
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
			fmt.Printf("step: %s\n", step.Name)
			fmt.Printf("response data:\n")
		}

		// print response data
		fmt.Println(formatString(body, rawFlag))

		if headerFlag {
			resHeader, err := json.Marshal(res.Header)
			if err != nil {
				return fmt.Errorf("client: could not process response header: %s", err)
			}

			if !quiet {
				fmt.Printf("response header:\n")
			}

			fmt.Println(formatString(resHeader, rawFlag))
		}

		if step.Http > 0 {
			respCode := res.StatusCode
			if respCode != step.Http {
				return fmt.Errorf("\033[0;31merror\033[0m: expected %v but got %v", step.Http, res.StatusCode)
			}

			if !quiet {
				fmt.Printf("\u001b[32;1msuccess\033[0m: status code is %v\n", res.StatusCode)
			}
		}

		// TODO: simplify this and remove from function
		if step.Capture != nil {
			for key, val := range step.Capture {
				// key = variable name from the capture: customer_id from "customer_id: .id" in the yaml file
				// val = the string to query the response json: .id from "customer_id: .id" in the yaml file
				// do the jq query on the response body string
				query, err := gojq.Parse(val)
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
					stack[key] = fmt.Sprintf("%v", v)
				}
			}
		}
	}
	return nil
}

func Execute(fileName string, rawFlag bool, header bool, quiet bool) {
	prog := []Step{}

	yamlFile, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Printf("could not read file: %v\n", err)
	}

	err = yaml.Unmarshal(yamlFile, &prog)
	if err != nil {
		fmt.Printf("invalid file format: %v\n", err)
		os.Exit(1)
	}

	errs := validate(prog)
	if errs != nil {
		for _, err := range errs {
			fmt.Println(err)
		}
		os.Exit(1)
	}

	rerr := run(prog, rawFlag, header, quiet)
	if rerr != nil {
		fmt.Printf("%v\n", rerr)
	}

}
