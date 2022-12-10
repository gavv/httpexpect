package httpexpect

import (
	"fmt"
)

type chain struct {
	context  AssertionContext
	handler  AssertionHandler
	severity AssertionSeverity
	failCb   func()
	failBit  bool
}

func newChainWithConfig(name string, config Config) *chain {
	c := &chain{
		context:  AssertionContext{},
		handler:  config.AssertionHandler,
		severity: SeverityError,
		failBit:  false,
	}

	c.context.TestName = config.TestName

	if name != "" {
		c.context.Path = []string{name}
	} else {
		c.context.Path = []string{}
	}

	if config.Environment != nil {
		c.context.Environment = config.Environment
	} else {
		c.context.Environment = newEnvironment(c)
	}

	return c
}

func newChainWithDefaults(name string, reporter Reporter) *chain {
	c := &chain{
		context: AssertionContext{},
		handler: &DefaultAssertionHandler{
			Formatter: &DefaultFormatter{},
			Reporter:  reporter,
		},
		severity: SeverityError,
		failBit:  false,
	}

	if name != "" {
		c.context.Path = []string{name}
	} else {
		c.context.Path = []string{}
	}

	c.context.Environment = newEnvironment(c)

	return c
}

func (c *chain) getEnv() *Environment {
	return c.context.Environment
}

func (c *chain) setFailCallback(failCb func()) {
	c.failCb = failCb
}

func (c *chain) setSeverity(severity AssertionSeverity) {
	c.severity = severity
}

func (c *chain) setRequestName(name string) {
	c.context.RequestName = name
}

func (c *chain) setRequest(req *Request) {
	c.context.Request = req
}

func (c *chain) setResponse(resp *Response) {
	c.context.Response = resp
}

func (c *chain) clone() *chain {
	ret := *c

	ret.context.Path = nil
	ret.context.Path = append(ret.context.Path, c.context.Path...)

	return &ret
}

func (c *chain) enter(name string, args ...interface{}) {
	c.context.Path = append(c.context.Path, fmt.Sprintf(name, args...))
}

func (c *chain) replace(name string, args ...interface{}) {
	if len(c.context.Path) == 0 {
		panic("unexpected replace")
	}

	c.context.Path[len(c.context.Path)-1] = fmt.Sprintf(name, args...)
}

func (c *chain) leave() {
	if len(c.context.Path) == 0 {
		panic("unpaired enter/leave")
	}

	if !c.failBit {
		c.handler.Success(&c.context)
	}

	c.context.Path = c.context.Path[:len(c.context.Path)-1]
}

var chainAssertionValidation = false

func (c *chain) fail(failure AssertionFailure) {
	if c.failBit {
		return
	}
	c.failBit = true

	failure.Severity = c.severity
	if c.severity == SeverityError {
		failure.IsFatal = true
	}

	c.handler.Failure(&c.context, &failure)

	if c.failCb != nil {
		c.failCb()
	}

	if chainAssertionValidation {
		if err := validateAssertion(&failure); err != nil {
			panic(err)
		}
	}
}

func (c *chain) setFailed() {
	c.failBit = true
}

func (c *chain) failed() bool {
	return c.failBit
}

func (c *chain) reset() {
	c.failBit = false
}

func (c *chain) assertOK(r Reporter) {
	if c.failBit {
		r.Errorf("failbit is true, but should be false")
	}
}

func (c *chain) assertFailed(r Reporter) {
	if !c.failBit {
		r.Errorf("failbit is false, but should be true")
	}
}
