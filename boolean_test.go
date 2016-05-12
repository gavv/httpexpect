package httpexpect

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBooleanFailed(t *testing.T) {
	chain := makeChain(mockReporter{t})

	chain.fail("fail")

	value := &Boolean{chain, false}

	value.chain.assertFailed(t)

	value.Equal(false)
	value.NotEqual(false)
	value.True()
	value.False()
}

func TestBooleanTrue(t *testing.T) {
	reporter := mockReporter{t}

	value := NewBoolean(reporter, true)

	assert.Equal(t, true, value.Raw())

	value.Equal(true)
	value.chain.assertOK(t)
	value.chain.reset()

	value.Equal(false)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqual(false)
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotEqual(true)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.True()
	value.chain.assertOK(t)
	value.chain.reset()

	value.False()
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestBooleanFalse(t *testing.T) {
	reporter := mockReporter{t}

	value := NewBoolean(reporter, false)

	assert.Equal(t, false, value.Raw())

	value.Equal(true)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Equal(false)
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotEqual(false)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqual(true)
	value.chain.assertOK(t)
	value.chain.reset()

	value.True()
	value.chain.assertFailed(t)
	value.chain.reset()

	value.False()
	value.chain.assertOK(t)
	value.chain.reset()
}
