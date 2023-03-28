package httpexpect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoolean_FailedChain(t *testing.T) {
	chain := newFailedChain(t)

	value := newBoolean(chain, false)
	value.chain.assert(t, failure)

	value.Path("$").chain.assert(t, failure)
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
		value.chain.assert(t, success)
	})

	t.Run("config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewBooleanC(Config{
			Reporter: reporter,
		}, true)
		value.IsEqual(true)
		value.chain.assert(t, success)
	})

	t.Run("chain", func(t *testing.T) {
		chain := newMockChain(t)
		value := newBoolean(chain, true)
		assert.NotSame(t, value.chain, &chain)
		assert.Equal(t, value.chain.context.Path, chain.context.Path)
	})
}

func TestBoolean_Decode(t *testing.T) {
	t.Run("target is empty interface", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewBoolean(reporter, true)

		var target interface{}
		value.Decode(&target)

		value.chain.assert(t, success)
		assert.Equal(t, true, target)
	})

	t.Run("target is bool", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewBoolean(reporter, true)

		var target bool
		value.Decode(&target)

		value.chain.assert(t, success)
		assert.Equal(t, true, target)
	})

	t.Run("target is nil", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewBoolean(reporter, true)

		value.Decode(nil)

		value.chain.assert(t, failure)
	})

	t.Run("target is unmarshable", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewBoolean(reporter, true)

		value.Decode(123)

		value.chain.assert(t, failure)
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
	value.chain.assert(t, success)
	value.chain.clear()

	assert.Equal(t, true, value.Path("$").Raw())
	value.chain.assert(t, success)
	value.chain.clear()

	value.Schema(`{"type": "boolean"}`)
	value.chain.assert(t, success)
	value.chain.clear()

	value.Schema(`{"type": "object"}`)
	value.chain.assert(t, failure)
	value.chain.clear()
}

func TestBoolean_True(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewBoolean(reporter, true)

	assert.Equal(t, true, value.Raw())

	value.IsEqual(true)
	value.chain.assert(t, success)
	value.chain.clear()

	value.IsEqual(false)
	value.chain.assert(t, failure)
	value.chain.clear()

	value.NotEqual(false)
	value.chain.assert(t, success)
	value.chain.clear()

	value.NotEqual(true)
	value.chain.assert(t, failure)
	value.chain.clear()

	value.IsTrue()
	value.chain.assert(t, success)
	value.chain.clear()

	value.IsFalse()
	value.chain.assert(t, failure)
	value.chain.clear()

	value.InList(true, true)
	value.chain.assert(t, success)
	value.chain.clear()

	value.NotInList(true, false)
	value.chain.assert(t, failure)
	value.chain.clear()
}

func TestBoolean_False(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewBoolean(reporter, false)

	assert.Equal(t, false, value.Raw())

	value.IsEqual(true)
	value.chain.assert(t, failure)
	value.chain.clear()

	value.IsEqual(false)
	value.chain.assert(t, success)
	value.chain.clear()

	value.NotEqual(false)
	value.chain.assert(t, failure)
	value.chain.clear()

	value.NotEqual(true)
	value.chain.assert(t, success)
	value.chain.clear()

	value.IsTrue()
	value.chain.assert(t, failure)
	value.chain.clear()

	value.IsFalse()
	value.chain.assert(t, success)
	value.chain.clear()

	value.InList(true, true)
	value.chain.assert(t, failure)
	value.chain.clear()

	value.NotInList(true, true)
	value.chain.assert(t, success)
	value.chain.clear()
}

func TestBoolean_Usage(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewBoolean(reporter, true)

	value.InList()
	value.chain.assert(t, failure)
	value.chain.clear()

	value.NotInList()
	value.chain.assert(t, failure)
	value.chain.clear()
}
