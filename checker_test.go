package httpexpect

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func testCheckerCompare(t *testing.T, checker Checker) {
	assert.True(t, checker.Compare(123, 123))
	assert.False(t, checker.Compare(123, 456))
	assert.False(t, checker.Failed())
}

func testCheckerFail(t *testing.T, checker Checker) {
	assert.False(t, checker.Failed())
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
}

func testCheckerEqual(t *testing.T, checker Checker) {
	checker.Equal(123, 123)
	assert.False(t, checker.Failed())

	checker.Equal(123, 456)
	assert.True(t, checker.Failed())

	checker.Equal(123, 123)
	assert.True(t, checker.Failed())
}

func testCheckerNotEqual(t *testing.T, checker Checker) {
	checker.NotEqual(123, 456)
	assert.False(t, checker.Failed())

	checker.NotEqual(123, 123)
	assert.True(t, checker.Failed())

	checker.NotEqual(123, 456)
	assert.True(t, checker.Failed())
}

func testCheckerType(t *testing.T, checker func() Checker) {
	testCheckerCompare(t, checker())
	testCheckerFail(t, checker())
	testCheckerClone(t, checker())
	testCheckerEqual(t, checker())
	testCheckerNotEqual(t, checker())
}

func TestAssertChecker(t *testing.T) {
	testCheckerType(t, func() Checker {
		return NewAssertChecker(&testing.T{})
	})
}

func TestMockChecker(t *testing.T) {
	testCheckerType(t, func() Checker {
		return &mockChecker{}
	})
}

func TestRequireChecker(t *testing.T) {
	checker := NewRequireChecker(t)

	assert.True(t, checker.Compare(123, 123))
	assert.False(t, checker.Compare(123, 456))
	assert.False(t, checker.Failed())

	checker.Equal(123, 123)
	assert.False(t, checker.Failed())

	checker.NotEqual(123, 456)
	assert.False(t, checker.Failed())

	clone := checker.Clone()
	assert.False(t, clone.Failed())
}
