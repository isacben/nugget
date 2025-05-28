package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"

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

func Execute(fileName string, rawFlag bool, header bool, quiet bool, parserFlag bool) {
	if parserFlag {
		fmt.Println("Using parser (experimental)")
		return
	}

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
