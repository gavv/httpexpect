package httpexpect

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// AssertChecker implements Checker interface using `testify/assert' package.
//
// All failures are non-fatal with this checker. After first failure, all
// further failures within this copy of checker and its clones are ignored.
type AssertChecker struct {
	*assert.Assertions
	failed bool
}

// NewAssertChecker returns a new AssertChecker object.
func NewAssertChecker(t *testing.T) *AssertChecker {
	return &AssertChecker{assert.New(t), false}
}

// Clone implements Checker.Clone.
func (c *AssertChecker) Clone() Checker {
	copy := *c
	return &copy
}

// Failed implements Checker.Failed.
func (c *AssertChecker) Failed() bool {
	return c.failed
}

// Fail implements Checker.Fail.
func (c *AssertChecker) Fail(message string, args ...interface{}) {
	if c.failed {
		return
	}
	c.Assertions.Fail(fmt.Sprintf(message, args...))
	c.failed = true
}

// RequireChecker implements Checker interface using `testify/require' package.
//
// All failures fatal with this checker. After first failure, Goexit() is called
// and test is terminated.
type RequireChecker struct {
	*require.Assertions
}

// NewRequireChecker returns a new RequireChecker object.
func NewRequireChecker(t *testing.T) *RequireChecker {
	return &RequireChecker{require.New(t)}
}

// Clone implements Checker.Clone.
func (c *RequireChecker) Clone() Checker {
	return c
}

// Failed implements Checker.Failed.
func (c *RequireChecker) Failed() bool {
	return false
}

// Fail implements Checker.Fail.
func (c *RequireChecker) Fail(message string, args ...interface{}) {
	c.Assertions.FailNow(fmt.Sprintf(message, args...))
}
