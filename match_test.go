package httpexpect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatch_FailedChain(t *testing.T) {
	chain := newMockChain(t)
	chain.setFailed()

	value := newMatch(chain, nil, nil)
	value.chain.assertFailed(t)

	value.Alias("foo")

	value.Length().chain.assertFailed(t)
	value.Value(0).chain.assertFailed(t)
	value.NamedValue("").chain.assertFailed(t)

	value.IsEmpty()
	value.NotEmpty()
	value.HasValues("")
	value.NotValues("")
}

func TestMatch_Constructors(t *testing.T) {
	matches := []string{"m0", "m1", "m2"}
	names := []string{"", "n1", "n2"}

	t.Run("reporter", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewMatch(reporter, matches, names)
		assert.Equal(t, matches, value.Raw())
		value.chain.assertNotFailed(t)
	})

	t.Run("config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewMatchC(Config{
			Reporter: reporter,
		}, matches, names)
		assert.Equal(t, matches, value.Raw())
		value.chain.assertNotFailed(t)
	})

	t.Run("chain", func(t *testing.T) {
		chain := newMockChain(t)
		value := newMatch(chain, matches, names)
		assert.NotSame(t, value.chain, chain)
		assert.Equal(t, value.chain.context.Path, chain.context.Path)
	})
}

func TestMatch_Alias(t *testing.T) {
	reporter := newMockReporter(t)

	matches := []string{"m0", "m1", "m2"}
	names := []string{"", "n1", "n2"}

	value := NewMatch(reporter, matches, names)
	assert.Equal(t, []string{"Match()"}, value.chain.context.Path)
	assert.Equal(t, []string{"Match()"}, value.chain.context.AliasedPath)

	value.Alias("foo")
	assert.Equal(t, []string{"Match()"}, value.chain.context.Path)
	assert.Equal(t, []string{"foo"}, value.chain.context.AliasedPath)

	childValue := value.Value(0)
	assert.Equal(t, []string{"Match()", "Index(0)"}, childValue.chain.context.Path)
	assert.Equal(t, []string{"foo", "Index(0)"}, childValue.chain.context.AliasedPath)
}

func TestMatch_Getters(t *testing.T) {
	reporter := newMockReporter(t)

	matches := []string{"m0", "m1", "m2"}
	names := []string{"", "n1", "n2"}

	value := NewMatch(reporter, matches, names)

	assert.Equal(t, matches, value.Raw())

	assert.Equal(t, 3.0, value.Length().Raw())

	assert.Equal(t, "m0", value.Value(0).Raw())
	assert.Equal(t, "m1", value.Value(1).Raw())
	assert.Equal(t, "m2", value.Value(2).Raw())
	value.chain.assertNotFailed(t)

	assert.Equal(t, "m1", value.NamedValue("n1").Raw())
	assert.Equal(t, "m2", value.NamedValue("n2").Raw())
	value.chain.assertNotFailed(t)

	assert.Equal(t, "", value.Value(-1).Raw())
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	assert.Equal(t, "", value.Value(3).Raw())
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	assert.Equal(t, "", value.NamedValue("").Raw())
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	assert.Equal(t, "", value.NamedValue("bad").Raw())
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestMatch_IsEmpty(t *testing.T) {
	reporter := newMockReporter(t)

	value1 := NewMatch(reporter, []string{"m"}, nil)
	value2 := NewMatch(reporter, []string{}, nil)
	value3 := NewMatch(reporter, nil, nil)

	assert.Equal(t, []string{}, value2.Raw())
	assert.Equal(t, []string{}, value3.Raw())

	value1.IsEmpty()
	value1.chain.assertFailed(t)
	value1.chain.clearFailed()

	value1.NotEmpty()
	value1.chain.assertNotFailed(t)
	value1.chain.clearFailed()

	value2.IsEmpty()
	value2.chain.assertNotFailed(t)
	value2.chain.clearFailed()

	value2.NotEmpty()
	value2.chain.assertFailed(t)
	value2.chain.clearFailed()

	value3.IsEmpty()
	value3.chain.assertNotFailed(t)
	value3.chain.clearFailed()

	value3.NotEmpty()
	value3.chain.assertFailed(t)
	value3.chain.clearFailed()
}

func TestMatch_Values(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		reporter := newMockReporter(t)

		value1 := NewMatch(reporter, nil, nil)
		value2 := NewMatch(reporter, []string{}, nil)
		value3 := NewMatch(reporter, []string{"m0"}, nil)

		value1.HasValues()
		value1.chain.assertNotFailed(t)
		value1.chain.clearFailed()

		value1.HasValues("")
		value1.chain.assertFailed(t)
		value1.chain.clearFailed()

		value2.HasValues()
		value2.chain.assertNotFailed(t)
		value2.chain.clearFailed()

		value2.HasValues("")
		value2.chain.assertFailed(t)
		value2.chain.clearFailed()

		value3.HasValues()
		value3.chain.assertNotFailed(t)
		value3.chain.clearFailed()

		value3.HasValues("m0")
		value3.chain.assertFailed(t)
		value3.chain.clearFailed()
	})

	t.Run("not empty", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewMatch(reporter, []string{"m0", "m1", "m2"}, nil)

		value.HasValues("m1", "m2")
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.HasValues("m2", "m1")
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.HasValues("m1")
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.HasValues()
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotValues("m1", "m2")
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotValues("m2", "m1")
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotValues("m1")
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotValues()
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()
	})
}
