package parser

import "fmt"

// ParseError returned when the input is an invalid model
type ParseError struct {
	Message string
	Err     error
}

// Error formats the ParseError as a printable string
func (e *ParseError) Error() string {
	return e.Message + ": " + e.Err.Error()
}

// NewParseError initializes a new error which happened during
// parsing and keeps the context of the model which is being
// processed
func NewParseError(name string, err error) *ParseError {
	return &ParseError{
		Message: fmt.Sprintf("Error parsing model '%s'", name),
		Err:     err,
	}
}
