package httpexpect

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Every matcher struct, e.g. Value, Object, Array, etc. contains a chain instance.
//
// Most important chain fields are:
//
//   - AssertionContext: provides test name, current request and response, and path
//     to current assertion starting from chain root
//
//   - AssertionHandler: provides methods to handle successful and failed assertions;
//     may be defined by user, but usually we just use DefaulAssertionHandler
//
//   - AssertionSeverity: severity to be used for failures (fatal or non-fatal)
//
//   - Reference to parent: every chain remembers its parent chain; on failure,
//     chain automatically marks its parents failed
//
//   - Failure flags: flags indicating whether a failure occurred on chain, or
//     on any of its children
//
// Chains are linked into a tree. Child chain corresponds to nested matchers
// and assertions. For example, when the user invokes:
//
//	e.GET("/test").Expect().JSON().Equal(...)
//
// each nested call (GET, Expect, JSON, Equal) will create a child chain.
//
// There are two ways to create a child chain:
//
//   - use enter() / leave()
//   - use clone()
//
// enter() creates a chain to be used during assertion. After calling enter(), you
// can use fail() to report any failures, which will pass it to AssertionHandler
// and mark chain as failed.
//
// After assertion is done, you should call leave(). If there were no failures,
// leave() will notify AssertionHandler about succeeded assertion. Otherwise,
// leave() will mark its parent as failed and notify grand-, grand-grand-, etc
// parents that they have failed children.
//
// If the assertion wants to create child matcher struct, it should invoke clone()
// after calling enter() and before calling leave().
//
// enter() receives assertion name as an argument. This name is appended to the
// path in AssertionContext. If you call clone() on this chain, it will inherit
// this path. This way chain maintains path of the nested assertions.
//
// Typical workflow looks like:
//
//	// create temporary chain for assertion
//	opChain := array.chain.enter("AssertionName()")
//
//	// optional: report assertion failure
//	opChain.fail(...)
//
//	// optional: create child matcher
//	child := &Value{chain: opChain.clone(), ...}
//
//	// if there was a failure, propagate it back to array.chain and notify
//	// parents of array.chain that they have failed children
//	opChain.leave()
type chain struct {
	mu sync.Mutex

	parent *chain
	state  chainState
	flags  chainFlags

	context  AssertionContext
	handler  AssertionHandler
	severity AssertionSeverity
}

// If enabled, chain will panic if used incorrectly or gets illformed AssertionFailure.
// Used only in our own tests.
var chainValidation = false

type chainState int

const (
	stateCloned  chainState = iota // chain was created using clone()
	stateEntered                   // chain was created using enter()
	stateLeaved                    // leave() was called
)

type chainFlags int

const (
	flagFailed         chainFlags = (1 << iota) // fail() was called on this chain
	flagFailedChildren                          // fail() was called on any child
)

// Construct chain using config.
func newChainWithConfig(name string, config Config) *chain {
	config.validate()

	c := &chain{
		context:  AssertionContext{},
		handler:  config.AssertionHandler,
		severity: SeverityError,
	}

	c.context.TestName = config.TestName

	if name != "" {
		c.context.Path = []string{name}
		c.context.AliasedPath = []string{name}
	} else {
		c.context.Path = []string{}
		c.context.AliasedPath = []string{}
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
	}

	if name != "" {
		c.context.Path = []string{name}
		c.context.AliasedPath = []string{name}
	} else {
		c.context.Path = []string{}
		c.context.AliasedPath = []string{}
	}

	c.context.Environment = newEnvironment(c)

	return c
}

// Get environment instance.
// Root chain constructor either gets environment from config or creates a new one.
// Child chains inherit environment from parent.
func (c *chain) env() *Environment {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.context.Environment
}

// Make this chain to be root.
// Chain's parent field is cleared.
// Failures wont be propagated to the upper chains anymore.
func (c *chain) setRoot() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if chainValidation && c.state == stateLeaved {
		panic("can't use chain after leave")
	}

	c.parent = nil
}

// Set severity of reported failures.
// Chain always overrides failure severity with configured one.
func (c *chain) setSeverity(severity AssertionSeverity) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if chainValidation && c.state == stateLeaved {
		panic("can't use chain after leave")
	}

	c.severity = severity
}

// Store request name in AssertionContext.
// Child chains inherit context from parent.
func (c *chain) setRequestName(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if chainValidation && c.state == stateLeaved {
		panic("can't use chain after leave")
	}

	c.context.RequestName = name
}

// Store request pointer in AssertionContext.
// Child chains inherit context from parent.
func (c *chain) setRequest(req *Request) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if chainValidation && c.state == stateLeaved {
		panic("can't use chain after leave")
	}

	if chainValidation && c.context.Request != nil {
		panic("context.Request already set")
	}

	c.context.Request = req
}

// Store response pointer in AssertionContext.
// Child chains inherit context from parent.
func (c *chain) setResponse(resp *Response) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if chainValidation && c.state == stateLeaved {
		panic("can't use chain after leave")
	}

	if chainValidation && c.context.Response != nil {
		panic("context.Response already set")
	}

	c.context.Response = resp
}

// Create chain clone.
// Typically is called between enter() and leave().
func (c *chain) clone() *chain {
	c.mu.Lock()
	defer c.mu.Unlock()

	if chainValidation && c.state == stateLeaved {
		panic("can't use chain after leave")
	}

	contextCopy := c.context
	contextCopy.Path = append(([]string)(nil), contextCopy.Path...)
	contextCopy.AliasedPath = append(([]string)(nil), c.context.AliasedPath...)

	return &chain{
		parent: c,
		state:  stateCloned,
		// since the new clone doesn't have children yet, flagFailedChildren
		// is not inherited
		flags:    (c.flags & ^flagFailedChildren),
		context:  contextCopy,
		handler:  c.handler,
		severity: c.severity,
	}
}

// Create temporary chain clone to be used in assertion.
// If name is not empty, it is appended to the path.
// You must call leave() at the end of assertion.
func (c *chain) enter(name string, args ...interface{}) *chain {
	chainCopy := c.clone()

	chainCopy.state = stateEntered
	if name != "" {
		chainCopy.context.Path = append(chainCopy.context.Path, fmt.Sprintf(name, args...))
		chainCopy.context.AliasedPath =
			append(c.context.AliasedPath, fmt.Sprintf(name, args...))
	}

	return chainCopy
}

// Like enter(), but it replaces last element of the path instead appending to it.
// Must be called between enter() and leave().
func (c *chain) replace(name string, args ...interface{}) *chain {
	if chainValidation {
		func() {
			c.mu.Lock()
			defer c.mu.Unlock()

			if c.state != stateEntered {
				panic("replace allowed only between enter/leave")
			}
			if len(c.context.Path) == 0 {
				panic("replace allowed only if path is non-empty")
			}
			if len(c.context.AliasedPath) == 0 {
				panic("replace allowed only if aliased path is non-empty")
			}
		}()
	}

	chainCopy := c.clone()

	chainCopy.state = stateEntered
	if len(chainCopy.context.Path) != 0 {
		last := len(chainCopy.context.Path) - 1
		chainCopy.context.Path[last] = fmt.Sprintf(name, args...)
	}
	if len(chainCopy.context.AliasedPath) != 0 {
		last := len(chainCopy.context.AliasedPath) - 1
		chainCopy.context.AliasedPath[last] = fmt.Sprintf(name, args...)
	}

	return chainCopy
}

// Finalize assertion.
// If there were no failures, report succeeded assertion to AssertionHandler.
// Otherwise, mark parent as failed and notify grandparents that they
// have faield children.
// Must be called after enter().
// Chain can't be used after this call.
func (c *chain) leave() {
	var (
		context       AssertionContext
		handler       AssertionHandler
		parent        *chain
		reportSuccess bool
		reportFailure bool
	)

	func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		if chainValidation && c.state != stateEntered {
			panic("unpaired enter/leave")
		}
		c.state = stateLeaved

		if c.flags&(flagFailed|flagFailedChildren) == 0 {
			context = c.context
			handler = c.handler
			reportSuccess = true
		} else if c.parent != nil {
			parent = c.parent
			reportFailure = true
		}
	}()

	if reportSuccess {
		handler.Success(&context)
	}

	if reportFailure {
		parent.mu.Lock()
		parent.flags |= flagFailed
		p := parent.parent
		parent.mu.Unlock()

		for p != nil {
			p.mu.Lock()
			p.flags |= flagFailedChildren
			pp := p.parent
			p.mu.Unlock()
			p = pp
		}
	}
}

// Initialize and set name to aliased path.
func (c *chain) setAlias(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if chainValidation && c.state == stateLeaved {
		panic("can't use chain after leave")
	}

	if name != "" {
		c.context.AliasedPath = []string{name}
	} else {
		c.context.AliasedPath = []string{}
	}
}

// Report assertion failure and mark chain as failed.
// Must be called between enter() and leave().
func (c *chain) fail(failure AssertionFailure) {
	var (
		context       AssertionContext
		handler       AssertionHandler
		reportFailure bool
	)

	func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		if chainValidation && c.state != stateEntered {
			panic("fail allowed only between enter/leave")
		}

		if c.flags&flagFailed != 0 {
			return
		}
		c.flags |= flagFailed

		failure.Severity = c.severity
		if c.severity == SeverityError {
			failure.IsFatal = true
		}

		context = c.context
		handler = c.handler
		reportFailure = true
	}()

	if reportFailure {
		handler.Failure(&context, &failure)

		if chainValidation {
			if err := validateAssertion(&failure); err != nil {
				panic(err)
			}
		}
	}
}

// Check if chain failed.
func (c *chain) failed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.flags&flagFailed != 0
}

// Check if chain or any of its children failed.
func (c *chain) treeFailed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.flags&(flagFailed|flagFailedChildren) != 0
}

// Set failure flag.
// For tests.
func (c *chain) setFailed() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.flags |= flagFailed
}

// Clear failure flags.
// For tests.
func (c *chain) clearFailed() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.flags &= ^(flagFailed | flagFailedChildren)
}

// Report failure unless chain is not failed.
// For tests.
func (c *chain) assertNotFailed(t testing.TB) {
	c.mu.Lock()
	defer c.mu.Unlock()

	assert.Equal(t, chainFlags(0), c.flags&flagFailed,
		"expected: chain is not failed")
}

// Report failure unless chain is failed.
// For tests.
func (c *chain) assertFailed(t testing.TB) {
	c.mu.Lock()
	defer c.mu.Unlock()

	assert.NotEqual(t, chainFlags(0), c.flags&flagFailed,
		"expected: chain is failed")
}

// Report failure unless chain has specified flags.
// For tests.
func (c *chain) assertFlags(t testing.TB, flags chainFlags) {
	c.mu.Lock()
	defer c.mu.Unlock()

	assert.Equal(t, flags, c.flags,
		"expected: chain has specified flags")
}
