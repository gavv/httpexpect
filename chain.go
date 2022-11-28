package httpexpect

import (
	"fmt"
)

type chain struct {
	context AssertionContext
	handler AssertionHandler
	isFatal bool
	failCb  func()
	failbit bool
}

func newChainWithConfig(name string, config Config) *chain {
	c := &chain{
		context: AssertionContext{},
		handler: config.AssertionHandler,
		isFatal: true,
		failbit: false,
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
		isFatal: true,
		failbit: false,
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

func (c *chain) setFatal(isFatal bool) {
	c.isFatal = isFatal
}

func (c *chain) setFailCallback(failCb func()) {
	c.failCb = failCb
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
	c.context.Path[len(c.context.Path)-1] = fmt.Sprintf(name, args...)
}

func (c *chain) leave() {
	if len(c.context.Path) == 0 {
		panic("unpaired enter/leave")
	}

	if !c.failbit {
		c.handler.Success(&c.context)
	}

	c.context.Path = c.context.Path[:len(c.context.Path)-1]
}

func (c *chain) fail(failure AssertionFailure) {
	if c.failbit {
		return
	}
	c.failbit = true

	if c.isFatal {
		failure.IsFatal = true
	}

	c.handler.Failure(&c.context, &failure)

	if c.failCb != nil {
		c.failCb()
	}
}

func (c *chain) setFailed() {
	c.failbit = true
}

func (c *chain) failed() bool {
	return c.failbit
}

func (c *chain) reset() {
	c.failbit = false
}

func (c *chain) assertOK(r Reporter) {
	if c.failbit {
		r.Errorf("failbit is true, but should be false")
	}
}

func (c *chain) assertFailed(r Reporter) {
	if !c.failbit {
		r.Errorf("failbit is false, but should be true")
	}
}
