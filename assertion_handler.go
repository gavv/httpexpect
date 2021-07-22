package httpexpect

// AssertionHandler takes care of reporting test Failure or Success.
//
// It implements Reporter for compatibility reason and must not be used for new code.
//
// When implementing your own AssertionHandler, you can chose to ignore those
// Reporter and Formatter interfaces/implementations and decide to handle
// everything by yourself instead.
//
// For example, you can provide a JSON output for ulterior processing.
//
// You should avoid "expensive I/O" in the implementation. Instead, write results
// to a "local" sink and publish those results from an external tool.
//
// For example: run the tests in your CI, but have a separate step to publish
// test results from said JSON file.
//
// This way you don't mix testing and reporting.
type AssertionHandler interface {
	// Reporter is implemented for compatibility only and shouldn't be used directly.
	Reporter
	Failure(ctx *Context, failure Failure)
	Success(ctx *Context)
}

// ChainAssertionHandler will call all Handlers one after another, as they are provided.
type ChainAssertionHandler struct {
	Handlers []AssertionHandler
}

func (c ChainAssertionHandler) Errorf(message string, args ...interface{}) {
	for _, h := range c.Handlers {
		h.Errorf(message, args...)
	}
}

func (c ChainAssertionHandler) Failure(ctx *Context, failure Failure) {
	for _, h := range c.Handlers {
		h.Failure(ctx, failure)
	}
}

func (c ChainAssertionHandler) Success(ctx *Context) {
	for _, h := range c.Handlers {
		h.Success(ctx)
	}
}

type DefaultAssertionHandler struct {
	Reporter  Reporter
	Formatter Formatter
}

// NewDefaultAssertionHandler uses AssertReporter and DefaultFormatter.
// The Formatter is called first and provides the string to Reporter.
func NewDefaultAssertionHandler(t LoggerReporterNamer) AssertionHandler {
	return newDefaultAssertionHandler(t)
}

func newDefaultAssertionHandler(t LoggerReporter) AssertionHandler {
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
	// FIXME: eventually, create a Failure and use d.Failure instead?
	d.Reporter.Errorf(message, args...)
}

func (d DefaultAssertionHandler) Failure(ctx *Context, failure Failure) {
	d.Errorf(d.Formatter.Failure(ctx, failure))
}

func (d DefaultAssertionHandler) Success(ctx *Context) {}
