package httpexpect

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStringFailed(t *testing.T) {
	checker := newMockChecker(t)

	checker.Fail("fail")

	value := NewString(checker, "")

	value.Empty()
	value.NotEmpty()
	value.Equal("")
	value.NotEqual("")
	value.EqualFold("")
	value.NotEqualFold("")
	value.Contains("")
	value.NotContains("")
	value.ContainsFold("")
	value.NotContainsFold("")
}

func TestStringEmpty(t *testing.T) {
	checker := newMockChecker(t)

	value1 := NewString(checker, "")

	value1.Empty()
	checker.AssertSuccess(t)
	checker.Reset()

	value1.NotEmpty()
	checker.AssertFailed(t)
	checker.Reset()

	value2 := NewString(checker, "a")

	value2.Empty()
	checker.AssertFailed(t)
	checker.Reset()

	value2.NotEmpty()
	checker.AssertSuccess(t)
	checker.Reset()
}

func TestStringEqual(t *testing.T) {
	checker := newMockChecker(t)

	value := NewString(checker, "foo")

	assert.Equal(t, "foo", value.Raw())

	value.Equal("foo")
	checker.AssertSuccess(t)
	checker.Reset()

	value.Equal("FOO")
	checker.AssertFailed(t)
	checker.Reset()

	value.NotEqual("FOO")
	checker.AssertSuccess(t)
	checker.Reset()

	value.NotEqual("foo")
	checker.AssertFailed(t)
	checker.Reset()
}

func TestStringEqualFold(t *testing.T) {
	checker := newMockChecker(t)

	value := NewString(checker, "foo")

	value.EqualFold("foo")
	checker.AssertSuccess(t)
	checker.Reset()

	value.EqualFold("FOO")
	checker.AssertSuccess(t)
	checker.Reset()

	value.EqualFold("foo2")
	checker.AssertFailed(t)
	checker.Reset()

	value.NotEqualFold("foo")
	checker.AssertFailed(t)
	checker.Reset()

	value.NotEqualFold("FOO")
	checker.AssertFailed(t)
	checker.Reset()

	value.NotEqualFold("foo2")
	checker.AssertSuccess(t)
	checker.Reset()
}

func TestStringContains(t *testing.T) {
	checker := newMockChecker(t)

	value := NewString(checker, "11-foo-22")

	value.Contains("foo")
	checker.AssertSuccess(t)
	checker.Reset()

	value.Contains("FOO")
	checker.AssertFailed(t)
	checker.Reset()

	value.NotContains("FOO")
	checker.AssertSuccess(t)
	checker.Reset()

	value.NotContains("foo")
	checker.AssertFailed(t)
	checker.Reset()
}

func TestStringContainsFold(t *testing.T) {
	checker := newMockChecker(t)

	value := NewString(checker, "11-foo-22")

	value.ContainsFold("foo")
	checker.AssertSuccess(t)
	checker.Reset()

	value.ContainsFold("FOO")
	checker.AssertSuccess(t)
	checker.Reset()

	value.ContainsFold("foo3")
	checker.AssertFailed(t)
	checker.Reset()

	value.NotContainsFold("foo")
	checker.AssertFailed(t)
	checker.Reset()

	value.NotContainsFold("FOO")
	checker.AssertFailed(t)
	checker.Reset()

	value.NotContainsFold("foo3")
	checker.AssertSuccess(t)
	checker.Reset()
}
