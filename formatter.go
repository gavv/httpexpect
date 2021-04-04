package httpexpect

import "fmt"

// Formatter is used for common formatting options.
type Formatter interface {
	BeginAssertion(Context)
	Success(Context)
	Failure(Context, Failure)
	EndAssertion(Context)
}

// DefaultFormatter is the default Formatter implementation.
type DefaultFormatter struct{}

// BeginAssertion implements Formatter.BeginAssertion.
//
// It is a no-op for now. Actual implementation may be
// added later.
func (DefaultFormatter) BeginAssertion(ctx Context) {}

// Success implements Formatter.Success.
//
// It is a no-op for now. Actual implementation may be
// added later.
func (DefaultFormatter) Success(ctx Context) {}

// Failure implements Formatter.Failure and reports failure.
//
// It formats the info from Context and Failure struct into
// a string and passes it to Context.Reporter for reporting.
func (DefaultFormatter) Failure(ctx Context, f Failure) {
	errString := ""
	if f.actual != nil {
		errString = fmt.Sprintf(
			"\nexpected:\n%s\nactual:\n%s\ndiff:\n%s",
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

	ctx.Reporter.Errorf(errString)
}

// EndAssertion implements Formatter.EndAssertion.
//
// It is a no-op for now. Actual implementation may be
// added later.
func (DefaultFormatter) EndAssertion(ctx Context) {}
