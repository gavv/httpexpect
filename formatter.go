package httpexpect

import (
	"fmt"
)

// Formatter is used for common formatting options.
type Formatter interface {
	Success(*Context) string
	Failure(*Context, Failure) string
}

// DefaultFormatter is the default Formatter implementation.
type DefaultFormatter struct{}

// Success implements Formatter.Success and returns an empty string.
func (DefaultFormatter) Success(ctx *Context) string {
	return ""
}

// Failure implements Formatter.Failure and reports failure.
//
// It formats the info from Context and Failure struct into
// a string and passes it to Context.Reporter for reporting.
func (DefaultFormatter) Failure(ctx *Context, f Failure) string {
	errString := ""
	if f.err != nil {
		errString += "\noriginal error: " + f.err.Error()
	}

	// FIXME: implement all failureAssert* cases
	errString += fmt.Sprintf(
		"\nassertion: %s\nexpected: %s\nactual: %s\ndiff:\n%s",
		f.assertionName,
		dumpValue(f.expected),
		dumpValue(f.actual),
		diffValues(f.expected, f.actual),
	)

	return errString
}
