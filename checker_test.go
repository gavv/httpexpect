package httpexpect

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func testCheckerFail(t *testing.T, checker Checker) {
	assert.False(t, checker.Failed())

	checker.Fail("fail")
	assert.True(t, checker.Failed())

	checker.Fail("fail")
	assert.True(t, checker.Failed())
}

func testCheckerClone(t *testing.T, checker Checker) {
	clone := checker.Clone()

	assert.False(t, checker.Failed())
	assert.False(t, clone.Failed())

	checker.Fail("fail")

	assert.True(t, checker.Failed())
	assert.False(t, clone.Failed())

	clone.Fail("fail")

	assert.True(t, checker.Failed())
	assert.True(t, clone.Failed())
}

func testCheckerType(t *testing.T, checker func() Checker) {
	testCheckerFail(t, checker())
	testCheckerClone(t, checker())
}

func TestAssertChecker(t *testing.T) {
	testCheckerType(t, func() Checker {
		return NewAssertChecker(&testing.T{})
	})
}

func TestMockChecker(t *testing.T) {
	testCheckerType(t, func() Checker {
		return newMockChecker(t)
	})
}

func TestRequireChecker(t *testing.T) {
	checker := NewRequireChecker(t)

	clone := checker.Clone()
	assert.False(t, clone.Failed())
}
