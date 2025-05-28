package runner

import (
	"encoding/json"
	"fmt"
)

func validate(prog []Step) []error {
	var errs []error

	if prog == nil {
		errs = append(errs, fmt.Errorf("no steps found"))
	}

	for k, step := range prog {
		if step.Name == "" {
			errs = append(errs, fmt.Errorf("missing keyword in step %v: name", k+1))
		}
		if step.Method == "" {
			errs = append(errs, fmt.Errorf("missing keyword in step %v: method", k+1))
		}
		if step.Url == "" {
			errs = append(errs, fmt.Errorf("missing keyword in step %v: url", k+1))
		}
		if step.Body != "" {
			var v interface{}
			data := []byte(step.Body)
			jerr := json.Unmarshal(data, &v)
			if jerr != nil {
				err := fmt.Errorf("syntax error in step %v: body near: %s", k+1, string(data[jerr.(*json.SyntaxError).Offset-1:]))
				errs = append(errs, err)
			}
		}
	}

	return errs
}
