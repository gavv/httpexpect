package httpexpect

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestMatch_Failed(t *testing.T) {
	chain := newMockChain(t)
	chain.fail(mockFailure())

	value := newMatch(chain, nil, nil)

	assert.NotNil(t, value.Length())
	assert.NotNil(t, value.Index(0))
	assert.NotNil(t, value.Name(""))

	value.Empty()
	value.NotEmpty()
	value.Values("")
	value.NotValues("")
}

func TestMatch_Constructors(t *testing.T) {
	matches := []string{"m0", "m1", "m2"}
	names := []string{"", "n1", "n2"}

	t.Run("Constructor without config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewMatch(reporter, matches, names)
		assert.Equal(t, matches, value.Raw())
		value.chain.assertNotFailed(t)
	})

	t.Run("Constructor with config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewMatchC(Config{
			Reporter: reporter,
		}, matches, names)
		assert.Equal(t, matches, value.Raw())
		value.chain.assertNotFailed(t)
	})

	t.Run("chain Constructor", func(t *testing.T) {
		chain := newMockChain(t)
		value := newMatch(chain, matches, names)
		assert.NotEqual(t, unsafe.Pointer(&(value.chain)), unsafe.Pointer(&chain))
		assert.Equal(t, value.chain.context.Path, chain.context.Path)
	})
}

func TestMatch_Getters(t *testing.T) {
	reporter := newMockReporter(t)

	matches := []string{"m0", "m1", "m2"}
	names := []string{"", "n1", "n2"}

	value := NewMatch(reporter, matches, names)

	assert.Equal(t, matches, value.Raw())

	assert.Equal(t, 3.0, value.Length().Raw())

	assert.Equal(t, "m0", value.Index(0).Raw())
	assert.Equal(t, "m1", value.Index(1).Raw())
	assert.Equal(t, "m2", value.Index(2).Raw())
	value.chain.assertNotFailed(t)

	assert.Equal(t, "m1", value.Name("n1").Raw())
	assert.Equal(t, "m2", value.Name("n2").Raw())
	value.chain.assertNotFailed(t)

	assert.Equal(t, "", value.Index(-1).Raw())
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	assert.Equal(t, "", value.Index(3).Raw())
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	assert.Equal(t, "", value.Name("").Raw())
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	assert.Equal(t, "", value.Name("bad").Raw())
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestMatch_Empty(t *testing.T) {
	reporter := newMockReporter(t)

	value1 := NewMatch(reporter, []string{"m"}, nil)
	value2 := NewMatch(reporter, []string{}, nil)
	value3 := NewMatch(reporter, nil, nil)

	assert.Equal(t, []string{}, value2.Raw())
	assert.Equal(t, []string{}, value3.Raw())

	value1.Empty()
	value1.chain.assertFailed(t)
	value1.chain.clearFailed()

	value1.NotEmpty()
	value1.chain.assertNotFailed(t)
	value1.chain.clearFailed()

	value2.Empty()
	value2.chain.assertNotFailed(t)
	value2.chain.clearFailed()

	value2.NotEmpty()
	value2.chain.assertFailed(t)
	value2.chain.clearFailed()

	value3.Empty()
	value3.chain.assertNotFailed(t)
	value3.chain.clearFailed()

	value3.NotEmpty()
	value3.chain.assertFailed(t)
	value3.chain.clearFailed()
}

func TestMatch_Values(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewMatch(reporter, []string{"m0", "m1", "m2"}, nil)

	value.Values("m1", "m2")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Values("m2", "m1")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Values("m1")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Values()
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
}

func TestMatch_ValuesEmpty(t *testing.T) {
	reporter := newMockReporter(t)

	value1 := NewMatch(reporter, nil, nil)
	value2 := NewMatch(reporter, []string{}, nil)
	value3 := NewMatch(reporter, []string{"m0"}, nil)

	value1.Values()
	value1.chain.assertNotFailed(t)
	value1.chain.clearFailed()

	value1.Values("")
	value1.chain.assertFailed(t)
	value1.chain.clearFailed()

	value2.Values()
	value2.chain.assertNotFailed(t)
	value2.chain.clearFailed()

	value2.Values("")
	value2.chain.assertFailed(t)
	value2.chain.clearFailed()

	value3.Values()
	value3.chain.assertNotFailed(t)
	value3.chain.clearFailed()

	value3.Values("m0")
	value3.chain.assertFailed(t)
	value3.chain.clearFailed()
}
