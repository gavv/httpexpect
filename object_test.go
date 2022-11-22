package httpexpect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectFailed(t *testing.T) {
	check := func(value *Object) {
		value.chain.assertFailed(t)

		value.Path("$")
		value.Schema("")

		assert.NotNil(t, value.Keys())
		assert.NotNil(t, value.Values())
		assert.NotNil(t, value.Value("foo"))

		value.Empty()
		value.NotEmpty()
		value.Equal(nil)
		value.NotEqual(nil)
		value.ContainsKey("foo")
		value.NotContainsKey("foo")
		value.ContainsValue("foo")
		value.NotContainsValue("foo")
		value.ContainsSubset(nil)
		value.NotContainsSubset(nil)
		value.ValueEqual("foo", nil)
		value.NotValueEqual("foo", nil)
	}

	t.Run("failed_chain", func(t *testing.T) {
		chain := newMockChain(t)
		chain.fail(AssertionFailure{})

		value := newObject(chain, map[string]interface{}{})

		check(value)
	})

	t.Run("nil_value", func(t *testing.T) {
		chain := newMockChain(t)

		value := newObject(chain, nil)

		check(value)
	})

	t.Run("failed_chain_nil_value", func(t *testing.T) {
		chain := newMockChain(t)
		chain.fail(AssertionFailure{})

		value := newObject(chain, nil)

		check(value)
	})
}

func TestObjectGetters(t *testing.T) {
	reporter := newMockReporter(t)

	m := map[string]interface{}{
		"foo": 123.0,
		"bar": []interface{}{"456", 789.0},
		"baz": map[string]interface{}{
			"a": "b",
		},
	}

	value := NewObject(reporter, m)

	keys := []interface{}{"foo", "bar", "baz"}

	values := []interface{}{
		123.0,
		[]interface{}{"456", 789.0},
		map[string]interface{}{
			"a": "b",
		},
	}

	assert.Equal(t, m, value.Raw())
	value.chain.assertOK(t)
	value.chain.reset()

	assert.Equal(t, m, value.Path("$").Raw())
	value.chain.assertOK(t)
	value.chain.reset()

	value.Schema(`{"type": "object"}`)
	value.chain.assertOK(t)
	value.chain.reset()

	value.Schema(`{"type": "array"}`)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Keys().ContainsOnly(keys...)
	value.chain.assertOK(t)
	value.chain.reset()

	value.Values().ContainsOnly(values...)
	value.chain.assertOK(t)
	value.chain.reset()

	assert.Equal(t, m["foo"], value.Value("foo").Raw())
	value.chain.assertOK(t)
	value.chain.reset()

	assert.Equal(t, m["bar"], value.Value("bar").Raw())
	value.chain.assertOK(t)
	value.chain.reset()

	assert.Equal(t, m["baz"], value.Value("baz").Raw())
	value.chain.assertOK(t)
	value.chain.reset()

	assert.Equal(t, nil, value.Value("BAZ").Raw())
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestObjectEmpty(t *testing.T) {
	reporter := newMockReporter(t)

	value1 := NewObject(reporter, nil)

	_ = value1
	value1.chain.assertFailed(t)
	value1.chain.reset()

	value2 := NewObject(reporter, map[string]interface{}{})

	value2.Empty()
	value2.chain.assertOK(t)
	value2.chain.reset()

	value2.NotEmpty()
	value2.chain.assertFailed(t)
	value2.chain.reset()

	value3 := NewObject(reporter, map[string]interface{}{"": nil})

	value3.Empty()
	value3.chain.assertFailed(t)
	value3.chain.reset()

	value3.NotEmpty()
	value3.chain.assertOK(t)
	value3.chain.reset()
}

func TestObjectEqualEmpty(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewObject(reporter, map[string]interface{}{})

	assert.Equal(t, map[string]interface{}{}, value.Raw())

	value.Equal(map[string]interface{}{})
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotEqual(map[string]interface{}{})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Equal(map[string]interface{}{"": nil})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqual(map[string]interface{}{"": nil})
	value.chain.assertOK(t)
	value.chain.reset()
}

func TestObjectEqual(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewObject(reporter, map[string]interface{}{"foo": 123.0})

	assert.Equal(t, map[string]interface{}{"foo": 123.0}, value.Raw())

	value.Equal(map[string]interface{}{})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqual(map[string]interface{}{})
	value.chain.assertOK(t)
	value.chain.reset()

	value.Equal(map[string]interface{}{"FOO": 123.0})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqual(map[string]interface{}{"FOO": 123.0})
	value.chain.assertOK(t)
	value.chain.reset()

	value.Equal(map[string]interface{}{"foo": 456.0})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqual(map[string]interface{}{"foo": 456.0})
	value.chain.assertOK(t)
	value.chain.reset()

	value.Equal(map[string]interface{}{"foo": 123.0})
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotEqual(map[string]interface{}{"foo": 123.0})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Equal(nil)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqual(nil)
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestObjectEqualStruct(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewObject(reporter, map[string]interface{}{
		"foo": 123,
		"bar": map[string]interface{}{
			"baz": []interface{}{true, false},
		},
	})

	type (
		Bar struct {
			Baz []bool `json:"baz"`
		}

		S struct {
			Foo int `json:"foo"`
			Bar Bar `json:"bar"`
		}
	)

	s := S{
		Foo: 123,
		Bar: Bar{
			Baz: []bool{true, false},
		},
	}

	value.Equal(s)
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotEqual(s)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Equal(S{})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqual(S{})
	value.chain.assertOK(t)
	value.chain.reset()
}

func TestObjectContainsKey(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewObject(reporter, map[string]interface{}{"foo": 123, "bar": ""})

	value.ContainsKey("foo")
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotContainsKey("foo")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.ContainsKey("bar")
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotContainsKey("bar")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.ContainsKey("BAR")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotContainsKey("BAR")
	value.chain.assertOK(t)
	value.chain.reset()
}

func TestObjectContainsValue(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewObject(reporter, map[string]interface{}{"foo": 123, "bar": "xxx"})

	value.ContainsValue(123)
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotContainsValue(123)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.ContainsValue("xxx")
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotContainsValue("xxx")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.ContainsValue("XXX")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotContainsValue("XXX")
	value.chain.assertOK(t)
	value.chain.reset()
}

func TestObjectContainsValueStruct(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewObject(reporter, map[string]interface{}{
		"foo": 123,
		"bar": []interface{}{"456", 789},
		"baz": map[string]interface{}{
			"a": map[string]interface{}{
				"b": 333,
				"c": 444,
			},
		},
	})

	type (
		A struct {
			B int `json:"b"`
			C int `json:"c"`
		}

		Baz struct {
			A A `json:"a"`
		}
	)

	barValue := []interface{}{"456", 789}
	bazValue := Baz{
		A: A{
			B: 333,
			C: 444,
		},
	}

	value.ContainsValue(barValue)
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotContainsValue(barValue)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.ContainsValue(bazValue)
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotContainsValue(bazValue)
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestObjectContainsSubsetSuccess(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewObject(reporter, map[string]interface{}{
		"foo": 123,
		"bar": []interface{}{"456", 789},
		"baz": map[string]interface{}{
			"a": map[string]interface{}{
				"b": 333,
				"c": 444,
			},
		},
	})

	submap1 := map[string]interface{}{
		"foo": 123,
		"bar": []interface{}{"456", 789},
	}

	value.ContainsSubset(submap1)
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotContainsSubset(submap1)
	value.chain.assertFailed(t)
	value.chain.reset()

	submap2 := map[string]interface{}{
		"bar": []interface{}{"456", 789},
		"baz": map[string]interface{}{
			"a": map[string]interface{}{
				"c": 444,
			},
		},
	}

	value.ContainsSubset(submap2)
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotContainsSubset(submap2)
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestObjectContainsSubsetFailed(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewObject(reporter, map[string]interface{}{
		"foo": 123,
		"bar": []interface{}{"456", 789},
		"baz": map[string]interface{}{
			"a": map[string]interface{}{
				"b": 333,
				"c": 444,
			},
		},
	})

	submap1 := map[string]interface{}{
		"foo": 123,
		"qux": 456,
	}

	value.ContainsSubset(submap1)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotContainsSubset(submap1)
	value.chain.assertOK(t)
	value.chain.reset()

	submap2 := map[string]interface{}{
		"foo": 123,
		"bar": []interface{}{"456", "789"},
	}

	value.ContainsSubset(submap2)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotContainsSubset(submap2)
	value.chain.assertOK(t)
	value.chain.reset()

	submap3 := map[string]interface{}{
		"baz": map[string]interface{}{
			"a": map[string]interface{}{
				"b": "333",
				"c": 444,
			},
		},
	}

	value.ContainsSubset(submap3)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotContainsSubset(submap3)
	value.chain.assertOK(t)
	value.chain.reset()

	value.ContainsSubset(nil)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotContainsSubset(nil)
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestObjectContainsSubsetStruct(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewObject(reporter, map[string]interface{}{
		"foo": 123,
		"bar": []interface{}{"456", 789},
		"baz": map[string]interface{}{
			"a": map[string]interface{}{
				"b": 333,
				"c": 444,
			},
		},
	})

	type (
		A struct {
			B int `json:"b"`
		}

		Baz struct {
			A A `json:"a"`
		}

		S struct {
			Foo int `json:"foo"`
			Baz Baz `json:"baz"`
		}
	)

	submap := S{
		Foo: 123,
		Baz: Baz{
			A: A{
				B: 333,
			},
		},
	}

	value.ContainsSubset(submap)
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotContainsSubset(submap)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.ContainsSubset(S{})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotContainsSubset(S{})
	value.chain.assertOK(t)
	value.chain.reset()
}

func TestObjectValueEqual(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewObject(reporter, map[string]interface{}{
		"foo": 123,
		"bar": []interface{}{"456", 789},
		"baz": map[string]interface{}{
			"a": "b",
		},
	})

	value.ValueEqual("foo", 123)
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotValueEqual("foo", 123)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.ValueEqual("bar", []interface{}{"456", 789})
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotValueEqual("bar", []interface{}{"456", 789})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.ValueEqual("baz", map[string]interface{}{"a": "b"})
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotValueEqual("baz", map[string]interface{}{"a": "b"})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.ValueEqual("baz", func() {})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotValueEqual("baz", func() {})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.ValueEqual("BAZ", 777)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotValueEqual("BAZ", 777)
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestObjectValueEqualStruct(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewObject(reporter, map[string]interface{}{
		"foo": 123,
		"bar": []interface{}{"456", 789},
		"baz": map[string]interface{}{
			"a": map[string]interface{}{
				"b": 333,
				"c": 444,
			},
		},
	})

	type (
		A struct {
			B int `json:"b"`
			C int `json:"c"`
		}

		Baz struct {
			A A `json:"a"`
		}
	)

	baz := Baz{
		A: A{
			B: 333,
			C: 444,
		},
	}

	value.ValueEqual("baz", baz)
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotValueEqual("baz", baz)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.ValueEqual("baz", Baz{})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotValueEqual("baz", Baz{})
	value.chain.assertOK(t)
	value.chain.reset()
}

func TestObjectConvertEqual(t *testing.T) {
	type (
		myMap map[string]interface{}
		myInt int
	)

	reporter := newMockReporter(t)

	value := NewObject(reporter, map[string]interface{}{"foo": 123})

	value.Equal(map[string]interface{}{"foo": "123"})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqual(map[string]interface{}{"foo": "123"})
	value.chain.assertOK(t)
	value.chain.reset()

	value.Equal(map[string]interface{}{"foo": 123.0})
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotEqual(map[string]interface{}{"foo": 123.0})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Equal(map[string]interface{}{"foo": 123})
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotEqual(map[string]interface{}{"foo": 123})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Equal(myMap{"foo": myInt(123)})
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotEqual(myMap{"foo": myInt(123)})
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestObjectConvertContainsSubset(t *testing.T) {
	type (
		myArray []interface{}
		myMap   map[string]interface{}
		myInt   int
	)

	reporter := newMockReporter(t)

	value := NewObject(reporter, map[string]interface{}{
		"foo": 123,
		"bar": []interface{}{"456", 789},
		"baz": map[string]interface{}{
			"a": "b",
		},
	})

	submap := myMap{
		"foo": myInt(123),
		"bar": myArray{"456", myInt(789)},
	}

	value.ContainsSubset(submap)
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotContainsSubset(submap)
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestObjectConvertValueEqual(t *testing.T) {
	type (
		myArray []interface{}
		myMap   map[string]interface{}
		myInt   int
	)

	reporter := newMockReporter(t)

	value := NewObject(reporter, map[string]interface{}{
		"foo": 123,
		"bar": []interface{}{"456", 789},
		"baz": map[string]interface{}{
			"a": "b",
		},
	})

	value.ValueEqual("bar", myArray{"456", myInt(789)})
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotValueEqual("bar", myArray{"456", myInt(789)})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.ValueEqual("baz", myMap{"a": "b"})
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotValueEqual("baz", myMap{"a": "b"})
	value.chain.assertFailed(t)
	value.chain.reset()
}
