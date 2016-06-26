package httpexpect

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStringFailed(t *testing.T) {
	chain := makeChain(newMockReporter(t))

	chain.fail("fail")

	value := &String{chain, ""}

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
	reporter := newMockReporter(t)

	value1 := NewString(reporter, "")

	value1.Empty()
	value1.chain.assertOK(t)
	value1.chain.reset()

	value1.NotEmpty()
	value1.chain.assertFailed(t)
	value1.chain.reset()

	value2 := NewString(reporter, "a")

	value2.Empty()
	value2.chain.assertFailed(t)
	value2.chain.reset()

	value2.NotEmpty()
	value2.chain.assertOK(t)
	value2.chain.reset()
}

func TestStringEqual(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "foo")

	assert.Equal(t, "foo", value.Raw())

	value.Equal("foo")
	value.chain.assertOK(t)
	value.chain.reset()

	value.Equal("FOO")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqual("FOO")
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotEqual("foo")
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestStringEqualFold(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "foo")

	value.EqualFold("foo")
	value.chain.assertOK(t)
	value.chain.reset()

	value.EqualFold("FOO")
	value.chain.assertOK(t)
	value.chain.reset()

	value.EqualFold("foo2")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqualFold("foo")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqualFold("FOO")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqualFold("foo2")
	value.chain.assertOK(t)
	value.chain.reset()
}

func TestStringContains(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "11-foo-22")

	value.Contains("foo")
	value.chain.assertOK(t)
	value.chain.reset()

	value.Contains("FOO")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotContains("FOO")
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotContains("foo")
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestStringContainsFold(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "11-foo-22")

	value.ContainsFold("foo")
	value.chain.assertOK(t)
	value.chain.reset()

	value.ContainsFold("FOO")
	value.chain.assertOK(t)
	value.chain.reset()

	value.ContainsFold("foo3")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotContainsFold("foo")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotContainsFold("FOO")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotContainsFold("foo3")
	value.chain.assertOK(t)
	value.chain.reset()
}

func TestStringLength(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "1234567")

	num := value.Length()
	value.chain.assertOK(t)
	num.chain.assertOK(t)
	assert.Equal(t, 7.0, num.Raw())
}
