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

func parse(s string, randVal RandomValues) string {
	urlTemplate, err := template.New("urlTemplate").Parse(s)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	err = urlTemplate.Execute(&buf, randVal)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

func run(prog Prog) {
	token := getToken()
	for _, step := range prog.Steps {
		// prepare struct with values for the template
		randVal := RandomValues{uuid.NewString()}

		fmt.Println(step.Name)

		step.Url = parse(step.Url, randVal)
		fmt.Println(step.Url)

		var reqBody *strings.Reader

		if step.Body != "" {
			step.Body = parse(step.Body, randVal)
			reqBody = strings.NewReader(step.Body)
		} else {
			reqBody = strings.NewReader("")
		}

		//var results interface{}
		//json.Unmarshal([]byte(step.Body.(string)), &results)
		//m := results.(map[string]interface{})
		//fmt.Println(m["hello"].(string))

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

		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Printf("client: could not read response body: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("client: response body: %s\n", body)

		//var f interface{}
		var output map[string]interface{}
		err2 := json.Unmarshal(body, &output)
		if err2 != nil {
			fmt.Println("Error")
		}

		//m := f.(map[string]interface{})
		//fmt.Println(m["id"])

		map2 := map[string]string{
			"customer": output["id"].(string),
		}

		fmt.Println("The id is", map2["customer"])
		fmt.Println("The additional info is", output["additional_info"].(map[string]interface{})["registered_via_social_media"])

		fmt.Println(step.Capture)

		for k, v := range step.Capture {
			fmt.Println(k, "value is", v)
		}
	}
}

func main() {

	// load yaml file
	prog := Prog{}

	yamlFile, err := os.ReadFile("../poget-examples/get-customer.yaml")
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
