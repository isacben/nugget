package runner

import (
	"fmt"
	"os"
	"strings"
)

func Execute(fileName string, rawFlag bool, quiet bool) {
    file, err := os.ReadFile(fileName)
    if err != nil {
        fmt.Printf("could not read file: %v\n", err)
        os.Exit(1)
    }

    input := string(file)
    lines := strings.Split(input, "\n")

    parser := NewParser(lines)
    requests := parser.parse()

	err = run(requests, rawFlag, quiet)
	if err != nil {
		fmt.Printf("%v\n", err)
	}

}
