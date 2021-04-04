package httpexpect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchFailed(t *testing.T) {
	chain := makeChain(newMockContext(t))

	chain.fail("fail")

	value := &Match{chain, nil, nil}

	value.chain.assertFailed(t)

	assert.False(t, value.Length() == nil)
	assert.False(t, value.Index(0) == nil)
	assert.False(t, value.Name("") == nil)

	value.Length().chain.assertFailed(t)
	value.Index(0).chain.assertFailed(t)
	value.Name("").chain.assertFailed(t)

	value.Empty()
	value.NotEmpty()
	value.Values("")
	value.NotValues("")
}

func TestMatchGetters(t *testing.T) {
	ctx := newMockContext(t)

	matches := []string{"m0", "m1", "m2"}
	names := []string{"", "n1", "n2"}

	value := NewMatch(ctx, matches, names)

	assert.Equal(t, matches, value.Raw())

	assert.Equal(t, 3.0, value.Length().Raw())

	assert.Equal(t, "m0", value.Index(0).Raw())
	assert.Equal(t, "m1", value.Index(1).Raw())
	assert.Equal(t, "m2", value.Index(2).Raw())
	value.chain.assertOK(t)

	assert.Equal(t, "m1", value.Name("n1").Raw())
	assert.Equal(t, "m2", value.Name("n2").Raw())
	value.chain.assertOK(t)

	assert.Equal(t, "", value.Index(-1).Raw())
	value.chain.assertFailed(t)
	value.chain.reset()

	assert.Equal(t, "", value.Index(3).Raw())
	value.chain.assertFailed(t)
	value.chain.reset()

	assert.Equal(t, "", value.Name("").Raw())
	value.chain.assertFailed(t)
	value.chain.reset()

	assert.Equal(t, "", value.Name("bad").Raw())
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestMatchEmpty(t *testing.T) {
	ctx := newMockContext(t)

	value1 := NewMatch(ctx, []string{"m"}, nil)
	value2 := NewMatch(ctx, []string{}, nil)
	value3 := NewMatch(ctx, nil, nil)

	assert.Equal(t, []string{}, value2.Raw())
	assert.Equal(t, []string{}, value3.Raw())

	value1.Empty()
	value1.chain.assertFailed(t)
	value1.chain.reset()

	value1.NotEmpty()
	value1.chain.assertOK(t)
	value1.chain.reset()

	value2.Empty()
	value2.chain.assertOK(t)
	value2.chain.reset()

	value2.NotEmpty()
	value2.chain.assertFailed(t)
	value2.chain.reset()

	value3.Empty()
	value3.chain.assertOK(t)
	value3.chain.reset()

	value3.NotEmpty()
	value3.chain.assertFailed(t)
	value3.chain.reset()
}

func TestMatchValues(t *testing.T) {
	ctx := newMockContext(t)

	value := NewMatch(ctx, []string{"m0", "m1", "m2"}, nil)

	value.Values("m1", "m2")
	value.chain.assertOK(t)
	value.chain.reset()

	value.Values("m2", "m1")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Values("m1")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Values()
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotValues("m1", "m2")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotValues("m2", "m1")
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotValues("m1")
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotValues()
	value.chain.assertOK(t)
	value.chain.reset()
}

func TestMatchValuesEmpty(t *testing.T) {
	ctx := newMockContext(t)

	value1 := NewMatch(ctx, nil, nil)
	value2 := NewMatch(ctx, []string{}, nil)
	value3 := NewMatch(ctx, []string{"m0"}, nil)

	value1.Values()
	value1.chain.assertOK(t)
	value1.chain.reset()

	value1.Values("")
	value1.chain.assertFailed(t)
	value1.chain.reset()

	value2.Values()
	value2.chain.assertOK(t)
	value2.chain.reset()

	value2.Values("")
	value2.chain.assertFailed(t)
	value2.chain.reset()

	value3.Values()
	value3.chain.assertOK(t)
	value3.chain.reset()

	value3.Values("m0")
	value3.chain.assertFailed(t)
	value3.chain.reset()
}
