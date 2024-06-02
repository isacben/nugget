package main

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

type Prog struct {
	Steps []Step `yaml:"steps"`
}

type Step struct {
	Name    string            `yaml:"name"`
	Method  string            `yaml:"method"`
	Url     string            `yaml:"url"`
	Body    string            `json:"body"`
	Capture map[string]string `yaml:"capture"`
}

type AuthJson struct {
	Expires_at string `json:"expires_at"`
	Token      string `json:"token"`
}

type RandomValues struct {
	Uuid string
}

type Stack struct {
	Uuid         string
	CapturedVals map[string]string
}

var clientId = os.Getenv("clientId")
var apiKey = os.Getenv("apiKey")

func validate(prog Prog) {
	err := false

	for k, step := range prog.Steps {
		if step.Name == "" {
			fmt.Printf(">>> Error - <name> missing in step %v\n", k+1)
			err = true
		}
		if step.Method == "" {
			fmt.Printf(">>> Error - <method> missing in step %v\n", k+1)
			err = true
		}
		if step.Url == "" {
			fmt.Printf(">>> Error - <url> missing in step %v\n", k+1)
			err = true
		}
	}

	if err {
		os.Exit(1)
	}
}

func getToken() string {
	requestURL := fmt.Sprintf("%s/api/v1/authentication/login", os.Getenv("apiUrl"))
	req, err := http.NewRequest("POST", requestURL, nil)

	req.Header = http.Header{
		"Content-Type": {"application/json"},
		"x-client-id":  {clientId},
		"x-api-key":    {apiKey},
	}

	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
		os.Exit(1)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		os.Exit(1)
	}

	defer res.Body.Close()

	var authJson AuthJson
	if err := json.NewDecoder(res.Body).Decode(&authJson); err != nil {
		panic(err)
	}

	return authJson.Token
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

func run(prog Prog) {
	token := getToken()

	// prepare struct with values for the template
	//randVal := RandomValues{uuid.NewString()}
	stack := Stack{
		CapturedVals: make(map[string]string),
	}

	stack2 := make(map[string]string)

	for _, step := range prog.Steps {

		stack.Uuid = uuid.NewString()
		stack2["uuid"] = uuid.NewString()

		fmt.Println(step.Name)

		step.Url = parse(step.Url, stack2)
		fmt.Println(step.Url)

		var reqBody *strings.Reader

		if step.Body != "" {
			step.Body = parse(step.Body, stack2)
			reqBody = strings.NewReader(step.Body)
		} else {
			reqBody = strings.NewReader("")
		}

		req, err := http.NewRequest(step.Method, step.Url, reqBody)
		authHeader := fmt.Sprintf("Bearer %s", token)
		req.Header = http.Header{
			"Content-Type":  {"application/json"},
			"Authorization": {authHeader},
		}

		if err != nil {
			fmt.Printf("client: could not create request: %s\n", err)
			os.Exit(1)
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("client: error making http request: %s\n", err)
			os.Exit(1)
		}

		// Response body
		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Printf("client: could not read response body: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("response body: %s\n", body)

		// if there are captures
		// add the jq query result to the captures in a loop
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
				err4 := json.Unmarshal(body, &bodyAny)
				if err4 != nil {
					panic(err4)
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
					// fmt.Printf("%s => %v\n", key, v)
					stack.CapturedVals[key] = fmt.Sprintf("%v", v)
					stack2[key] = fmt.Sprintf("%v", v)
					//fmt.Printf("The stack value: %v\n", stack.CapturedVals[key])
				}

				// stack.CapturedVals[key] = val
			}
		}
	}
}

func main() {

	// load yaml file
	prog := Prog{}

	yamlFile, err := os.ReadFile("../poget-examples/create-customer.yaml")
	if err != nil {
		fmt.Printf("yamlFile.Get err #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &prog)
	if err != nil {
		fmt.Printf("Unmarshal: %v", err)
	}

	validate(prog)
	run(prog)

}
