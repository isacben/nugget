package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Prog struct {
	Steps []Step `yaml:"steps"`
}

type Step struct {
	Name   string `yaml:"name"`
	Method string `yaml:"method"`
	URL    string `yaml:"url"`
}

func main() {
	var clientId = os.Getenv("clientId")
	var apiKey = os.Getenv("apiKey")

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

	fmt.Printf("--- prog:\n%v\n\n", prog)

	// dump
	d, err := yaml.Marshal(&prog)
	if err != nil {
		fmt.Printf("error: %v", err)
	}
	fmt.Printf("--- prog dump:\n%s\n\n", string(d))

	fmt.Println(clientId, apiKey)
}
