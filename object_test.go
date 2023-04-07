package httpexpect

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObject_FailedChain(t *testing.T) {
	check := func(value *Object) {
		value.chain.assert(t, failure)

		value.Path("$").chain.assert(t, failure)
		value.Schema("")
		value.Alias("foo")

		var target interface{}
		value.Decode(&target)

		value.Keys().chain.assert(t, failure)
		value.Values().chain.assert(t, failure)
		value.Value("foo").chain.assert(t, failure)

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
		value.HasValue("foo", nil)
		value.NotHasValue("foo", nil)

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
		chain := newMockChain(t, flagFailed)
		value := newObject(chain, map[string]interface{}{})

		check(value)
	})

	t.Run("nil value", func(t *testing.T) {
		chain := newMockChain(t)
		value := newObject(chain, nil)

		check(value)
	})

	t.Run("failed chain, nil value", func(t *testing.T) {
		chain := newMockChain(t, flagFailed)
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
		value.chain.assert(t, success)
	})

	t.Run("config", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewObjectC(Config{
			Reporter: reporter,
		}, test)

		value.IsEqual(test)
		value.chain.assert(t, success)
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

		value.chain.assert(t, failure)
	})
}

func TestObject_Decode(t *testing.T) {
	t.Run("target is empty interface", func(t *testing.T) {
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

		value.chain.assert(t, success)
		assert.Equal(t, target, m)
	})

	t.Run("target is map", func(t *testing.T) {
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

		value.chain.assert(t, success)
		assert.Equal(t, target, m)
	})

	t.Run("target is struct", func(t *testing.T) {
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

		value.chain.assert(t, success)
		assert.Equal(t, target, actualStruct)
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

		value.chain.assert(t, failure)
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

		value.chain.assert(t, failure)
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
	value.chain.assert(t, success)
	value.chain.clear()

	assert.Equal(t, m, value.Path("$").Raw())
	value.chain.assert(t, success)
	value.chain.clear()

	value.Schema(`{"type": "object"}`)
	value.chain.assert(t, success)
	value.chain.clear()

	value.Schema(`{"type": "array"}`)
	value.chain.assert(t, failure)
	value.chain.clear()

	value.Keys().ContainsOnly(keys...)
	value.chain.assert(t, success)
	value.chain.clear()

	value.Values().ContainsOnly(values...)
	value.chain.assert(t, success)
	value.chain.clear()

	assert.Equal(t, m["foo"], value.Value("foo").Raw())
	value.chain.assert(t, success)
	value.chain.clear()

	assert.Equal(t, m["bar"], value.Value("bar").Raw())
	value.chain.assert(t, success)
	value.chain.clear()

	assert.Equal(t, m["baz"], value.Value("baz").Raw())
	value.chain.assert(t, success)
	value.chain.clear()

	assert.Equal(t, nil, value.Value("BAZ").Raw())
	value.chain.assert(t, failure)
	value.chain.clear()

	it := value.Iter()
	assert.Equal(t, 3, len(it))
	assert.Equal(t, it["foo"].value, value.Value("foo").Raw())
	assert.Equal(t, it["bar"].value, value.Value("bar").Raw())
	it["foo"].chain.assert(t, success)
	it["bar"].chain.assert(t, success)
	value.chain.assert(t, success)
	value.chain.clear()
}

func TestObject_IsEmpty(t *testing.T) {
	t.Run("empty map", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewObject(reporter, map[string]interface{}{})

		value.IsEmpty()
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotEmpty()
		value.chain.assert(t, failure)
		value.chain.clear()
	})

	t.Run("one empty element", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewObject(reporter, map[string]interface{}{"": nil})

		value.IsEmpty()
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEmpty()
		value.chain.assert(t, success)
		value.chain.clear()
	})
}

func TestObject_IsEqual(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewObject(reporter, map[string]interface{}{})

		assert.Equal(t, map[string]interface{}{}, value.Raw())

		value.IsEqual(map[string]interface{}{})
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotEqual(map[string]interface{}{})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.IsEqual(map[string]interface{}{"": nil})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqual(map[string]interface{}{"": nil})
		value.chain.assert(t, success)
		value.chain.clear()
	})

	t.Run("not empty", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewObject(reporter, map[string]interface{}{"foo": 123.0})

		assert.Equal(t, map[string]interface{}{"foo": 123.0}, value.Raw())

		value.IsEqual(map[string]interface{}{})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqual(map[string]interface{}{})
		value.chain.assert(t, success)
		value.chain.clear()

		value.IsEqual(map[string]interface{}{"FOO": 123.0})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqual(map[string]interface{}{"FOO": 123.0})
		value.chain.assert(t, success)
		value.chain.clear()

		value.IsEqual(map[string]interface{}{"foo": 456.0})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqual(map[string]interface{}{"foo": 456.0})
		value.chain.assert(t, success)
		value.chain.clear()

		value.IsEqual(map[string]interface{}{"foo": 123.0})
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotEqual(map[string]interface{}{"foo": 123.0})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.IsEqual(nil)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqual(nil)
		value.chain.assert(t, failure)
		value.chain.clear()
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
				Foo float64 `json:"foo"`
				Bar Bar     `json:"bar"`
			}
		)

		s := S{
			Foo: 123.0,
			Bar: Bar{
				Baz: []bool{true, false},
			},
		}

		value.IsEqual(s)
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotEqual(s)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.IsEqual(S{})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqual(S{})
		value.chain.assert(t, success)
		value.chain.clear()
	})

	t.Run("canonization", func(t *testing.T) {
		type (
			myMap map[string]interface{}
			myInt int
		)

		reporter := newMockReporter(t)

		value := NewObject(reporter, map[string]interface{}{"foo": 123})

		value.IsEqual(map[string]interface{}{"foo": 123.0})
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotEqual(map[string]interface{}{"foo": 123.0})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.IsEqual(myMap{"foo": myInt(123)})
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotEqual(myMap{"foo": myInt(123)})
		value.chain.assert(t, failure)
		value.chain.clear()
	})
}

func TestObject_InList(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewObject(reporter, map[string]interface{}{"foo": 123.0})

		value.InList(
			map[string]interface{}{"FOO": 123.0},
			map[string]interface{}{"BAR": 456.0},
		)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotInList(
			map[string]interface{}{"FOO": 123.0},
			map[string]interface{}{"BAR": 456.0},
		)
		value.chain.assert(t, success)
		value.chain.clear()

		value.InList(
			map[string]interface{}{"foo": 456.0},
			map[string]interface{}{"bar": 123.0},
		)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotInList(
			map[string]interface{}{"foo": 456.0},
			map[string]interface{}{"bar": 123.0},
		)
		value.chain.assert(t, success)
		value.chain.clear()

		value.InList(
			map[string]interface{}{"foo": 123.0},
			map[string]interface{}{"bar": 456.0},
		)
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotInList(
			map[string]interface{}{"foo": 123.0},
			map[string]interface{}{"bar": 456.0},
		)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.InList(struct {
			Foo float64 `json:"foo"`
		}{Foo: 123.00})
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotInList(struct {
			Foo float64 `json:"foo"`
		}{Foo: 123.00})
		value.chain.assert(t, failure)
		value.chain.clear()
	})

	t.Run("empty", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewObject(reporter, map[string]interface{}{})

		value.InList(map[string]interface{}{})
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotInList(map[string]interface{}{})
		value.chain.assert(t, failure)
		value.chain.clear()
	})

	t.Run("not object", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewObject(reporter, map[string]interface{}{"foo": 123.0})

		value.InList(map[string]interface{}{"bar": 123.0}, "NOT OBJECT")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotInList(map[string]interface{}{"bar": 123.0}, "NOT OBJECT")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.InList(map[string]interface{}{"foo": 123.0}, "NOT OBJECT")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotInList(map[string]interface{}{"foo": 123.0}, "NOT OBJECT")
		value.chain.assert(t, failure)
		value.chain.clear()
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewObject(reporter, map[string]interface{}{})

		value.InList()
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotInList()
		value.chain.assert(t, failure)
		value.chain.clear()

		value.InList(nil)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotInList(nil)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.InList(func() {})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotInList(func() {})
		value.chain.assert(t, failure)
		value.chain.clear()
	})

	t.Run("canonization", func(t *testing.T) {
		type (
			myMap map[string]interface{}
			myInt int
		)

		reporter := newMockReporter(t)

		value := NewObject(reporter, map[string]interface{}{"foo": 123, "bar": 456})

		value.InList(myMap{"foo": 123.0, "bar": 456.0})
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotInList(myMap{"foo": 123.0, "bar": 456.0})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.InList(myMap{"foo": "123", "bar": myInt(456.0)})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotInList(myMap{"foo": "123", "bar": myInt(456.0)})
		value.chain.assert(t, success)
		value.chain.clear()
	})
}

func TestObject_ContainsKey(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewObject(reporter, map[string]interface{}{"foo": 123, "bar": ""})

	value.ContainsKey("foo")
	value.chain.assert(t, success)
	value.chain.clear()

	value.NotContainsKey("foo")
	value.chain.assert(t, failure)
	value.chain.clear()

	value.ContainsKey("bar")
	value.chain.assert(t, success)
	value.chain.clear()

	value.NotContainsKey("bar")
	value.chain.assert(t, failure)
	value.chain.clear()

	value.ContainsKey("BAR")
	value.chain.assert(t, failure)
	value.chain.clear()

	value.NotContainsKey("BAR")
	value.chain.assert(t, success)
	value.chain.clear()
}

func TestObject_ContainsValue(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewObject(reporter, map[string]interface{}{"foo": 123, "bar": "xxx"})

		value.ContainsValue(123)
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsValue(123)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.ContainsValue("xxx")
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsValue("xxx")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.ContainsValue("XXX")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotContainsValue("XXX")
		value.chain.assert(t, success)
		value.chain.clear()
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
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsValue(barValue)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.ContainsValue(bazValue)
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsValue(bazValue)
		value.chain.assert(t, failure)
		value.chain.clear()
	})

	t.Run("canonization", func(t *testing.T) {
		type (
			myInt int
		)

		reporter := newMockReporter(t)

		value := NewObject(reporter, map[string]interface{}{"foo": 123, "bar": 789})

		value.ContainsValue(myInt(789.0))
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsValue(myInt(789.0))
		value.chain.assert(t, failure)
		value.chain.clear()
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewObject(reporter, map[string]interface{}{"foo": 123, "bar": "xxx"})

		value.ContainsValue(make(chan int))
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotContainsValue(make(chan int))
		value.chain.assert(t, failure)
		value.chain.clear()
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
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsSubset(submap1)
		value.chain.assert(t, failure)
		value.chain.clear()

		submap2 := map[string]interface{}{
			"bar": []interface{}{"456", 789},
			"baz": map[string]interface{}{
				"a": map[string]interface{}{
					"c": 444,
				},
			},
		}

		value.ContainsSubset(submap2)
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsSubset(submap2)
		value.chain.assert(t, failure)
		value.chain.clear()
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
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotContainsSubset(submap1)
		value.chain.assert(t, success)
		value.chain.clear()

		submap2 := map[string]interface{}{
			"foo": 123,
			"bar": []interface{}{"456", "789"},
		}

		value.ContainsSubset(submap2)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotContainsSubset(submap2)
		value.chain.assert(t, success)
		value.chain.clear()

		submap3 := map[string]interface{}{
			"baz": map[string]interface{}{
				"a": map[string]interface{}{
					"b": "333",
					"c": 444,
				},
			},
		}

		value.ContainsSubset(submap3)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotContainsSubset(submap3)
		value.chain.assert(t, success)
		value.chain.clear()

		value.ContainsSubset(nil)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotContainsSubset(nil)
		value.chain.assert(t, failure)
		value.chain.clear()
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
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsSubset(submap)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.ContainsSubset(S{})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotContainsSubset(S{})
		value.chain.assert(t, success)
		value.chain.clear()
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
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsSubset(submap)
		value.chain.assert(t, failure)
		value.chain.clear()
	})
}

func TestObject_HasValue(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewObject(reporter, map[string]interface{}{
			"foo": 123,
			"bar": []interface{}{"456", 789},
			"baz": map[string]interface{}{
				"a": "b",
			},
		})

		value.HasValue("foo", 123)
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotHasValue("foo", 123)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.HasValue("bar", []interface{}{"456", 789})
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotHasValue("bar", []interface{}{"456", 789})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.HasValue("baz", map[string]interface{}{"a": "b"})
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotHasValue("baz", map[string]interface{}{"a": "b"})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.HasValue("baz", func() {})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotHasValue("baz", func() {})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.HasValue("BAZ", 777)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotHasValue("BAZ", 777)
		value.chain.assert(t, failure)
		value.chain.clear()
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

		value.HasValue("baz", baz)
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotHasValue("baz", baz)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.HasValue("baz", Baz{})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotHasValue("baz", Baz{})
		value.chain.assert(t, success)
		value.chain.clear()
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

		value.HasValue("bar", myArray{"456", myInt(789)})
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotHasValue("bar", myArray{"456", myInt(789)})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.HasValue("baz", myMap{"a": "b"})
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotHasValue("baz", myMap{"a": "b"})
		value.chain.assert(t, failure)
		value.chain.clear()
	})
}

func TestObject_Every(t *testing.T) {
	t.Run("check value", func(t *testing.T) {
		reporter := newMockReporter(t)
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
		object.chain.assert(t, success)
	})

	t.Run("check key", func(t *testing.T) {
		reporter := newMockReporter(t)
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
					assert.Equal(t, "foo", key)
				case "456":
					assert.Equal(t, "bar", key)
				case "baz":
					assert.Equal(t, "baz", key)
				}
			}
		})

		assert.Equal(t, 3, invoked)
		object.chain.assert(t, success)
	})

	t.Run("empty object", func(t *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{})

		invoked := 0
		object.Every(func(_ string, value *Value) {
			invoked++
			value.String().NotEmpty()
		})

		assert.Equal(t, 0, invoked)
		object.chain.assert(t, success)
	})

	t.Run("one assertion fails", func(t *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{"foo": "", "bar": "bar"})

		invoked := 0
		object.Every(func(_ string, val *Value) {
			invoked++
			val.String().NotEmpty()
		})

		assert.Equal(t, 2, invoked)
		object.chain.assert(t, failure)
	})

	t.Run("all assertions fail", func(t *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{"foo": "", "bar": ""})

		invoked := 0
		object.Every(func(_ string, val *Value) {
			invoked++
			val.String().NotEmpty()
		})

		assert.Equal(t, 2, invoked)
		object.chain.assert(t, failure)
	})

	t.Run("call order", func(t *testing.T) {
		reporter := newMockReporter(t)
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

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{})
		object.Every((func(key string, value *Value))(nil))
		object.chain.assert(t, failure)
	})
}

func TestObject_Transform(t *testing.T) {
	t.Run("check index", func(t *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo": "123",
			"bar": "456",
			"baz": "baz",
		})

		newObject := object.Transform(func(key string, value interface{}) interface{} {
			if v, ok := value.(string); ok {
				switch v {
				case "123":
					assert.Equal(t, "foo", key)
				case "456":
					assert.Equal(t, "bar", key)
				case "baz":
					assert.Equal(t, "baz", key)
				}
			}
			return value
		})

		newObject.chain.assert(t, success)
	})

	t.Run("transform value", func(t *testing.T) {
		reporter := newMockReporter(t)
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

	t.Run("empty object", func(t *testing.T) {
		reporter := newMockReporter(t)
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

		newObject.chain.assert(t, success)
	})

	t.Run("call order", func(t *testing.T) {
		reporter := newMockReporter(t)
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

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo": "123",
			"bar": "456",
			"baz": "b",
		})

		newObject := object.Transform(nil)

		newObject.chain.assert(t, failure)
	})

	t.Run("canonization", func(t *testing.T) {
		type (
			myInt int
		)

		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo": "123",
			"bar": "456",
			"baz": "b",
		})

		newObject := object.Transform(func(_ string, val interface{}) interface{} {
			if v, err := strconv.ParseFloat(val.(string), 64); err == nil {
				return myInt(v)
			} else {
				return val
			}
		})

		assert.Equal(t,
			map[string]interface{}{
				"foo": 123.0,
				"bar": 456.0,
				"baz": "b",
			},
			newObject.Raw())
		newObject.chain.assert(t, success)
	})
}

func TestObject_Filter(t *testing.T) {
	t.Run("elements of same type", func(t *testing.T) {
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

		filteredObject.chain.assert(t, success)
		object.chain.assert(t, success)
	})

	t.Run("elements of different types", func(t *testing.T) {
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

		filteredObject.chain.assert(t, success)
		object.chain.assert(t, success)
	})

	t.Run("empty object", func(t *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{})

		filteredObject := object.Filter(func(key string, value *Value) bool {
			return false
		})
		assert.Equal(t, map[string]interface{}{}, filteredObject.Raw())

		filteredObject.chain.assert(t, success)
		object.chain.assert(t, success)
	})

	t.Run("no match", func(t *testing.T) {
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

		filteredObject.chain.assert(t, success)
		object.chain.assert(t, success)
	})

	t.Run("assertion fails", func(t *testing.T) {
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

		filteredObject.chain.assert(t, success)
		object.chain.assert(t, success)
	})

	t.Run("call order", func(t *testing.T) {
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

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{})
		filteredObject := object.Filter((func(key string, value *Value) bool)(nil))
		object.chain.assert(t, failure)
		filteredObject.chain.assert(t, failure)
	})
}

func TestObject_Find(t *testing.T) {
	t.Run("elements of same type", func(t *testing.T) {
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

		foundValue.chain.assert(t, success)
		object.chain.assert(t, success)
	})

	t.Run("elements of different types", func(t *testing.T) {
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

		foundValue.chain.assert(t, success)
		object.chain.assert(t, success)
	})

	t.Run("no match", func(t *testing.T) {
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

		foundValue.chain.assert(t, failure)
		object.chain.assert(t, failure)
	})

	t.Run("empty object", func(t *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{})

		foundValue := object.Find(func(key string, value *Value) bool {
			n := value.Number().Raw()
			return n == 3
		})

		assert.Equal(t, nil, foundValue.Raw())
		assert.Equal(t, object.Raw(), map[string]interface{}{})

		foundValue.chain.assert(t, failure)
		object.chain.assert(t, failure)
	})

	t.Run("predicate returns true, assertion fails, no match", func(t *testing.T) {
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

		foundValue.chain.assert(t, failure)
		object.chain.assert(t, failure)
	})

	t.Run("predicate returns true, assertion fails, have match", func(t *testing.T) {
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

		foundValue.chain.assert(t, success)
		object.chain.assert(t, success)
	})

	t.Run("invalid argument", func(t *testing.T) {
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

		foundValue.chain.assert(t, failure)
		object.chain.assert(t, failure)
	})
}

func TestObject_FindAll(t *testing.T) {
	t.Run("elements of same type", func(t *testing.T) {
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
			value.chain.assert(t, success)
		}
		object.chain.assert(t, success)
	})

	t.Run("elements of different types", func(t *testing.T) {
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
			value.chain.assert(t, success)
		}
		object.chain.assert(t, success)
	})

	t.Run("no match", func(t *testing.T) {
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
			value.chain.assert(t, success)
		}
		object.chain.assert(t, success)
	})

	t.Run("empty object", func(t *testing.T) {
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
			value.chain.assert(t, success)
		}
		object.chain.assert(t, success)
	})

	t.Run("predicate returns true, assertion fails, no match", func(t *testing.T) {
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
			value.chain.assert(t, success)
		}
		object.chain.assert(t, success)
	})

	t.Run("predicate returns true, assertion fails, have matches", func(t *testing.T) {
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
			value.chain.assert(t, success)
		}
		object.chain.assert(t, success)
	})

	t.Run("invalid argument", func(t *testing.T) {
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
			value.chain.assert(t, failure)
		}
		object.chain.assert(t, failure)
	})
}

func TestObject_NotFind(t *testing.T) {
	t.Run("no match", func(t *testing.T) {
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
		object.chain.assert(t, success)
	})

	t.Run("have match", func(t *testing.T) {
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
		object.chain.assert(t, failure)
	})

	t.Run("empty object", func(t *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{})

		afterObject := object.NotFind(func(key string, value *Value) bool {
			return key == "corge"
		})

		assert.Same(t, object, afterObject)
		object.chain.assert(t, success)
	})

	t.Run("predicate returns true, assertion fails, no match", func(t *testing.T) {
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
		object.chain.assert(t, success)
	})

	t.Run("predicate returns true, assertion fails, have match", func(t *testing.T) {
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
		object.chain.assert(t, failure)
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)
		object := NewObject(reporter, map[string]interface{}{
			"foo": 1,
			"bar": 2,
		})

		afterObject := object.NotFind(nil)

		assert.Same(t, object, afterObject)
		object.chain.assert(t, failure)
	})
}
