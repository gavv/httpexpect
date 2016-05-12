package httpexpect

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValueFailed(t *testing.T) {
	chain := makeChain(mockReporter{t})

	chain.fail("fail")

	value := &Value{chain, nil}

	value.chain.assertFailed(t)

	assert.False(t, value.Object() == nil)
	assert.False(t, value.Array() == nil)
	assert.False(t, value.String() == nil)
	assert.False(t, value.Number() == nil)
	assert.False(t, value.Boolean() == nil)

	value.Object().chain.assertFailed(t)
	value.Array().chain.assertFailed(t)
	value.String().chain.assertFailed(t)
	value.Number().chain.assertFailed(t)
	value.Boolean().chain.assertFailed(t)

	value.Null()
	value.NotNull()
}

func TestValueCastNull(t *testing.T) {
	reporter := mockReporter{t}

	var data interface{}

	NewValue(reporter, data).Object().chain.assertFailed(t)
	NewValue(reporter, data).Array().chain.assertFailed(t)
	NewValue(reporter, data).String().chain.assertFailed(t)
	NewValue(reporter, data).Number().chain.assertFailed(t)
	NewValue(reporter, data).Boolean().chain.assertFailed(t)
	NewValue(reporter, data).NotNull().chain.assertFailed(t)
	NewValue(reporter, data).Null().chain.assertOK(t)
}

func TestValueCastIndirectNull(t *testing.T) {
	reporter := mockReporter{t}

	var data []interface{}

	NewValue(reporter, data).Object().chain.assertFailed(t)
	NewValue(reporter, data).Array().chain.assertFailed(t)
	NewValue(reporter, data).String().chain.assertFailed(t)
	NewValue(reporter, data).Number().chain.assertFailed(t)
	NewValue(reporter, data).Boolean().chain.assertFailed(t)
	NewValue(reporter, data).NotNull().chain.assertFailed(t)
	NewValue(reporter, data).Null().chain.assertOK(t)
}

func TestValueCastBad(t *testing.T) {
	reporter := mockReporter{t}

	data := func() {}

	NewValue(reporter, data).Object().chain.assertFailed(t)
	NewValue(reporter, data).Array().chain.assertFailed(t)
	NewValue(reporter, data).String().chain.assertFailed(t)
	NewValue(reporter, data).Number().chain.assertFailed(t)
	NewValue(reporter, data).Boolean().chain.assertFailed(t)
	NewValue(reporter, data).NotNull().chain.assertFailed(t)
	NewValue(reporter, data).Null().chain.assertFailed(t)
}

func TestValueCastObject(t *testing.T) {
	reporter := mockReporter{t}

	data := map[string]interface{}{}

	NewValue(reporter, data).Object().chain.assertOK(t)
	NewValue(reporter, data).Array().chain.assertFailed(t)
	NewValue(reporter, data).String().chain.assertFailed(t)
	NewValue(reporter, data).Number().chain.assertFailed(t)
	NewValue(reporter, data).Boolean().chain.assertFailed(t)
	NewValue(reporter, data).NotNull().chain.assertOK(t)
	NewValue(reporter, data).Null().chain.assertFailed(t)
}

func TestValueCastArray(t *testing.T) {
	reporter := mockReporter{t}

	data := []interface{}{}

	NewValue(reporter, data).Object().chain.assertFailed(t)
	NewValue(reporter, data).Array().chain.assertOK(t)
	NewValue(reporter, data).String().chain.assertFailed(t)
	NewValue(reporter, data).Number().chain.assertFailed(t)
	NewValue(reporter, data).Boolean().chain.assertFailed(t)
	NewValue(reporter, data).NotNull().chain.assertOK(t)
	NewValue(reporter, data).Null().chain.assertFailed(t)
}

func TestValueCastString(t *testing.T) {
	reporter := mockReporter{t}

	data := ""

	NewValue(reporter, data).Object().chain.assertFailed(t)
	NewValue(reporter, data).Array().chain.assertFailed(t)
	NewValue(reporter, data).String().chain.assertOK(t)
	NewValue(reporter, data).Number().chain.assertFailed(t)
	NewValue(reporter, data).Boolean().chain.assertFailed(t)
	NewValue(reporter, data).NotNull().chain.assertOK(t)
	NewValue(reporter, data).Null().chain.assertFailed(t)
}

func TestValueCastNumber(t *testing.T) {
	reporter := mockReporter{t}

	data := 0.0

	NewValue(reporter, data).Object().chain.assertFailed(t)
	NewValue(reporter, data).Array().chain.assertFailed(t)
	NewValue(reporter, data).String().chain.assertFailed(t)
	NewValue(reporter, data).Number().chain.assertOK(t)
	NewValue(reporter, data).Boolean().chain.assertFailed(t)
	NewValue(reporter, data).NotNull().chain.assertOK(t)
	NewValue(reporter, data).Null().chain.assertFailed(t)
}

func TestValueCastBoolean(t *testing.T) {
	reporter := mockReporter{t}

	data := false

	NewValue(reporter, data).Object().chain.assertFailed(t)
	NewValue(reporter, data).Array().chain.assertFailed(t)
	NewValue(reporter, data).String().chain.assertFailed(t)
	NewValue(reporter, data).Number().chain.assertFailed(t)
	NewValue(reporter, data).Boolean().chain.assertOK(t)
	NewValue(reporter, data).NotNull().chain.assertOK(t)
	NewValue(reporter, data).Null().chain.assertFailed(t)
}

func TestValueGetObject(t *testing.T) {
	type (
		myMap map[string]interface{}
	)

	reporter := mockReporter{t}

	data1 := map[string]interface{}{"foo": 123.0}

	value1 := NewValue(reporter, data1)
	inner1 := value1.Object()

	inner1.chain.assertOK(t)
	inner1.chain.reset()
	assert.Equal(t, data1, inner1.Raw())

	data2 := myMap{"foo": 123.0}

	value2 := NewValue(reporter, data2)
	inner2 := value2.Object()

	inner2.chain.assertOK(t)
	inner2.chain.reset()
	assert.Equal(t, map[string]interface{}(data2), inner2.Raw())
}

func TestValueGetArray(t *testing.T) {
	type (
		myArray []interface{}
	)

	reporter := mockReporter{t}

	data1 := []interface{}{"foo", 123.0}

	value1 := NewValue(reporter, data1)
	inner1 := value1.Array()

	inner1.chain.assertOK(t)
	inner1.chain.reset()
	assert.Equal(t, data1, inner1.Raw())

	data2 := myArray{"foo", 123.0}

	value2 := NewValue(reporter, data2)
	inner2 := value2.Array()

	inner2.chain.assertOK(t)
	inner2.chain.reset()
	assert.Equal(t, []interface{}(data2), inner2.Raw())
}

func TestValueGetString(t *testing.T) {
	reporter := mockReporter{t}

	value := NewValue(reporter, "foo")
	inner := value.String()

	inner.chain.assertOK(t)
	inner.chain.reset()
	assert.Equal(t, "foo", inner.Raw())
}

func TestValueGetNumber(t *testing.T) {
	type (
		myInt int
	)

	reporter := mockReporter{t}

	data1 := 123.0

	value1 := NewValue(reporter, data1)
	inner1 := value1.Number()

	inner1.chain.assertOK(t)
	inner1.chain.reset()
	assert.Equal(t, data1, inner1.Raw())

	data2 := 123

	value2 := NewValue(reporter, data2)
	inner2 := value2.Number()

	inner2.chain.assertOK(t)
	inner2.chain.reset()
	assert.Equal(t, float64(data2), inner2.Raw())

	data3 := myInt(123)

	value3 := NewValue(reporter, data3)
	inner3 := value3.Number()

	inner3.chain.assertOK(t)
	inner3.chain.reset()
	assert.Equal(t, float64(data3), inner3.Raw())
}

func TestValueGetBoolean(t *testing.T) {
	reporter := mockReporter{t}

	value1 := NewValue(reporter, true)
	inner1 := value1.Boolean()

	inner1.chain.assertOK(t)
	inner1.chain.reset()
	assert.Equal(t, true, inner1.Raw())

	value2 := NewValue(reporter, false)
	inner2 := value2.Boolean()

	inner2.chain.assertOK(t)
	inner2.chain.reset()
	assert.Equal(t, false, inner2.Raw())
}
