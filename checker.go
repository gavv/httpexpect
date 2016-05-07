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
// futher failures within this copy of checker and its clones are ignored.
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

// Compare implements Checker.Compare.
func (_ *AssertChecker) Compare(a, b interface{}) bool {
	return assert.ObjectsAreEqual(a, b)
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

// Equal implements Checker.Equal.
func (c *AssertChecker) Equal(expected, actual interface{}) {
	if c.failed {
		return
	}
	if !c.Assertions.Equal(expected, actual) {
		c.failed = true
	}
}

// NotEqual implements Checker.NotEqual.
func (c *AssertChecker) NotEqual(expected, actual interface{}) {
	if c.failed {
		return
	}
	if !c.Assertions.NotEqual(expected, actual) {
		c.failed = true
	}
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

// Compare implements Checker.Compare.
func (_ *RequireChecker) Compare(a, b interface{}) bool {
	return assert.ObjectsAreEqual(a, b)
}

// Failed implements Checker.Failed.
func (_ *RequireChecker) Failed() bool {
	return false
}

// Fail implements Checker.Fail.
func (c *RequireChecker) Fail(message string, args ...interface{}) {
	c.Assertions.FailNow(fmt.Sprintf(message, args...))
}

// Equal implements Checker.Equal.
func (c *RequireChecker) Equal(expected, actual interface{}) {
	c.Assertions.Equal(expected, actual)
}

// NotEqual implements Checker.NotEqual.
func (c *RequireChecker) NotEqual(expected, actual interface{}) {
	c.Assertions.NotEqual(expected, actual)
}
