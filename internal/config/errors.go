package config

import (
	"fmt"
	"strings"

	"github.com/rotisserie/eris"
)

// ErrorNoConfig indicates the config file doesn't exist
var ErrorNoConfig = eris.New("No config file")

// ErrorInvalidConfig indicates the config file has one or
// multiple errors.
type ErrorInvalidConfig struct {
	Errors []error
}

func (err ErrorInvalidConfig) Error() string {
	if len(err.Errors) == 0 {
		return ""
	}
	if len(err.Errors) == 1 {
		return err.Errors[0].Error()
	}

	var sb strings.Builder
	for i, e := range err.Errors {
		fmt.Fprintf(&sb, "%d: %s\n", i, e.Error())
	}
	return sb.String()
}

func (err ErrorInvalidConfig) Unwrap() []error {
	return err.Errors
}

func (err *ErrorInvalidConfig) Append(errs ...error) {
	err.Errors = append(err.Errors, errs...)
}
