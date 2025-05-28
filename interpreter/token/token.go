package token

import "fmt"

// Types of tokens
const (
	// Unrecognize token or character
	Ilegal string = "ILEGAL"

	// End of file
	EOF string = "EOF"

	// Literals
	String string = "STRING"
	Number string = "NUMBER"

	// Structural tokens
	LeftBrace  string = "{"
	RightBrace string = "}"
	Whitespace string = "WHITESPACE"
	NewLine    string = "NEWLINE"

	// Comments
	Comment string = "COMMENT"

	// Content
	Json string = "JSON"

	// Methods
	Post string = "POST"
	Get  string = "GET"

	// Response
	Http    string = "HTTP"
	Capture string = "CAPTURE"
)

type Token struct {
	Type    string
	Literal string
	Line    int
	Start   int
	End     int
}

var validKeywords = map[string]string{
	"POST":      Post,
	"GET":       Get,
	"HTTP":      Http,
	"[Capture]": Capture,
}

func LookupMethod(identifier string) (string, error) {
	if token, ok := validKeywords[identifier]; ok {
		return token, nil
	}
	return "", fmt.Errorf("expected a valid method, found: %s", identifier)
}
