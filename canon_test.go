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

	d1, ok := canonNumber(chain, 123)
	assert.True(t, ok)
	assert.Equal(t, 123.0, d1)
	chain.assertNotFailed(t)
	chain.clearFailed()

	d2, ok := canonNumber(chain, 123.0)
	assert.True(t, ok)
	assert.Equal(t, 123.0, d2)
	chain.assertNotFailed(t)
	chain.clearFailed()

	d3, ok := canonNumber(chain, myInt(123))
	assert.True(t, ok)
	assert.Equal(t, 123.0, d3)
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

	d1, ok := canonArray(chain, []interface{}{123.0, 456.0})
	assert.True(t, ok)
	assert.Equal(t, []interface{}{123.0, 456.0}, d1)
	chain.assertNotFailed(t)
	chain.clearFailed()

	d2, ok := canonArray(chain, myArray{myInt(123), 456.0})
	assert.True(t, ok)
	assert.Equal(t, []interface{}{123.0, 456.0}, d2)
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

	d1, ok := canonMap(chain, map[string]interface{}{"foo": 123.0})
	assert.True(t, ok)
	assert.Equal(t, map[string]interface{}{"foo": 123.0}, d1)
	chain.assertNotFailed(t)
	chain.clearFailed()

	d2, ok := canonMap(chain, myMap{"foo": myInt(123)})
	assert.True(t, ok)
	assert.Equal(t, map[string]interface{}{"foo": 123.0}, d2)
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
	t.Run("target is nill", func(t *testing.T) {
		chain := newMockChain(t)

		canonDecode(chain, 123, nil)

		chain.assertFailed(t)
	})

	t.Run("value is not marshallable", func(t *testing.T) {
		chain := newMockChain(t)

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
		chain := newMockChain(t)

		var target int
		canonDecode(chain, true, target)

		chain.assertFailed(t)
	})
}
