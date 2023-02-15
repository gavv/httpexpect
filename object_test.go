package httpexpect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObject_FailedChain(t *testing.T) {
	check := func(value *Object) {
		value.chain.assertFailed(t)

		value.Path("$").chain.assertFailed(t)
		value.Schema("")
		value.Alias("foo")

		var target interface{}
		value.Decode(&target)

		value.Keys().chain.assertFailed(t)
		value.Values().chain.assertFailed(t)
		value.Value("foo").chain.assertFailed(t)

		value.IsEmpty()
		value.NotEmpty()
		value.IsEqual(nil)
		value.NotEqual(nil)
		value.InList(nil)
		value.NotInList(nil)
		value.ContainsKey("foo")
		value.NotContainsKey("foo")
		value.ContainsValue("foo")
		value.NotContainsValue("foo")
		value.ContainsSubset(nil)
		value.NotContainsSubset(nil)
		value.IsValueEqual("foo", nil)
		value.NotValueEqual("foo", nil)

		assert.NotNil(t, value.Iter())
		assert.Equal(t, 0, len(value.Iter()))

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

	t.Run("failed chain", func(t *testing.T) {
		chain := newMockChain(t)
		chain.setFailed()

		value := newObject(chain, map[string]interface{}{})

		check(value)
	})

	t.Run("nil value", func(t *testing.T) {
		chain := newMockChain(t)

		value := newObject(chain, nil)

		check(value)
	})

	t.Run("failed chain, nil value", func(t *testing.T) {
		chain := newMockChain(t)
		chain.setFailed()

		value := newObject(chain, nil)

		check(value)
	})
}

func TestObject_Constructors(t *testing.T) {
	test := map[string]interface{}{
		"foo": 100.0,
	}

	t.Run("reporter", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewObject(reporter, test)

		value.IsEqual(test)
		value.chain.assertNotFailed(t)
	})

	t.Run("config", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewObjectC(Config{
			Reporter: reporter,
		}, test)

		value.IsEqual(test)
		value.chain.assertNotFailed(t)
	})

	t.Run("chain", func(t *testing.T) {
		chain := newMockChain(t)

		value := newObject(chain, test)

		assert.NotSame(t, value.chain, chain)
		assert.Equal(t, value.chain.context.Path, chain.context.Path)
	})

	t.Run("invalid value", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewObject(reporter, nil)

		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})
}

func TestObject_Decode(t *testing.T) {
	t.Run("empty interface", func(t *testing.T) {
		reporter := newMockReporter(t)

		m := map[string]interface{}{
			"foo": 123.0,
			"bar": []interface{}{"123", 234.0},
			"baz": map[string]interface{}{
				"a": "b",
			},
		}

		value := NewObject(reporter, m)

		var target interface{}
		value.Decode(&target)

		value.chain.assertNotFailed(t)
		assert.Equal(t, target, m)
	})

	t.Run("map", func(t *testing.T) {
		reporter := newMockReporter(t)

		m := map[string]interface{}{
			"foo": 123.0,
			"bar": []interface{}{"123", 234.0},
			"baz": map[string]interface{}{
				"a": "b",
			},
		}

		value := NewObject(reporter, m)

		var target map[string]interface{}
		value.Decode(&target)

		value.chain.assertNotFailed(t)
		assert.Equal(t, target, m)
	})

	t.Run("struct", func(t *testing.T) {
		reporter := newMockReporter(t)

		type S struct {
			Foo int                    `json:"foo"`
			Bar []interface{}          `json:"bar"`
			Baz map[string]interface{} `json:"baz"`
			Bat struct{ A int }        `json:"bat"`
		}

		m := map[string]interface{}{
			"foo": 123,
			"bar": []interface{}{"123", 234.0},
			"baz": map[string]interface{}{
				"a": "b",
			},
			"bat": struct{ A int }{123},
		}

		value := NewObject(reporter, m)

		actualStruct := S{
			Foo: 123,
			Bar: []interface{}{"123", 234.0},
			Baz: map[string]interface{}{"a": "b"},
			Bat: struct{ A int }{123},
		}

		var target S
		value.Decode(&target)

		value.chain.assertNotFailed(t)
		assert.Equal(t, target, actualStruct)
	})

	t.Run("target is unmarshable", func(t *testing.T) {
		reporter := newMockReporter(t)

		m := map[string]interface{}{
			"foo": 123.0,
			"bar": []interface{}{"123", 234.0},
			"baz": map[string]interface{}{
				"a": "b",
			},
		}

		value := NewObject(reporter, m)

		value.Decode(123)

		value.chain.assertFailed(t)
	})

	t.Run("target is nil", func(t *testing.T) {
		reporter := newMockReporter(t)

		m := map[string]interface{}{
			"foo": 123.0,
			"bar": []interface{}{"123", 234.0},
			"baz": map[string]interface{}{
				"a": "b",
			},
		}

		value := NewObject(reporter, m)

		value.Decode(nil)

		value.chain.assertFailed(t)
	})
}

func TestObject_Alias(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewObject(reporter, map[string]interface{}{
		"foo": 100.0,
	})
	assert.Equal(t, []string{"Object()"}, value.chain.context.Path)
	assert.Equal(t, []string{"Object()"}, value.chain.context.AliasedPath)

	value.Alias("bar")
	assert.Equal(t, []string{"Object()"}, value.chain.context.Path)
	assert.Equal(t, []string{"bar"}, value.chain.context.AliasedPath)

	childValue := value.Values()
	assert.Equal(t, []string{"Object()", "Values()"}, childValue.chain.context.Path)
	assert.Equal(t, []string{"bar", "Values()"}, childValue.chain.context.AliasedPath)
}

func TestObject_Getters(t *testing.T) {
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

func TestObject_IsEmpty(t *testing.T) {
	t.Run("empty map", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewObject(reporter, map[string]interface{}{})

		value.IsEmpty()
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotEmpty()
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})

	t.Run("one empty element", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewObject(reporter, map[string]interface{}{"": nil})

		value.IsEmpty()
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEmpty()
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()
	})
}

func TestObject_IsEqual(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewObject(reporter, map[string]interface{}{})

		assert.Equal(t, map[string]interface{}{}, value.Raw())

		value.IsEqual(map[string]interface{}{})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotEqual(map[string]interface{}{})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.IsEqual(map[string]interface{}{"": nil})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqual(map[string]interface{}{"": nil})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()
	})

	t.Run("not empty", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewObject(reporter, map[string]interface{}{"foo": 123.0})

		assert.Equal(t, map[string]interface{}{"foo": 123.0}, value.Raw())

		value.IsEqual(map[string]interface{}{})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqual(map[string]interface{}{})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.IsEqual(map[string]interface{}{"FOO": 123.0})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqual(map[string]interface{}{"FOO": 123.0})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.IsEqual(map[string]interface{}{"foo": 456.0})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqual(map[string]interface{}{"foo": 456.0})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.IsEqual(map[string]interface{}{"foo": 123.0})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotEqual(map[string]interface{}{"foo": 123.0})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.IsEqual(nil)
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqual(nil)
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})

	t.Run("struct", func(t *testing.T) {
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

		value.IsEqual(s)
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotEqual(s)
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.IsEqual(S{})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqual(S{})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()
	})

	t.Run("canonization", func(t *testing.T) {
		type (
			myMap map[string]interface{}
			myInt int
		)

		reporter := newMockReporter(t)

		value := NewObject(reporter, map[string]interface{}{"foo": 123})

		value.IsEqual(map[string]interface{}{"foo": "123"})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqual(map[string]interface{}{"foo": "123"})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.IsEqual(map[string]interface{}{"foo": 123.0})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotEqual(map[string]interface{}{"foo": 123.0})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.IsEqual(map[string]interface{}{"foo": 123})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotEqual(map[string]interface{}{"foo": 123})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.IsEqual(myMap{"foo": myInt(123)})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotEqual(myMap{"foo": myInt(123)})
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})
}

func TestObject_InList(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewObject(reporter, map[string]interface{}{"foo": 123.0})

	assert.Equal(t, map[string]interface{}{"foo": 123.0}, value.Raw())

	value.InList()
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInList()
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InList(map[string]interface{}{})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInList(map[string]interface{}{})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.InList(
		map[string]interface{}{"FOO": 123.0},
		map[string]interface{}{"BAR": 456.0},
	)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInList(
		map[string]interface{}{"FOO": 123.0},
		map[string]interface{}{"BAR": 456.0},
	)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.InList(
		map[string]interface{}{"foo": 456.0},
		map[string]interface{}{"bar": 123.0},
	)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInList(
		map[string]interface{}{"foo": 456.0},
		map[string]interface{}{"bar": 123.0},
	)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.InList(
		map[string]interface{}{"foo": 123.0},
		map[string]interface{}{"bar": 456.0},
	)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotInList(
		map[string]interface{}{"foo": 123.0},
		map[string]interface{}{"bar": 456.0},
	)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InList(struct {
		Foo float64 `json:"foo"`
	}{Foo: 123.00})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotInList(struct {
		Foo float64 `json:"foo"`
	}{Foo: 123.00})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InList(map[string]interface{}{"bar": 123.0}, "NOT OBJECT")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInList(map[string]interface{}{"bar": 123.0}, "NOT OBJECT")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InList(map[string]interface{}{"foo": 123.0}, "NOT OBJECT")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInList(map[string]interface{}{"foo": 123.0}, "NOT OBJECT")
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestObject_ContainsKey(t *testing.T) {
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

func TestObject_ContainsValue(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
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
	})

	t.Run("struct", func(t *testing.T) {
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
	})
}

func TestObject_ContainsSubset(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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
	})

	t.Run("failure", func(t *testing.T) {
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
	})

	t.Run("struct", func(t *testing.T) {
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
	})

	t.Run("canonization", func(t *testing.T) {
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
	})
}

func TestObject_IsValueEqual(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewObject(reporter, map[string]interface{}{
			"foo": 123,
			"bar": []interface{}{"456", 789},
			"baz": map[string]interface{}{
				"a": "b",
			},
		})

		value.IsValueEqual("foo", 123)
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotValueEqual("foo", 123)
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.IsValueEqual("bar", []interface{}{"456", 789})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotValueEqual("bar", []interface{}{"456", 789})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.IsValueEqual("baz", map[string]interface{}{"a": "b"})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotValueEqual("baz", map[string]interface{}{"a": "b"})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.IsValueEqual("baz", func() {})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotValueEqual("baz", func() {})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.IsValueEqual("BAZ", 777)
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotValueEqual("BAZ", 777)
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})

	t.Run("struct", func(t *testing.T) {
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

		value.IsValueEqual("baz", baz)
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotValueEqual("baz", baz)
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.IsValueEqual("baz", Baz{})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotValueEqual("baz", Baz{})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()
	})

	t.Run("canonization", func(t *testing.T) {
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

		value.IsValueEqual("bar", myArray{"456", myInt(789)})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotValueEqual("bar", myArray{"456", myInt(789)})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.IsValueEqual("baz", myMap{"a": "b"})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotValueEqual("baz", myMap{"a": "b"})
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})
}

func TestObject_Every(t *testing.T) {
	t.Run("check value", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		object := NewObject(reporter, map[string]interface{}{
			"foo": "123",
			"bar": "456",
			"baz": "b",
		})

		invoked := 0
		object.Every(func(_ string, value *Value) {
			invoked++
			value.String().NotEmpty()
		})

		assert.Equal(t, 3, invoked)
		object.chain.assertNotFailed(ts)
	})

	t.Run("check key", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		object := NewObject(reporter, map[string]interface{}{
			"foo": "123",
			"bar": "456",
			"baz": "b",
		})

		invoked := 0
		object.Every(func(key string, value *Value) {
			if v, ok := value.Raw().(string); ok {
				invoked++
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

		assert.Equal(t, 3, invoked)
		object.chain.assertNotFailed(ts)
	})

	t.Run("empty object", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		object := NewObject(reporter, map[string]interface{}{})

		invoked := 0
		object.Every(func(_ string, value *Value) {
			invoked++
			value.String().NotEmpty()
		})

		assert.Equal(t, 0, invoked)
		object.chain.assertNotFailed(ts)
	})

	t.Run("one assertion fails", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		object := NewObject(reporter, map[string]interface{}{"foo": "", "bar": "bar"})

		invoked := 0
		object.Every(func(_ string, val *Value) {
			invoked++
			val.String().NotEmpty()
		})

		assert.Equal(t, 2, invoked)
		object.chain.assertFailed(ts)
	})

	t.Run("all assertions fail", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		object := NewObject(reporter, map[string]interface{}{"foo": "", "bar": ""})

		invoked := 0
		object.Every(func(_ string, val *Value) {
			invoked++
			val.String().NotEmpty()
		})

		assert.Equal(t, 2, invoked)
		object.chain.assertFailed(ts)
	})

	t.Run("call order", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		object := NewObject(reporter, map[string]interface{}{
			"bar": "123",
			"baz": "456",
			"foo": "foo",
			"foz": "foo",
			"b":   "789",
			"c":   "987",
		})

		var actualOrder []string
		object.Every(func(key string, val *Value) {
			actualOrder = append(actualOrder, key)
		})

		expectedOrder := []string{"b", "bar", "baz", "c", "foo", "foz"}
		assert.Equal(t, expectedOrder, actualOrder)
	})
}

func TestObject_Transform(t *testing.T) {
	t.Run("check index", func(ts *testing.T) {
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

	t.Run("transform value", func(ts *testing.T) {
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
			map[string]interface{}{
				"foo": "Hello 123",
				"bar": "Hello 456",
				"baz": "Hello b",
			},
			newObject.value,
		)
	})

	t.Run("empty object", func(ts *testing.T) {
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

	t.Run("call order", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		object := NewObject(reporter, map[string]interface{}{
			"foo": "123",
			"bar": "456",
			"b":   "456",
			"baz": "baz",
		})

		actualOrder := []string{}
		object.Transform(func(key string, value interface{}) interface{} {
			actualOrder = append(actualOrder, key)
			return value
		})

		expectedOrder := []string{"b", "bar", "baz", "foo"}
		assert.Equal(t, expectedOrder, actualOrder)
	})

	t.Run("invalid argument", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		object := NewObject(reporter, map[string]interface{}{
			"foo": "123",
			"bar": "456",
			"baz": "b",
		})

		newObject := object.Transform(nil)

		newObject.chain.assertFailed(t)
	})
}

func TestObject_Filter(t *testing.T) {
	t.Run("elements of same type", func(ts *testing.T) {
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

	t.Run("elements of different types", func(ts *testing.T) {
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

	t.Run("empty object", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{})

		filteredObject := object.Filter(func(key string, value *Value) bool {
			return false
		})
		assert.Equal(t, map[string]interface{}{}, filteredObject.Raw())

		filteredObject.chain.assertNotFailed(t)
		object.chain.assertNotFailed(t)
	})

	t.Run("no match", func(ts *testing.T) {
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

	t.Run("assertion fails", func(ts *testing.T) {
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

	t.Run("call order", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo":  "bar",
			"baz":  "qux",
			"bar":  "qux",
			"b":    "qux",
			"quux": "corge",
		})

		var actualOrder []string
		object.Filter(func(key string, value *Value) bool {
			actualOrder = append(actualOrder, key)
			return false
		})

		expectedOrder := []string{"b", "bar", "baz", "foo", "quux"}
		assert.Equal(t, expectedOrder, actualOrder)
	})
}

func TestObject_Find(t *testing.T) {
	t.Run("elements of same type", func(ts *testing.T) {
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

	t.Run("elements of different types", func(ts *testing.T) {
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

	t.Run("no match", func(ts *testing.T) {
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

	t.Run("empty object", func(ts *testing.T) {
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

	t.Run("predicate returns true, assertion fails, no match", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo": 1,
			"bar": 2,
		})

		foundValue := object.Find(func(key string, value *Value) bool {
			value.String()
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

	t.Run("predicate returns true, assertion fails, have match", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo": 1,
			"bar": 2,
			"baz": "str",
		})

		foundValue := object.Find(func(key string, value *Value) bool {
			value.String()
			return true
		})

		assert.Equal(t, "str", foundValue.Raw())
		assert.Equal(t, object.Raw(), map[string]interface{}{
			"foo": 1.0,
			"bar": 2.0,
			"baz": "str",
		})

		foundValue.chain.assertNotFailed(t)
		object.chain.assertNotFailed(t)
	})

	t.Run("invalid argument", func(ts *testing.T) {
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

func TestObject_FindAll(t *testing.T) {
	t.Run("elements of same type", func(ts *testing.T) {
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

	t.Run("elements of different types", func(ts *testing.T) {
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

	t.Run("no match", func(ts *testing.T) {
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

	t.Run("empty object", func(ts *testing.T) {
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

	t.Run("predicate returns true, assertion fails, no match", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo": 1,
			"bar": 2,
		})

		foundValues := object.FindAll(func(key string, value *Value) bool {
			value.String()
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

	t.Run("predicate returns true, assertion fails, have matches", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo":  "bar",
			"baz":  1,
			"qux":  2,
			"quux": "corge",
		})

		foundValues := object.FindAll(func(key string, value *Value) bool {
			value.String()
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

	t.Run("invalid argument", func(ts *testing.T) {
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

func TestObject_NotFind(t *testing.T) {
	t.Run("no match", func(ts *testing.T) {
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

		assert.Same(t, object, afterObject)
		object.chain.assertNotFailed(t)
	})

	t.Run("have match", func(ts *testing.T) {
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

		assert.Same(t, object, afterObject)
		object.chain.assertFailed(t)
	})

	t.Run("empty object", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{})

		afterObject := object.NotFind(func(key string, value *Value) bool {
			return key == "corge"
		})

		assert.Same(t, object, afterObject)
		object.chain.assertNotFailed(t)
	})

	t.Run("predicate returns true, assertion fails, no match", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo": 1,
			"bar": 2,
		})

		afterObject := object.NotFind(func(key string, value *Value) bool {
			value.String()
			return true
		})

		assert.Same(t, object, afterObject)
		object.chain.assertNotFailed(t)
	})

	t.Run("predicate returns true, assertion fails, have match", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo": 1,
			"bar": 2,
			"baz": "str",
		})

		afterObject := object.NotFind(func(key string, value *Value) bool {
			value.String()
			return true
		})

		assert.Same(t, object, afterObject)
		object.chain.assertFailed(t)
	})

	t.Run("invalid argument", func(ts *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo": 1,
			"bar": 2,
		})

		afterObject := object.NotFind(nil)

		assert.Same(t, object, afterObject)
		object.chain.assertFailed(t)
	})
}
