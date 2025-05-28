package json

import (
	"encoding/json"
	"fmt"
)

type lineTracker struct {
	data   []byte
	line   int
	column int
	index  int
}

func newLineTracker(data []byte) *lineTracker {
	return &lineTracker{data: data, line: 1, column: 1, index: 0}
}

func (lt *lineTracker) read() (byte, bool) {
	if lt.index >= len(lt.data) {
		return 0, false
	}
	b := lt.data[lt.index]
	lt.index++
	if b == '\n' {
		lt.line++
		lt.column = 1
	} else {
		lt.column++
	}
	return b, true
}

func (lt *lineTracker) getPosition(offset int) (int, int) {
	lt.line = 1
	lt.column = 1
	lt.index = 0
	for i := 0; i < offset; i++ {
		lt.read()
	}
	return lt.line, lt.column
}

func unmarshalWithLineNumber(data []byte, v interface{}) error {
	err := json.Unmarshal(data, v)
	if err != nil {
		if syntaxErr, ok := err.(*json.SyntaxError); ok {
			lt := newLineTracker(data)
			line, column := lt.getPosition(int(syntaxErr.Offset))
			return fmt.Errorf("json syntax error at line %d, column %d: %v", line, column, err)
		}
		return err
	}
	return nil
}

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	invalidJSON := `{"name": "John", "age": 30`            // Missing closing brace
	invalidTypeJSON := `{"name": "John", "age": "thirty"}` // Invalid Type
	var person Person

	err := unmarshalWithLineNumber([]byte(invalidJSON), &person)
	if err != nil {
		fmt.Println(err) // Output: json syntax error at line 1, column 27: unexpected end of JSON input
	}

	err = unmarshalWithLineNumber([]byte(invalidTypeJSON), &person)
	if err != nil {
		fmt.Println(err) // Output: json unmarshal field error at line 1, column 20: json: cannot unmarshal string into Go struct field Person.age of type int
	}
}
