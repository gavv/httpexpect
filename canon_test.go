package httpexpect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanon_Number(t *testing.T) {
	type (
		myInt int
	)

	chain := newMockChain(t)

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

	chain := newMockChain(t)

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

	chain := newMockChain(t)

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
