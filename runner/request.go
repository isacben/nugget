package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"strings"
)

type Step struct {
	Name    string
	Method  string
	Url     string
	Http    int
	Headers  []keyValue
	Body    string            
	Captures []keyValue
	Wait    int               
}

var input string

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
	prog := []Step{}

    file, err := os.ReadFile(fileName)
    if err != nil {
        fmt.Printf("could not read file: %v\n", err)
    }

    input = string(file)
    lines := strings.Split(input, "\n")

    parser := NewParser(lines)
    requests := parser.parse()

    for _, request := range(requests) {
        step := Step{
            "",
            request.method,
            request.url,
            request.statusCode,
            request.headers,
            "",
            request.captures,
            request.wait,
        }

        if request.body != nil {
            bodyBytes, _ := json.Marshal(request.body)
            step.Body = string(bodyBytes)
        }

        prog = append(prog, step)
    }


	err = run(prog, rawFlag, header, quiet)
	if err != nil {
		fmt.Printf("%v\n", err)
	}

}
