package httpexpect

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoolean_FailedChain(t *testing.T) {
	chain := newMockChain(t, flagFailed)

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

func TestBoolean_IsEqual(t *testing.T) {
	for _, value := range []bool{true, false} {
		t.Run(fmt.Sprintf("%v", value), func(t *testing.T) {
			reporter := newMockReporter(t)

			NewBoolean(reporter, value).IsEqual(value).
				chain.assert(t, success)

			NewBoolean(reporter, value).IsEqual(!value).
				chain.assert(t, failure)

			NewBoolean(reporter, value).NotEqual(value).
				chain.assert(t, failure)

			NewBoolean(reporter, value).NotEqual(!value).
				chain.assert(t, success)
		})
	}
}

func TestBoolean_IsValue(t *testing.T) {
	for _, value := range []bool{true, false} {
		t.Run(fmt.Sprintf("%v", value), func(t *testing.T) {
			reporter := newMockReporter(t)

			if value {
				NewBoolean(reporter, value).IsTrue().
					chain.assert(t, success)

				NewBoolean(reporter, value).IsFalse().
					chain.assert(t, failure)
			} else {
				NewBoolean(reporter, value).IsTrue().
					chain.assert(t, failure)

				NewBoolean(reporter, value).IsFalse().
					chain.assert(t, success)
			}
		})
	}
}

func TestBoolean_InList(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		for _, value := range []bool{true, false} {
			t.Run(fmt.Sprintf("%v", value), func(t *testing.T) {
				reporter := newMockReporter(t)

				NewBoolean(reporter, value).InList(value).
					chain.assert(t, success)

				NewBoolean(reporter, value).InList(!value, value).
					chain.assert(t, success)

				NewBoolean(reporter, value).InList(!value, !value).
					chain.assert(t, failure)

				NewBoolean(reporter, value).NotInList(value).
					chain.assert(t, failure)

				NewBoolean(reporter, value).NotInList(!value, value).
					chain.assert(t, failure)

				NewBoolean(reporter, value).NotInList(!value, !value).
					chain.assert(t, success)
			})
		}
	})

	t.Run("invalid argument", func(t *testing.T) {
		for _, value := range []bool{true, false} {
			reporter := newMockReporter(t)

			NewBoolean(reporter, value).InList().
				chain.assert(t, failure)

			NewBoolean(reporter, value).NotInList().
				chain.assert(t, failure)
		}
	})
}
