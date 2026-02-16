package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
    "strconv"
)

type Parser struct {
	lines   []string
	current int
}

type Json any
type Request struct {
	method     string
	url        string
	headers    []keyValue
	captures   []keyValue
	statusCode int
	wait       int
	body       Json
}

type keyValue struct {
	key   string
	value string
}

func NewParser(lines []string) *Parser {
	p := &Parser{lines: lines}
	return p
}

func (p *Parser) parse() []Request {
	var requests []Request
	for p.current < len(p.lines) {
		left, right := p.parseLine()

		if !isKeyword(left) {
			p.error(left, false, "Uknown symbol.")
		}

		if !isMethod(left) {
			p.error(left, false, "Expecting http method.")
		}

		if isEmpty(right) {
			p.error(left, true, "Expecting url.")
		}

		requests = append(requests, p.request(left, right))
	}

	return requests
}

func (p *Parser) request(method, url string) Request {
	var req Request

	req.method = strings.TrimSpace(method)
	req.url = strings.TrimSpace(url)

	for p.current < len(p.lines) {
		left, right := p.parseLine()
		if isEmpty(left) {
			break
		}

		left = strings.ToLower(left)

		if left[0] == '{' || left[0] == '[' {
			req.body = p.body()
			continue
		}

		if isMethod(left) {
			p.current--
			return req
		}

		switch left {
		case "http":
			req.statusCode = p.number(left, right)

		case "header":
			req.headers = append(req.headers, p.pair(left, right))

		case "wait":
			req.wait = p.number(left, right)

		case "save":
			req.captures = append(req.captures, p.pair(left, right))

		default:
			p.error(left, false, "Unknown sysmbol.")
		}
	}

	return req
}

func (p *Parser) number(left, right string) int {
    if isEmpty(right) {
        p.error(left, true, "Expected number.")
    }

    num, err := strconv.Atoi(strings.TrimSpace(right))
    if err != nil {
        p.error(strings.TrimSpace(right), false, "Expected number.")
    }

    return num
}

func (p *Parser) pair(left, right string) keyValue {
    if isEmpty(right) {
        p.error(left, true, "Expected identifier.")
    }

    arguments := strings.TrimSpace(right)
	parts := strings.SplitN(arguments, " ", 2)

	if !isValid(parts) {
		p.error(parts[0], true, "Expected value.")
	}

	return keyValue{parts[0], parts[1]}
}

func (p *Parser) body() Json {
	startLine := p.current - 1 // Store the start line (0-based)
	var jsonLines []string
	jsonLines = append(jsonLines, p.lines[p.current-1])

	for p.current < len(p.lines) {
		left, _ := p.parseLine()
		if isEmpty(left) || isKeyword(left) {
			p.current--
			break
		}

		jsonLines = append(jsonLines, p.lines[p.current-1])
	}

	jsonString := strings.Join(jsonLines, "\n")

	var jsonData Json
	err := json.Unmarshal([]byte(jsonString), &jsonData)

	// TODO (isaac): improve this code
	if err != nil {
		if syntaxErr, ok := err.(*json.SyntaxError); ok {
			// Get line and column within the JSON string
			errorLine, errorCol := getLineAndColumn([]byte(jsonString), syntaxErr.Offset)

			// Map back to the actual line in the original input
			actualLineIndex := startLine + errorLine - 1

			// Bounds check
			if actualLineIndex >= 0 && actualLineIndex < len(p.lines) {
				p.current = actualLineIndex + 1 // Set parser to the error line

				currentLine := p.lines[actualLineIndex]

				// For error positioning, we want to point to the exact column
				var errorPart string
				if errorCol > 1 && errorCol-1 <= len(currentLine) {
					errorPart = currentLine[:errorCol-1]
					fmt.Println(errorPart)
				} else {
					// If column calculation fails, use the whole line
					errorPart = currentLine
				}

				p.error(errorPart, true, "JSON syntax error: "+err.Error())
			} else {
				// Fallback if line calculation fails
				fmt.Printf("JSON syntax error: %s\n", err.Error())
				os.Exit(1)
			}
		} else {
			// Handle other JSON errors
			fmt.Printf("JSON error: %s\n", err.Error())
			os.Exit(1)
		}
	}

	return jsonData
}

func getLineAndColumn(data []byte, offset int64) (line, col int) {
	line = 1
	col = 1

	for i := int64(0); i < offset && i < int64(len(data)); i++ {
		if data[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}

	return line, col
}

func (p *Parser) parseLine() (string, string) {
	for p.current < len(p.lines) {
		s := strings.TrimSpace(p.lines[p.current])
		p.current++

		if s == "" {
			continue
		}

		if s[0] == '#' {
			continue
		}

		parts := strings.SplitN(s, " ", 2)
		if isValid(parts) {
			return parts[0], parts[1]
		}

		return parts[0], ""
	}

	return "", ""
}

func isMethod(param string) bool {
	methods := []string{
		"get", "post", "put", "delete", "patch",
	}

	return slices.Contains(methods, strings.ToLower(param))
}

func isKeyword(param string) bool {
	if isMethod(param) {
		return true
	}

	keywords := []string{
		"header", "http", "wait", "save",
	}
	return slices.Contains(keywords, strings.ToLower(param))
}

func isValid(parts []string) bool {
	return len(parts) > 1
}

func isEmpty(str string) bool {
	return str == ""
}

func (p *Parser) error(part string, offset bool, message string) {
	line := p.lines[p.current-1]

    chars := 1
	if offset {
		chars = len(part)
	}

	byteIndex := strings.Index(strings.ToLower(line), strings.ToLower(part))

	if byteIndex == -1 {
		fmt.Println("Unknown error.")
	}

	column := byteIndex + chars

	fmt.Printf("Invalid syntax at position %d.\n", column)

	prefix := fmt.Sprintf("   %d | ", p.current)

	fmt.Printf("%s%s\n", prefix, line)
	fmt.Print(strings.Repeat(" ", len(prefix)+column-1))
	fmt.Println("^")
	fmt.Println(message)

	os.Exit(1)
}
