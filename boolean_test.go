package httpexpect

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBooleanTrue(t *testing.T) {
	checker := newMockChecker(t)

	value := NewBoolean(checker, true)

	assert.Equal(t, true, value.Raw())

	value.Equal(true)
	checker.AssertSuccess(t)
	checker.Reset()

	value.Equal(false)
	checker.AssertFailed(t)
	checker.Reset()

	value.NotEqual(false)
	checker.AssertSuccess(t)
	checker.Reset()

	value.NotEqual(true)
	checker.AssertFailed(t)
	checker.Reset()

	value.True()
	checker.AssertSuccess(t)
	checker.Reset()

	value.False()
	checker.AssertFailed(t)
	checker.Reset()
}

func TestBooleanFalse(t *testing.T) {
	checker := newMockChecker(t)

	value := NewBoolean(checker, false)

	assert.Equal(t, false, value.Raw())

	value.Equal(true)
	checker.AssertFailed(t)
	checker.Reset()

	value.Equal(false)
	checker.AssertSuccess(t)
	checker.Reset()

	value.NotEqual(false)
	checker.AssertFailed(t)
	checker.Reset()

	value.NotEqual(true)
	checker.AssertSuccess(t)
	checker.Reset()

	value.True()
	checker.AssertFailed(t)
	checker.Reset()

	value.False()
	checker.AssertSuccess(t)
	checker.Reset()
}
