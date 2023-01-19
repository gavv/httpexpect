package httpexpect

import (
	"fmt"
)

// Every matcher struct, e.g. Value, Object, Array, etc. contains a chain instance.
//
// Its most important fields are:
//
//   - AssertionContext: provides test name, current request and response, path
//     to matcher struct relative to chain root
//
//   - AssertionHandler: provides methods to handle successful and failed assertions;
//     may be defined by user, but usually we just use DefaulAssertionHandler
//
//   - Fail bit: set after first failure; never cleared; once it's set,
//     all subsequent failures for this chain will be ignored
//
// When a matcher creates a child matcher, e.g. you call Array.Element() and it returns
// a new Value for given index, it usually looks like this:
//
//	parent.chain.enter("Child()") // appends "Child()" to context.Path
//	defer parent.chain.leave() // removes "Child()" from context.Path
//	return newChild(parent.chain) // calls chain.clone()
//
// In result, child matcher will have a clone of chain, which will inherit context,
// handler, fail bit, and will have "Child()" appended to its context.Path.
//
// This has the following consequences:
//
//   - context and handler are inherited from parent to child chains
//
//   - context.Path automatically contains a path to current chain from root
//
//   - fail bit is inherited as well; if there were a failure in parent chain,
//     subsequent failures will be ignored not only in parent chain, but also
//     in all newly created child chains
type chain struct {
	noCopy noCopy

	context  AssertionContext
	handler  AssertionHandler
	severity AssertionSeverity

	failCb  func()
	failBit bool
}

// Construct chain using config.
func newChainWithConfig(name string, config Config) *chain {
	config.validate()

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

// Construct chain using DefaultAssertionHandler and provided Reporter.
func newChainWithDefaults(name string, reporter Reporter) *chain {
	if reporter == nil {
		panic("Reporter is nil")
	}

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

// Get environment associated with chain
// Chain constructor either gets environment from config or creates a new one.
// Children chains inherit environment.
func (c *chain) env() *Environment {
	return c.context.Environment
}

// Set callback to be invoked on failure.
// Callback is invoked after failure is passed to AssertionHandler.
// Children chains inherit callback.
func (c *chain) setFailCallback(failCb func()) {
	c.failCb = failCb
}

// Set severity of reported failures.
// Chain will override severity with this one.
func (c *chain) setSeverity(severity AssertionSeverity) {
	c.severity = severity
}

// Store request name in AssertionContext.
// Children chains inherit context.
func (c *chain) setRequestName(name string) {
	c.context.RequestName = name
}

// Store request pointer in AssertionContext.
// Children chains inherit context.
func (c *chain) setRequest(req *Request) {
	c.context.Request = req
}

// Store response pointer in AssertionContext.
// Children chains inherit context.
func (c *chain) setResponse(resp *Response) {
	c.context.Response = resp
}

// Create a clone of the chain.
// Modifications of the clone wont affect the original.
func (c *chain) clone() *chain {
	copy := *c //nolint
	copy.context.Path = append(([]string)(nil), c.context.Path...)

	return &copy
}

// Append string to chain path.
func (c *chain) enter(name string, args ...interface{}) {
	c.context.Path = append(c.context.Path, fmt.Sprintf(name, args...))
}

// Replace last element in chain path.
func (c *chain) replace(name string, args ...interface{}) {
	if len(c.context.Path) == 0 {
		panic("unexpected replace")
	}

	c.context.Path[len(c.context.Path)-1] = fmt.Sprintf(name, args...)
}

// Remove last element from chain path.
func (c *chain) leave() {
	if len(c.context.Path) == 0 {
		panic("unpaired enter/leave")
	}

	if !c.failBit {
		c.handler.Success(&c.context)
	}

	c.context.Path = c.context.Path[:len(c.context.Path)-1]
}

// If enabled, chain.fail() will panic on illformed AssertionFailure.
// For tests.
var chainAssertionValidation = false

// Report failure to AssertionHandler and set fail bit.
// If fail bit is already set, failure is ignored.
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

// Set fail bit.
func (c *chain) setFailed() {
	c.failBit = true
}

// Clear fail bit.
// For tests.
func (c *chain) clearFailed() {
	c.failBit = false
}

// Get fail bit.
// For tests.
func (c *chain) failed() bool {
	return c.failBit
}

// Check that chain is not failed.
// Otherwise report failure to Reporter.
// For tests.
func (c *chain) assertNotFailed(r Reporter) {
	if c.failBit {
		r.Errorf("expected: chain is not failed")
	}
}

// Check that chain is failed.
// Otherwise report failure to Reporter.
// For tests.
func (c *chain) assertFailed(r Reporter) {
	if !c.failBit {
		r.Errorf("expected: chain is failed")
	}
}
