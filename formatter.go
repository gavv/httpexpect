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

// Success implements Formatter.Success and returns the test name.
func (DefaultFormatter) Success(ctx *Context) string {
	return ctx.TestName
}

// Failure implements Formatter.Failure and reports failure.
//
// It formats the info from Context and Failure struct into
// a string and passes it to Context.Reporter for reporting.
func (DefaultFormatter) Failure(ctx *Context, f Failure) string {
	errString := ""
	if f.OriginalError != nil {
		errString += "\noriginal error: " + f.OriginalError.Error()
	}

	// FIXME: implement all failureAssert* cases
	errString += fmt.Sprintf(
		"\nassertion: %s\nexpected: %s\nactual: %s\ndiff:\n%s",
		f.AssertionName,
		dumpValue(f.Expected),
		dumpValue(f.Actual),
		diffValues(f.Expected, f.Actual),
	)

	return errString
}
