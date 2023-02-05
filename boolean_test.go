package httpexpect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoolean_FailedChain(t *testing.T) {
	chain := newMockChain(t)
	chain.setFailed()

	value := newBoolean(chain, false)
	value.chain.assertFailed(t)

	value.Path("$").chain.assertFailed(t)
	value.Schema("")
	value.Alias("foo")

	var target interface{}
	value.Decode(&target)

	value.IsTrue()
	value.IsFalse()
	value.IsEqual(false)
	value.NotEqual(false)
	value.InList(false)
	value.NotInList(false)
}

func TestBoolean_Constructors(t *testing.T) {
	t.Run("reporter", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewBoolean(reporter, true)
		value.IsEqual(true)
		value.chain.assertNotFailed(t)
	})

	t.Run("config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewBooleanC(Config{
			Reporter: reporter,
		}, true)
		value.IsEqual(true)
		value.chain.assertNotFailed(t)
	})

	t.Run("chain", func(t *testing.T) {
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

func TestBoolean_Alias(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewBoolean(reporter, true)
	assert.Equal(t, []string{"Boolean()"}, value.chain.context.Path)
	assert.Equal(t, []string{"Boolean()"}, value.chain.context.AliasedPath)

	value.Alias("foo")
	assert.Equal(t, []string{"Boolean()"}, value.chain.context.Path)
	assert.Equal(t, []string{"foo"}, value.chain.context.AliasedPath)
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

	value.IsEqual(true)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.IsEqual(false)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual(false)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotEqual(true)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.IsTrue()
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.IsFalse()
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InList(true, true)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotInList(true, false)
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestBoolean_False(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewBoolean(reporter, false)

	assert.Equal(t, false, value.Raw())

	value.IsEqual(true)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.IsEqual(false)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotEqual(false)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual(true)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.IsTrue()
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.IsFalse()
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.InList(true, true)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInList(true, true)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}

func TestBoolean_UsageChecks(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewBoolean(reporter, true)

	value.InList()
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInList()
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}
