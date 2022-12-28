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
		assert.NotNil(t, value.Iter())

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
		value.Every(func(_ string, value *Value) {
			value.String().NotEmpty()
		})
		value.Transform(func(key string, value interface{}) interface{} {
			return nil
		})
		value.Filter(func(_ string, value *Value) bool {
			value.String().NotEmpty()
			return true
		})
		value.Find(func(key string, value *Value) bool {
			value.String().NotEmpty()
			return true
		})
		value.FindAll(func(key string, value *Value) bool {
			value.String().NotEmpty()
			return true
		})
		value.NotFind(func(key string, value *Value) bool {
			value.String().NotEmpty()
			return true
		})
	}

	t.Run("failed_chain", func(t *testing.T) {
		chain := newMockChain(t)
		chain.fail(mockFailure())

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
		chain.fail(mockFailure())

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
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	assert.Equal(t, m, value.Path("$").Raw())
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Schema(`{"type": "object"}`)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Schema(`{"type": "array"}`)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Keys().ContainsOnly(keys...)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Values().ContainsOnly(values...)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	assert.Equal(t, m["foo"], value.Value("foo").Raw())
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	assert.Equal(t, m["bar"], value.Value("bar").Raw())
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	assert.Equal(t, m["baz"], value.Value("baz").Raw())
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	assert.Equal(t, nil, value.Value("BAZ").Raw())
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	it := value.Iter()
	assert.Equal(t, 3, len(it))
	assert.Equal(t, it["foo"].value, value.Value("foo").Raw())
	assert.Equal(t, it["bar"].value, value.Value("bar").Raw())
	it["foo"].chain.assertNotFailed(t)
	it["bar"].chain.assertNotFailed(t)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}

func TestObjectEmpty(t *testing.T) {
	reporter := newMockReporter(t)

	value1 := NewObject(reporter, nil)

	_ = value1
	value1.chain.assertFailed(t)
	value1.chain.clearFailed()

	value2 := NewObject(reporter, map[string]interface{}{})

	value2.Empty()
	value2.chain.assertNotFailed(t)
	value2.chain.clearFailed()

	value2.NotEmpty()
	value2.chain.assertFailed(t)
	value2.chain.clearFailed()

	value3 := NewObject(reporter, map[string]interface{}{"": nil})

	value3.Empty()
	value3.chain.assertFailed(t)
	value3.chain.clearFailed()

	value3.NotEmpty()
	value3.chain.assertNotFailed(t)
	value3.chain.clearFailed()
}

func TestObjectEqualEmpty(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewObject(reporter, map[string]interface{}{})

	assert.Equal(t, map[string]interface{}{}, value.Raw())

	value.Equal(map[string]interface{}{})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotEqual(map[string]interface{}{})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Equal(map[string]interface{}{"": nil})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual(map[string]interface{}{"": nil})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}

func TestObjectEqual(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewObject(reporter, map[string]interface{}{"foo": 123.0})

	assert.Equal(t, map[string]interface{}{"foo": 123.0}, value.Raw())

	value.Equal(map[string]interface{}{})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual(map[string]interface{}{})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Equal(map[string]interface{}{"FOO": 123.0})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual(map[string]interface{}{"FOO": 123.0})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Equal(map[string]interface{}{"foo": 456.0})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual(map[string]interface{}{"foo": 456.0})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Equal(map[string]interface{}{"foo": 123.0})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotEqual(map[string]interface{}{"foo": 123.0})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Equal(nil)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual(nil)
	value.chain.assertFailed(t)
	value.chain.clearFailed()
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
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotEqual(s)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Equal(S{})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual(S{})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}

func TestObjectContainsKey(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewObject(reporter, map[string]interface{}{"foo": 123, "bar": ""})

	value.ContainsKey("foo")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotContainsKey("foo")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.ContainsKey("bar")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotContainsKey("bar")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.ContainsKey("BAR")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotContainsKey("BAR")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}

func TestObjectContainsValue(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewObject(reporter, map[string]interface{}{"foo": 123, "bar": "xxx"})

	value.ContainsValue(123)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotContainsValue(123)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.ContainsValue("xxx")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotContainsValue("xxx")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.ContainsValue("XXX")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotContainsValue("XXX")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
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
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotContainsValue(barValue)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.ContainsValue(bazValue)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotContainsValue(bazValue)
	value.chain.assertFailed(t)
	value.chain.clearFailed()
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
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotContainsSubset(submap1)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	submap2 := map[string]interface{}{
		"bar": []interface{}{"456", 789},
		"baz": map[string]interface{}{
			"a": map[string]interface{}{
				"c": 444,
			},
		},
	}

	value.ContainsSubset(submap2)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotContainsSubset(submap2)
	value.chain.assertFailed(t)
	value.chain.clearFailed()
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
	value.chain.clearFailed()

	value.NotContainsSubset(submap1)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	submap2 := map[string]interface{}{
		"foo": 123,
		"bar": []interface{}{"456", "789"},
	}

	value.ContainsSubset(submap2)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotContainsSubset(submap2)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

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
	value.chain.clearFailed()

	value.NotContainsSubset(submap3)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.ContainsSubset(nil)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotContainsSubset(nil)
	value.chain.assertFailed(t)
	value.chain.clearFailed()
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
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotContainsSubset(submap)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.ContainsSubset(S{})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotContainsSubset(S{})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
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
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotValueEqual("foo", 123)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.ValueEqual("bar", []interface{}{"456", 789})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotValueEqual("bar", []interface{}{"456", 789})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.ValueEqual("baz", map[string]interface{}{"a": "b"})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotValueEqual("baz", map[string]interface{}{"a": "b"})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.ValueEqual("baz", func() {})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotValueEqual("baz", func() {})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.ValueEqual("BAZ", 777)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotValueEqual("BAZ", 777)
	value.chain.assertFailed(t)
	value.chain.clearFailed()
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
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotValueEqual("baz", baz)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.ValueEqual("baz", Baz{})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotValueEqual("baz", Baz{})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
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
	value.chain.clearFailed()

	value.NotEqual(map[string]interface{}{"foo": "123"})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Equal(map[string]interface{}{"foo": 123.0})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotEqual(map[string]interface{}{"foo": 123.0})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Equal(map[string]interface{}{"foo": 123})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotEqual(map[string]interface{}{"foo": 123})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Equal(myMap{"foo": myInt(123)})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotEqual(myMap{"foo": myInt(123)})
	value.chain.assertFailed(t)
	value.chain.clearFailed()
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
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotContainsSubset(submap)
	value.chain.assertFailed(t)
	value.chain.clearFailed()
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
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotValueEqual("bar", myArray{"456", myInt(789)})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.ValueEqual("baz", myMap{"a": "b"})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotValueEqual("baz", myMap{"a": "b"})
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestObjectEvery(t *testing.T) {
	t.Run("Check Validation", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		object := NewObject(reporter, map[string]interface{}{
			"foo": "123",
			"bar": "456",
			"baz": "b",
		})
		object.Every(func(_ string, value *Value) {
			value.String().NotEmpty()
		})
		object.chain.assertNotFailed(ts)
	})

	t.Run("Empty Object", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		object := NewObject(reporter, map[string]interface{}{})
		object.Every(func(_ string, value *Value) {
			value.String().NotEmpty()
		})
		object.chain.assertNotFailed(ts)
	})

	t.Run("Test Keys", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		object := NewObject(reporter, map[string]interface{}{
			"foo": "123",
			"bar": "456",
			"baz": "b",
		})
		object.Every(func(key string, value *Value) {
			if v, ok := value.Raw().(string); ok {
				switch v {
				case "123":
					assert.Equal(ts, "foo", key)
				case "456":
					assert.Equal(ts, "bar", key)
				case "baz":
					assert.Equal(ts, "baz", key)
				}
			}
		})
		object.chain.assertNotFailed(ts)
	})

	t.Run("Assertion failed for any", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		object := NewObject(reporter, map[string]interface{}{"foo": "", "bar": "bar"})
		invoked := 0
		object.Every(func(_ string, val *Value) {
			invoked++
			val.String().NotEmpty()
		})
		object.chain.assertFailed(ts)
		assert.Equal(t, 2, invoked)
	})

	t.Run("Assertion failed for all", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		object := NewObject(reporter, map[string]interface{}{"foo": "", "bar": ""})
		invoked := 0
		object.Every(func(_ string, val *Value) {
			invoked++
			val.String().NotEmpty()
		})
		object.chain.assertFailed(ts)
		assert.Equal(t, 2, invoked)
	})
}

func TestObjectTransform(t *testing.T) {
	t.Run("Add Hello", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		object := NewObject(reporter, map[string]interface{}{
			"foo": "123",
			"bar": "456",
			"baz": "b",
		})
		newObject := object.Transform(func(_ string, value interface{}) interface{} {
			if v, ok := value.(string); ok {
				return "Hello " + v
			}
			return nil
		})
		assert.Equal(t,
			map[string]interface{}{"foo": "Hello 123", "bar": "Hello 456", "baz": "Hello b"},
			newObject.value,
		)
	})

	t.Run("Chain fail on nil function value", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		object := NewObject(reporter, map[string]interface{}{
			"foo": "123",
			"bar": "456",
			"baz": "b",
		})
		newObject := object.Transform(nil)
		newObject.chain.assertFailed(reporter)
	})

	t.Run("Empty object", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		object := NewObject(reporter, map[string]interface{}{
			"foo": "123",
			"bar": "456",
			"baz": "b",
		})
		newObject := object.Transform(func(_ string, value interface{}) interface{} {
			if v, ok := value.(string); ok {
				return "Hello " + v
			}
			return nil
		})
		newObject.chain.assertNotFailed(ts)
	})

	t.Run("Test correct index", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		object := NewObject(reporter, map[string]interface{}{
			"foo": "123",
			"bar": "456",
			"baz": "baz",
		})
		newObject := object.Transform(func(key string, value interface{}) interface{} {
			if v, ok := value.(string); ok {
				switch v {
				case "123":
					assert.Equal(ts, "foo", key)
				case "456":
					assert.Equal(ts, "bar", key)
				case "baz":
					assert.Equal(ts, "baz", key)
				}
			}
			return value
		})
		newObject.chain.assertNotFailed(ts)
	})
}

func TestObjectFilter(t *testing.T) {
	t.Run("Filter an object of elements of the same type and validate", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo": "bar",
			"baz": "qux", "quux": "corge",
		})
		filteredObject := object.Filter(func(key string, value *Value) bool {
			return value.Raw() != "qux" && key != "quux"
		})
		assert.Equal(t, map[string]interface{}{"foo": "bar"}, filteredObject.Raw())
		assert.Equal(t, object.Raw(), map[string]interface{}{
			"foo": "bar",
			"baz": "qux", "quux": "corge",
		})

		filteredObject.chain.assertNotFailed(t)
		object.chain.assertNotFailed(t)
	})

	t.Run("Filter throws when an assertion within predicate fails", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo": "bar", "baz": 6.0,
			"qux": "quux",
		})
		filteredObject := object.Filter(func(key string, value *Value) bool {
			stringifiedValue := value.String().NotEmpty().Raw()
			return stringifiedValue != "bar"
		})
		assert.Equal(t, map[string]interface{}{"qux": "quux"}, filteredObject.Raw())
		assert.Equal(t, object.Raw(), map[string]interface{}{
			"foo": "bar", "baz": 6.0,
			"qux": "quux",
		})

		filteredObject.chain.assertNotFailed(t)
		object.chain.assertNotFailed(t)
	})

	t.Run("Filter an object of different types and validate", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo": "bar",
			"baz": 3.0, "qux": false,
		})
		filteredObject := object.Filter(func(key string, value *Value) bool {
			return value.Raw() != "bar" && value.Raw() != 3.0
		})
		assert.Equal(t, map[string]interface{}{"qux": false}, filteredObject.Raw())
		assert.Equal(t, object.Raw(), map[string]interface{}{
			"foo": "bar",
			"baz": 3.0, "qux": false,
		})

		filteredObject.chain.assertNotFailed(t)
		object.chain.assertNotFailed(t)
	})

	t.Run("Filter an empty object", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{})
		filteredObject := object.Filter(func(key string, value *Value) bool {
			return false
		})
		assert.Equal(t, map[string]interface{}{}, filteredObject.Raw())

		filteredObject.chain.assertNotFailed(t)
		object.chain.assertNotFailed(t)
	})

	t.Run("Filter should return an empty non-nil object if no items pass",
		func(ts *testing.T) {
			reporter := newMockReporter(t)
			object := NewObject(reporter, map[string]interface{}{
				"foo": "bar",
				"baz": "qux", "quux": "corge",
			})
			filteredObject := object.Filter(func(key string, value *Value) bool {
				return false
			})
			assert.Equal(t, map[string]interface{}{}, filteredObject.Raw())
			assert.Equal(t, object.Raw(), map[string]interface{}{
				"foo": "bar",
				"baz": "qux", "quux": "corge",
			})

			filteredObject.chain.assertNotFailed(t)
			object.chain.assertNotFailed(t)
		})
}

func TestObjectFind(t *testing.T) {
	t.Run("Find an object of elements of the same type and validate", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo":  "bar",
			"baz":  "qux",
			"quux": "corge",
		})
		foundValue := object.Find(func(key string, value *Value) bool {
			return key == "baz"
		})
		assert.Equal(t, "qux", foundValue.Raw())
		assert.Equal(t, object.Raw(), map[string]interface{}{
			"foo":  "bar",
			"baz":  "qux",
			"quux": "corge",
		})

		foundValue.chain.assertNotFailed(t)
		object.chain.assertNotFailed(t)
	})

	t.Run("Find an object of elements of the multi type and validate", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo":  "bar",
			"baz":  true,
			"qux":  -1,
			"quux": 2,
		})
		foundValue := object.Find(func(key string, value *Value) bool {
			n := value.Number().Raw()
			return n > 1
		})
		assert.Equal(t, 2.0, foundValue.Raw())
		assert.Equal(t, object.Raw(), map[string]interface{}{
			"foo":  "bar",
			"baz":  true,
			"qux":  -1.0,
			"quux": 2.0,
		})

		foundValue.chain.assertNotFailed(t)
		object.chain.assertNotFailed(t)
	})

	t.Run("No match", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo":  "bar",
			"baz":  true,
			"qux":  -1,
			"quux": 2,
		})
		foundValue := object.Find(func(key string, value *Value) bool {
			n := value.Number().Raw()
			return n == 3
		})
		assert.Equal(t, nil, foundValue.Raw())
		assert.Equal(t, object.Raw(), map[string]interface{}{
			"foo":  "bar",
			"baz":  true,
			"qux":  -1.0,
			"quux": 2.0,
		})

		foundValue.chain.assertFailed(t)
		object.chain.assertFailed(t)
	})

	t.Run("Empty Object", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{})
		foundValue := object.Find(func(key string, value *Value) bool {
			n := value.Number().Raw()
			return n == 3
		})
		assert.Equal(t, nil, foundValue.Raw())
		assert.Equal(t, object.Raw(), map[string]interface{}{})

		foundValue.chain.assertFailed(t)
		object.chain.assertFailed(t)
	})

	t.Run("When predicate returns true, but assertion fails, predicate is failed",
		func(ts *testing.T) {
			reporter := newMockReporter(t)
			object := NewObject(reporter, map[string]interface{}{
				"foo": 1,
				"bar": 2,
			})
			foundValue := object.Find(func(key string, value *Value) bool {
				value.String().Raw()
				return true
			})
			assert.Equal(t, nil, foundValue.Raw())
			assert.Equal(t, object.Raw(), map[string]interface{}{
				"foo": 1.0,
				"bar": 2.0,
			})

			foundValue.chain.assertFailed(t)
			object.chain.assertFailed(t)
		})

	t.Run("Predicate func is nil", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo": 1,
			"bar": 2,
		})
		foundValue := object.Find(nil)
		assert.Equal(t, nil, foundValue.Raw())
		assert.Equal(t, object.Raw(), map[string]interface{}{
			"foo": 1.0,
			"bar": 2.0,
		})

		foundValue.chain.assertFailed(t)
		object.chain.assertFailed(t)
	})
}
func TestObjectFindAll(t *testing.T) {
	t.Run("Find values in array of the same type", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo":  "bar",
			"baz":  "qux",
			"quux": "corge",
		})
		foundValues := object.FindAll(func(key string, value *Value) bool {
			return key == "baz" || key == "quux"
		})

		actual := []interface{}{}
		for _, value := range foundValues {
			actual = append(actual, value.Raw())
		}

		assert.Equal(t, []interface{}{"qux", "corge"}, actual)
		assert.Equal(t, object.Raw(), map[string]interface{}{
			"foo":  "bar",
			"baz":  "qux",
			"quux": "corge",
		})

		for _, value := range foundValues {
			value.chain.assertNotFailed(t)
		}
		object.chain.assertNotFailed(t)
	})

	t.Run("Find values in array of the multi types", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo":   "bar",
			"baz":   6,
			"qux":   "quux",
			"corge": "grault",
		})
		foundValues := object.FindAll(func(key string, value *Value) bool {
			value.String().NotEmpty()
			return key != "qux"
		})

		actual := []interface{}{}
		for _, value := range foundValues {
			actual = append(actual, value.Raw())
		}

		assert.Equal(t, []interface{}{"grault", "bar"}, actual)
		assert.Equal(t, object.Raw(), map[string]interface{}{
			"foo":   "bar",
			"baz":   6.0,
			"qux":   "quux",
			"corge": "grault",
		})

		for _, value := range foundValues {
			value.chain.assertNotFailed(t)
		}
		object.chain.assertNotFailed(t)
	})

	t.Run("No match", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo":  "bar",
			"baz":  true,
			"qux":  -1,
			"quux": 2,
		})
		foundValues := object.FindAll(func(key string, value *Value) bool {
			return value.Number().Raw() == 3.0
		})

		actual := []interface{}{}
		for _, value := range foundValues {
			actual = append(actual, value.Raw())
		}

		assert.Equal(t, []interface{}{}, actual)
		assert.Equal(t, object.Raw(), map[string]interface{}{
			"foo":  "bar",
			"baz":  true,
			"qux":  -1.0,
			"quux": 2.0,
		})

		for _, value := range foundValues {
			value.chain.assertNotFailed(t)
		}
		object.chain.assertNotFailed(t)
	})

	t.Run("Empty array", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{})
		foundValues := object.FindAll(func(key string, value *Value) bool {
			return value.Number().Raw() == 3.0
		})

		actual := []interface{}{}
		for _, value := range foundValues {
			actual = append(actual, value.Raw())
		}

		assert.Equal(t, []interface{}{}, actual)
		assert.Equal(t, object.Raw(), map[string]interface{}{})

		for _, value := range foundValues {
			value.chain.assertNotFailed(t)
		}
		object.chain.assertNotFailed(t)
	})

	t.Run("When predicate returns true, but assertion fails, predicate is failed",
		func(ts *testing.T) {
			reporter := newMockReporter(t)
			object := NewObject(reporter, map[string]interface{}{
				"foo": 1,
				"bar": 2,
			})
			foundValues := object.FindAll(func(key string, value *Value) bool {
				value.String().Raw()
				return true
			})

			actual := []interface{}{}
			for _, value := range foundValues {
				actual = append(actual, value.Raw())
			}

			assert.Equal(t, []interface{}{}, actual)
			assert.Equal(t, object.Raw(), map[string]interface{}{
				"foo": 1.0,
				"bar": 2.0,
			})

			for _, value := range foundValues {
				value.chain.assertNotFailed(t)
			}
			object.chain.assertNotFailed(t)
		})

	t.Run("Assertion failure does not affect subsequent matches", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo":  "bar",
			"baz":  1,
			"qux":  2,
			"quux": "corge",
		})
		foundValues := object.FindAll(func(key string, value *Value) bool {
			value.String().Raw()
			return true
		})

		actual := []interface{}{}
		for _, value := range foundValues {
			actual = append(actual, value.Raw())
		}

		assert.Equal(t, []interface{}{"bar", "corge"}, actual)
		assert.Equal(t, object.Raw(), map[string]interface{}{
			"foo":  "bar",
			"baz":  1.0,
			"qux":  2.0,
			"quux": "corge",
		})

		for _, value := range foundValues {
			value.chain.assertNotFailed(t)
		}
		object.chain.assertNotFailed(t)
	})

	t.Run("Predicate func is nil", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo": 1,
			"bar": 2,
		})
		foundValues := object.FindAll(nil)

		actual := []interface{}{}
		for _, value := range foundValues {
			actual = append(actual, value.Raw())
		}

		assert.Equal(t, []interface{}{}, actual)
		assert.Equal(t, object.Raw(), map[string]interface{}{
			"foo": 1.0,
			"bar": 2.0,
		})

		for _, value := range foundValues {
			value.chain.assertFailed(t)
		}
		object.chain.assertFailed(t)
	})
}

func TestObjectNotFind(t *testing.T) {
	t.Run("Succeeds if no element matched predicate", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo":  "bar",
			"baz":  true,
			"qux":  -1,
			"quux": 2,
		})
		afterObject := object.NotFind(func(key string, value *Value) bool {
			return key == "corge"
		})
		assert.Equal(t, map[string]interface{}{
			"foo":  "bar",
			"baz":  true,
			"qux":  -1.0,
			"quux": 2.0,
		}, afterObject.Raw())
		assert.Equal(t, object.Raw(), map[string]interface{}{
			"foo":  "bar",
			"baz":  true,
			"qux":  -1.0,
			"quux": 2.0,
		})

		afterObject.chain.assertNotFailed(t)
		object.chain.assertNotFailed(t)
	})

	t.Run("Fails if there is a match", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo":  "bar",
			"baz":  true,
			"qux":  -1,
			"quux": 2,
		})
		afterObject := object.NotFind(func(key string, value *Value) bool {
			return key == "qux"
		})
		assert.Equal(t, map[string]interface{}(nil), afterObject.Raw())
		assert.Equal(t, object.Raw(), map[string]interface{}{
			"foo":  "bar",
			"baz":  true,
			"qux":  -1.0,
			"quux": 2.0,
		})

		afterObject.chain.assertFailed(t)
		object.chain.assertFailed(t)
	})

	t.Run("Empty object", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{})
		afterObject := object.NotFind(func(key string, value *Value) bool {
			return key == "corge"
		})
		assert.Equal(t, map[string]interface{}{}, afterObject.Raw())
		assert.Equal(t, object.Raw(), map[string]interface{}{})

		afterObject.chain.assertNotFailed(t)
		object.chain.assertNotFailed(t)
	})

	t.Run("When predicate returns true, but assertion fails, predicate is failed",
		func(ts *testing.T) {
			reporter := newMockReporter(t)
			object := NewObject(reporter, map[string]interface{}{
				"foo": 1,
				"bar": 2,
			})
			afterObject := object.NotFind(func(key string, value *Value) bool {
				value.String().Raw()
				return true
			})
			assert.Equal(t, map[string]interface{}{
				"foo": 1.0,
				"bar": 2.0,
			}, afterObject.Raw())
			assert.Equal(t, object.Raw(), map[string]interface{}{
				"foo": 1.0,
				"bar": 2.0,
			})

			afterObject.chain.assertNotFailed(t)
			object.chain.assertNotFailed(t)
		})

	t.Run("Predicate func is nil", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo": 1,
			"bar": 2,
		})
		afterObject := object.NotFind(nil)
		assert.Equal(t, map[string]interface{}(nil), afterObject.Raw())
		assert.Equal(t, object.Raw(), map[string]interface{}{
			"foo": 1.0,
			"bar": 2.0,
		})

		afterObject.chain.assertFailed(t)
		object.chain.assertFailed(t)
	})
}
