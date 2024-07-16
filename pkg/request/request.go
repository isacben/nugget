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

type Prog struct {
	Steps []Step `yaml:"steps"`
}

type Step struct {
	Name    string            `yaml:"name"`
	Method  string            `yaml:"method"`
	Url     string            `yaml:"url"`
	Header  map[string]string `yaml:"header"`
	Body    string            `json:"body"`
	Capture map[string]string `yaml:"capture"`
}

type AuthJson struct {
	Expires_at string `json:"expires_at"`
	Token      string `json:"token"`
}

type Output struct {
	Name       string              `json:"name"`
	Url        string              `json:"url"`
	ReqBody    json.RawMessage     `json:"request_body"`
	ReqHeaders map[string][]string `json:"request_header"`
	ResHeaders map[string][]string `json:"response_header"`
	ResBody    json.RawMessage     `json:"response_body"`
}

var clientId = os.Getenv("clientId")
var apiKey = os.Getenv("apiKey")

func validate(prog Prog) {
	err := false

	if prog.Steps == nil {
		fmt.Printf("no steps found\n")
		os.Exit(1)
	}

	for k, step := range prog.Steps {
		if step.Name == "" {
			fmt.Printf("missing keyword in step %v: name\n", k+1)
			err = true
		}
		if step.Method == "" {
			fmt.Printf("missing keyword in step %v: method\n", k+1)
			err = true
		}
		if step.Url == "" {
			fmt.Printf("missing keyword in step %v: url\n", k+1)
			err = true
		}
		if step.Body != "" {
			var v interface{}
			data := []byte(step.Body)
			jerr := json.Unmarshal(data, &v)
			if jerr != nil {
				fmt.Printf("syntax error in step %v body near: `%s`\n", k+1, string(data[jerr.(*json.SyntaxError).Offset-1:]))
				err = true
			}
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

	// prepare stack map with values for the template
	stack := make(map[string]string)

	for _, step := range prog.Steps {

		stack["uuid"] = uuid.NewString()

		fmt.Printf("%s\n", step.Name)
		step.Url = parse(step.Url, stack)
		//fmt.Printf("%s\n", step.Url)

		var reqBody *strings.Reader

		if step.Body != "" {
			step.Body = parse(step.Body, stack)
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

		if step.Header != nil {
			for key, val := range step.Header {
				val = parse(val, stack)
				req.Header.Add(key, val)
			}
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

		// Response body
		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Printf("client: could not read response body: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("%v\n", string(body))
		//output.ResHeaders = res.Header

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
}

func PrintErr(errors []error) {
	type Output struct {
		Status string   `json:"status"`
		Data   []string `json:"data"`
	}

	output := Output{
		Status: "error",
		Data:   []string{},
	}

	for _, myerr := range errors {
		output.Data = append(output.Data, myerr.Error())
	}

	outputPrint, err := json.Marshal(output)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Print(string(outputPrint))
}

func Execute(fileName string) {

	// load yaml file
	prog := Prog{}

	yamlFile, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Printf("yamlFile.Get err #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &prog)
	if err != nil {
		fmt.Printf("invalid file format: %v\n", err)
		os.Exit(1)
	}

	validate(prog)
	run(prog)

}
