package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

type Prog struct {
	Steps []Step `yaml:"steps"`
}

type Step struct {
	Name   string `yaml:"name"`
	Method string `yaml:"method"`
	Url    string `yaml:"url"`
}

type AuthJson struct {
	Expires_at string `json:"expires_at"`
	Token      string `json:"token"`
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

func run(prog Prog) {
	token := getToken()
	for _, step := range prog.Steps {
		fmt.Println(step.Name)

		req, err := http.NewRequest(step.Method, step.Url, nil)
		authHeader := fmt.Sprintf("Bearer %s", token)
		req.Header = http.Header{
			"Content-Type":  {"application/json"},
			"Authorization": {authHeader},
		}

		// req.Header.Add("x-client-id", clientId)
		// req.Header.Add("x-api-key", apiKey)

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

	fmt.Println(clientId, apiKey)
}
