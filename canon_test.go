package httpexpect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanon_Number(t *testing.T) {
	type (
		myInt int
	)

	chain := newMockChain(t).enter("test")
	defer chain.leave()

	var (
		val float64
		ok  bool
	)

	val, ok = canonNumber(chain, 123)
	assert.True(t, ok)
	assert.Equal(t, 123.0, val)
	chain.assertNotFailed(t)
	chain.clearFailed()

	val, ok = canonNumber(chain, 123.0)
	assert.True(t, ok)
	assert.Equal(t, 123.0, val)
	chain.assertNotFailed(t)
	chain.clearFailed()

	val, ok = canonNumber(chain, myInt(123))
	assert.True(t, ok)
	assert.Equal(t, 123.0, val)
	chain.assertNotFailed(t)
	chain.clearFailed()

	_, ok = canonNumber(chain, "123")
	assert.False(t, ok)
	chain.assertFailed(t)
	chain.clearFailed()

	_, ok = canonNumber(chain, nil)
	assert.False(t, ok)
	chain.assertFailed(t)
	chain.clearFailed()
}

func TestCanon_Array(t *testing.T) {
	type (
		myArray []interface{}
		myInt   int
	)

	chain := newMockChain(t).enter("test")
	defer chain.leave()

	var (
		val []interface{}
		ok  bool
	)

	val, ok = canonArray(chain, []interface{}{123.0, 456.0})
	assert.True(t, ok)
	assert.Equal(t, []interface{}{123.0, 456.0}, val)
	chain.assertNotFailed(t)
	chain.clearFailed()

	val, ok = canonArray(chain, myArray{myInt(123), 456.0})
	assert.True(t, ok)
	assert.Equal(t, []interface{}{123.0, 456.0}, val)
	chain.assertNotFailed(t)
	chain.clearFailed()

	_, ok = canonArray(chain, "123")
	assert.False(t, ok)
	chain.assertFailed(t)
	chain.clearFailed()

	_, ok = canonArray(chain, func() {})
	assert.False(t, ok)
	chain.assertFailed(t)
	chain.clearFailed()

	_, ok = canonArray(chain, nil)
	assert.False(t, ok)
	chain.assertFailed(t)
	chain.clearFailed()

	_, ok = canonArray(chain, []interface{}(nil))
	assert.False(t, ok)
	chain.assertFailed(t)
	chain.clearFailed()
}

func TestCanon_Map(t *testing.T) {
	type (
		myMap map[string]interface{}
		myInt int
	)

	chain := newMockChain(t).enter("test")
	defer chain.leave()

	var (
		val map[string]interface{}
		ok  bool
	)

	val, ok = canonMap(chain, map[string]interface{}{"foo": 123.0})
	assert.True(t, ok)
	assert.Equal(t, map[string]interface{}{"foo": 123.0}, val)
	chain.assertNotFailed(t)
	chain.clearFailed()

	val, ok = canonMap(chain, myMap{"foo": myInt(123)})
	assert.True(t, ok)
	assert.Equal(t, map[string]interface{}{"foo": 123.0}, val)
	chain.assertNotFailed(t)
	chain.clearFailed()

	_, ok = canonMap(chain, "123")
	assert.False(t, ok)
	chain.assertFailed(t)
	chain.clearFailed()

	_, ok = canonMap(chain, func() {})
	assert.False(t, ok)
	chain.assertFailed(t)
	chain.clearFailed()

	_, ok = canonMap(chain, nil)
	assert.False(t, ok)
	chain.assertFailed(t)
	chain.clearFailed()

	_, ok = canonMap(chain, map[string]interface{}(nil))
	assert.False(t, ok)
	chain.assertFailed(t)
	chain.clearFailed()
}

func TestCannon_Decode(t *testing.T) {
	t.Run("target is nil", func(t *testing.T) {
		chain := newMockChain(t).enter("test")
		defer chain.leave()

		canonDecode(chain, 123, nil)

		chain.assertFailed(t)
	})

	t.Run("value is not marshallable", func(t *testing.T) {
		chain := newMockChain(t).enter("test")
		defer chain.leave()

		type S struct {
			MyFunc func() string
		}

		value := &S{
			MyFunc: func() string { return "foo" },
		}

		var target S
		canonDecode(chain, value, &target)

		chain.assertFailed(t)
	})

	t.Run("value is not unmarshallable into target", func(t *testing.T) {
		chain := newMockChain(t).enter("test")
		defer chain.leave()

		var target int
		canonDecode(chain, true, target)

		chain.assertFailed(t)
	})
}
