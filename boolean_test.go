package httpexpect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoolean_Failed(t *testing.T) {
	chain := newMockChain(t)
	chain.setFailed()

	value := newBoolean(chain, false)

	value.Path("$")
	value.Schema("")

	var target interface{}
	value.Decode(&target)

	value.Equal(false)
	value.NotEqual(false)
	value.True()
	value.False()
}

func TestBoolean_Constructors(t *testing.T) {
	t.Run("Constructor without config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewBoolean(reporter, true)
		value.Equal(true)
		value.chain.assertNotFailed(t)
	})

	t.Run("Constructor with config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewBooleanC(Config{
			Reporter: reporter,
		}, true)
		value.Equal(true)
		value.chain.assertNotFailed(t)
	})

	t.Run("chain Constructor", func(t *testing.T) {
		chain := newMockChain(t)
		value := newBoolean(chain, true)
		assert.NotSame(t, value.chain, &chain)
		assert.Equal(t, value.chain.context.Path, chain.context.Path)
	})
}

func TestBoolean_Decode(t *testing.T) {
	t.Run("Decode into empty interface", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewBoolean(reporter, true)

		var target interface{}
		value.Decode(&target)

		value.chain.assertNotFailed(t)
		assert.Equal(t, true, target)
	})

	t.Run("Decode into boolean", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewBoolean(reporter, true)

		var target bool
		value.Decode(&target)

		value.chain.assertNotFailed(t)
		assert.Equal(t, true, target)
	})

	t.Run("Target is unmarshable", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewBoolean(reporter, true)

		value.Decode(123)

		value.chain.assertFailed(t)
	})

	t.Run("Target is nil", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewBoolean(reporter, true)

		value.Decode(nil)

		value.chain.assertFailed(t)
	})
}

func TestBoolean_Getters(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewBoolean(reporter, true)

	assert.Equal(t, true, value.Raw())
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	assert.Equal(t, true, value.Path("$").Raw())
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Schema(`{"type": "boolean"}`)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Schema(`{"type": "object"}`)
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestBoolean_True(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewBoolean(reporter, true)

	assert.Equal(t, true, value.Raw())

	value.Equal(true)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Equal(false)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual(false)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotEqual(true)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.True()
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.False()
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestBoolean_False(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewBoolean(reporter, false)

	assert.Equal(t, false, value.Raw())

	value.Equal(true)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Equal(false)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotEqual(false)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual(true)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.True()
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.False()
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}
