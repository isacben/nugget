package lexer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/isacben/nugget/interpreter/token"
)

type Lexer struct {
	Input        []rune
	char         rune // current char under examination
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position (after current char)
	line         int  // line number for error reporting
}

// New() creates a pointer to the Lexer
func New(input string) *Lexer {
	l := &Lexer{Input: []rune(input)}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.Input) {
		// End of input: haven't read anything yet or EOF
		// 0 is ASCCII code for "NULL" character
		l.char = 0
	} else {
		l.char = l.Input[l.readPosition]
	}

	l.position = l.readPosition
	l.readPosition++
}

// NextToken switches through the lexer's current char and creates a new token.
// It then it calls readChar() to advance the lexer and it returns the token
func (l *Lexer) NextToken() token.Token {
	var t token.Token

	l.skipWhiteSpace()

	switch l.char {
	case 0:
		t.Literal = ""
		t.Type = token.EOF
		t.Line = l.line
	case '{':
		t = newToken(token.Json, l.line, l.position, l.position+1, l.char)
		jsonStr, err := l.readJson()
		if err != nil {
			fmt.Println(err)
		}
		t.Literal = jsonStr
	default:
		if isValidChar(l.char) {
			t.Start = l.position
			ident := l.readIdentifier()
			t.Literal = ident
			t.Line = l.line
			t.End = l.position

			if isNumber(ident) {
				t.Type = token.Number
				return t
			}

			tokenType, err := token.LookupMethod(ident)
			if err != nil {
				t.Type = token.String
				return t
			}

			t.Type = tokenType
			t.End = l.position
			return t
		}

		t = newToken(token.Ilegal, l.line, 1, 2, l.char)
	}

	// advance to next character
	l.readChar()

	return t
}

func (l *Lexer) skipWhiteSpace() {
	for l.char == ' ' || l.char == '\t' || l.char == '\n' || l.char == '\r' {
		if l.char == '\n' {
			l.line++
		}
		l.readChar()
	}
}

func newToken(tokenType string, line, start, end int, char ...rune) token.Token {
	return token.Token{
		Type:    tokenType,
		Literal: string(char),
		Line:    line,
		Start:   start,
		End:     end,
	}
}

// readString sets a start position and reads through characters
// When it finds a closing `"`, it stops consuming characters and
// returns the string between the start and end positions.
func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar() // this moves the reading position to the next char
		if l.char == ' ' || l.char == 0 {
			break
		}
	}
	return string(l.Input[position:l.position])
}

func (l *Lexer) readJson() (string, error) {
	position := l.position
	braketCount := 1
	for {
		l.readChar() // moves the reading position to the next char
		if l.char == '{' {
			braketCount += 1
		}

		if l.char == '}' {
			braketCount -= 1
		}

		if braketCount == 0 {
			break
		}

		if l.char == 0 {
			return "", fmt.Errorf("expected `}`, got %v", l.char)
		}
	}
	return string(l.Input[position : l.position+1]), nil
}

func isNumber(s string) bool {
	match, _ := regexp.MatchString(`^-?[0-9]\d*(\.\d+)?$`, s)
	return match
}

func isValidChar(char rune) bool {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789;/?:@&=+$,#-_.!~*'()[]\""
	return strings.Contains(chars, string(char))
}

func (l *Lexer) readIdentifier() string {
	position := l.position

	for isValidChar(l.char) {
		l.readChar()
	}

	return string(l.Input[position:l.position])
}
