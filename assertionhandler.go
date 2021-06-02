package httpexpect

import "testing"

type AssertionHandler interface {
	// Reporter is implemented for compatibility only and shouldn't be used directly.
	Reporter
	Failure(ctx *Context, failure Failure)
	Success(ctx *Context)
}

type DefaultAssertionHandler struct {
	Reporter  Reporter
	Formatter Formatter
}

func NewDefaultAssertionHandler(t *testing.T) AssertionHandler {
	return DefaultAssertionHandler{
		Reporter:  NewAssertReporter(t),
		Formatter: DefaultFormatter{},
	}
}

func ensureAssertionHandler(config Config) AssertionHandler {
	if config.AssertionHandler != nil {
		return config.AssertionHandler
	}

	if config.Reporter == nil {
		panic("compat Reporter is nil: you should provide an AssertionHandler")
	}

	return DefaultAssertionHandler{
		Reporter:  config.Reporter,
		Formatter: DefaultFormatter{},
	}
}

func (d DefaultAssertionHandler) Errorf(message string, args ...interface{}) {
	d.Reporter.Errorf(message, args...)
}

func (d DefaultAssertionHandler) Failure(ctx *Context, failure Failure) {
	d.Errorf(d.Formatter.Failure(ctx, failure))
}

func (d DefaultAssertionHandler) Success(ctx *Context) {}
