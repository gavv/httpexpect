package httpexpect

import "fmt"

// Formatter is used for common formatting options.
type Formatter interface {
	Success(*Context) string
	Failure(*Context, Failure) string
}

// DefaultFormatter is the default Formatter implementation.
type DefaultFormatter struct{}

// Success implements Formatter.Success and returns the current
// test name from context.
func (DefaultFormatter) Success(ctx *Context) string {
	return ctx.TestName + " passed"
}

// Failure implements Formatter.Failure and reports failure.
//
// It formats the info from Context and Failure struct into
// a string and passes it to Context.Reporter for reporting.
func (DefaultFormatter) Failure(ctx *Context, f Failure) string {
	errString := ""
	if f.actual != nil {
		errString = fmt.Sprintf(
			"\nassertion:\n%s\nexpected:\n%s\nactual:\n%s\ndiff:\n%s",
			f.assertionName,
			dumpValue(f.expected),
			dumpValue(f.actual),
			diffValues(f.expected, f.actual),
		)
	} else {
		errString = fmt.Sprintf(
			"expected value not equal to:\n%s",
			dumpValue(f.expected),
		)
	}

	return errString
}
