package httpexpect

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type AssertChecker struct {
	*assert.Assertions
	failed bool
}

func NewAssertChecker(t *testing.T) *AssertChecker {
	return &AssertChecker{assert.New(t), false}
}

func (c *AssertChecker) Clone() Checker {
	copy := *c
	return &copy
}

func (_ *AssertChecker) Compare(a, b interface{}) bool {
	return assert.ObjectsAreEqual(a, b)
}

func (c *AssertChecker) Failed() bool {
	return c.failed
}

func (c *AssertChecker) Fail(message string, args... interface{}) {
	if c.failed {
		return
	}
	c.Assertions.Fail(message, args...)
	c.failed = true
}

func (c *AssertChecker) Equal(expected, actual interface{}) {
	if c.failed {
		return
	}
	if !c.Assertions.Equal(expected, actual) {
		c.failed = true
	}
}

func (c *AssertChecker) NotEqual(expected, actual interface{}) {
	if c.failed {
		return
	}
	if !c.Assertions.NotEqual(expected, actual) {
		c.failed = true
	}
}

type RequireChecker struct {
	*require.Assertions
}

func NewRequireChecker(t *testing.T) *RequireChecker {
	return &RequireChecker{require.New(t)}
}

func (c *RequireChecker) Clone() Checker {
	return c
}

func (_ *RequireChecker) Compare(a, b interface{}) bool {
	return assert.ObjectsAreEqual(a, b)
}

func (_ *RequireChecker) Failed() bool {
	return false
}

func (c *RequireChecker) Fail(message string, args... interface{}) {
	c.Assertions.FailNow(message, args...)
}

func (c *RequireChecker) Equal(expected, actual interface{}) {
	c.Assertions.Equal(expected, actual)
}

func (c *RequireChecker) NotEqual(expected, actual interface{}) {
	c.Assertions.NotEqual(expected, actual)
}

type mockChecker struct {
	testing *testing.T
	failed bool
}

func newMockChecker(t *testing.T) *mockChecker {
	return &mockChecker{testing: t}
}

func (c *mockChecker) AssertSuccess(t *testing.T) {
	assert.False(t, c.failed)
}

func (c *mockChecker) AssertFailed(t *testing.T) {
	assert.True(t, c.failed)
}

func (c *mockChecker) Reset() {
	c.failed = false
}

func (c *mockChecker) Clone() Checker {
	copy := *c
	return &copy
}

func (_ *mockChecker) Compare(a, b interface{}) bool {
	return assert.ObjectsAreEqual(a, b)
}

func (c *mockChecker) Failed() bool {
	return c.failed
}

func (c *mockChecker) Fail(message string, args... interface{}) {
	c.testing.Logf("Fail: " + message, args...)
	c.failed = true
}

func (c *mockChecker) Equal(expected, actual interface{}) {
	if !c.Compare(expected, actual) {
		c.testing.Logf("Equal: `%v` (expected) != `%v` (actual)", expected, actual)
		c.failed = true
	}
}

func (c *mockChecker) NotEqual(expected, actual interface{}) {
	if c.Compare(expected, actual) {
		c.testing.Logf("NotEqual: `%v` (expected) == `%v` (actual)", expected, actual)
		c.failed = true
	}
}
