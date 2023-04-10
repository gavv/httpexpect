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

		NewObject(reporter, map[string]interface{}{}).IsEmpty().
			chain.assert(t, success)

		NewObject(reporter, map[string]interface{}{}).NotEmpty().
			chain.assert(t, failure)
	})

	t.Run("one empty element", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewObject(reporter, map[string]interface{}{"": nil}).IsEmpty().
			chain.assert(t, failure)

		NewObject(reporter, map[string]interface{}{"": nil}).NotEmpty().
			chain.assert(t, success)
	})
}

func TestObject_IsEqual(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewObject(reporter, map[string]interface{}{}).IsEqual(map[string]interface{}{}).
			chain.assert(t, success)

		NewObject(reporter, map[string]interface{}{}).NotEqual(map[string]interface{}{}).
			chain.assert(t, failure)

		NewObject(reporter, map[string]interface{}{}).IsEqual(map[string]interface{}{"": nil}).
			chain.assert(t, failure)

		NewObject(reporter, map[string]interface{}{}).NotEqual(map[string]interface{}{"": nil}).
			chain.assert(t, success)
	})

	t.Run("basic", func(t *testing.T) {
		cases := []struct {
			name      string
			value     map[string]interface{}
			testValue map[string]interface{}
			wantEqual chainResult
		}{
			{
				name:      "compare empty object with empty object",
				value:     map[string]interface{}{},
				testValue: map[string]interface{}{},
				wantEqual: success,
			},
			{
				name:      "compare empty object with non-empty object",
				value:     map[string]interface{}{"": nil},
				testValue: map[string]interface{}{},
				wantEqual: failure,
			},
			{
				name:      "with empty object",
				value:     map[string]interface{}{"foo": 123.0},
				testValue: map[string]interface{}{},
				wantEqual: failure,
			},
			{
				name:      "with (FOO: 123.0)",
				value:     map[string]interface{}{"foo": 123.0},
				testValue: map[string]interface{}{"FOO": 123.0},
				wantEqual: failure,
			},
			{
				name:      "with (foo: 456.0)",
				value:     map[string]interface{}{"foo": 123.0},
				testValue: map[string]interface{}{"foo": 456.0},
				wantEqual: failure,
			},
			{
				name:      "with (foo: 123.0)",
				value:     map[string]interface{}{"foo": 123.0},
				testValue: map[string]interface{}{"foo": 123.0},
				wantEqual: success,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewObject(reporter, tc.value).IsEqual(tc.testValue).
					chain.assert(t, tc.wantEqual)

				NewObject(reporter, tc.value).NotEqual(tc.testValue).
					chain.assert(t, !tc.wantEqual)
			})
		}
	})

	t.Run("struct", func(t *testing.T) {
		reporter := newMockReporter(t)

		type (
			Bar struct {
				Baz []bool `json:"baz"`
			}

			S struct {
				Foo int `json:"foo"`
				Bar Bar `json:"bar"`
			}
		)

		value := map[string]interface{}{
			"foo": 123,
			"bar": map[string]interface{}{
				"baz": []interface{}{true, false},
			},
		}

		s := S{
			Foo: 123,
			Bar: Bar{
				Baz: []bool{true, false},
			},
		}

		NewObject(reporter, value).IsEqual(s).
			chain.assert(t, success)

		NewObject(reporter, value).NotEqual(s).
			chain.assert(t, failure)

		NewObject(reporter, value).IsEqual(S{}).
			chain.assert(t, failure)

		NewObject(reporter, value).NotEqual(S{}).
			chain.assert(t, success)
	})

	t.Run("canonization", func(t *testing.T) {
		reporter := newMockReporter(t)

		type (
			myMap map[string]interface{}
			myInt int
		)

		NewObject(reporter, map[string]interface{}{"foo": 123}).
			IsEqual(map[string]interface{}{"foo": 123}).
			chain.assert(t, success)

		NewObject(reporter, map[string]interface{}{"foo": 123}).
			NotEqual(map[string]interface{}{"foo": 123}).
			chain.assert(t, failure)

		NewObject(reporter, map[string]interface{}{"foo": 123}).
			IsEqual(myMap{"foo": myInt(123)}).
			chain.assert(t, success)

		NewObject(reporter, map[string]interface{}{"foo": 123}).
			NotEqual(myMap{"foo": myInt(123)}).
			chain.assert(t, failure)
	})
}

func TestObject_InList(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		cases := []struct {
			name      string
			value     map[string]interface{}
			testValue []interface{}
			wantEqual chainResult
		}{
			{
				name:      "with empty object",
				value:     map[string]interface{}{},
				testValue: []interface{}{map[string]interface{}{}},
				wantEqual: success,
			},
			{
				name:  "with (FOO: 123.0, BAR: 456.0)",
				value: map[string]interface{}{"foo": 123.0},
				testValue: []interface{}{
					map[string]interface{}{"FOO": 123.0},
					map[string]interface{}{"BAR": 456.0},
				},
				wantEqual: failure,
			},
			{
				name:  "with (foo: 456.0, bar: 123.0)",
				value: map[string]interface{}{"foo": 123.0},
				testValue: []interface{}{
					map[string]interface{}{"foo": 456.0},
					map[string]interface{}{"bar": 123.0},
				},
				wantEqual: failure,
			},
			{
				name:  "with (foo: 123.0, bar: 456.0)",
				value: map[string]interface{}{"foo": 123.0},
				testValue: []interface{}{
					map[string]interface{}{"foo": 123.0},
					map[string]interface{}{"bar": 456.0},
				},
				wantEqual: success,
			},
			{
				name:  "with (foo: 123.0)",
				value: map[string]interface{}{"foo": 123.0},
				testValue: []interface{}{struct {
					Foo float64 `json:"foo"`
				}{Foo: 123.00}},
				wantEqual: success,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewObject(reporter, tc.value).InList(tc.testValue...).
					chain.assert(t, tc.wantEqual)

				NewObject(reporter, tc.value).NotInList(tc.testValue...).
					chain.assert(t, !tc.wantEqual)
			})
		}
	})

	t.Run("invalid argument", func(t *testing.T) {
		cases := []struct {
			name     string
			value    map[string]interface{}
			testList []interface{}
		}{
			{
				name:     "empty list",
				value:    map[string]interface{}{},
				testList: []interface{}{},
			},
			{
				name:     "nil list",
				value:    map[string]interface{}{},
				testList: nil,
			},
			{
				name:     "invalid type",
				value:    map[string]interface{}{},
				testList: []interface{}{func() {}},
			},
			{
				name:  "one inequal object, another not object",
				value: map[string]interface{}{"foo": 123.0},
				testList: []interface{}{
					map[string]interface{}{"bar": 123.0},
					"NOT OBJECT",
				},
			},
			{
				name:  "one equal object, another not object",
				value: map[string]interface{}{"foo": 123.0},
				testList: []interface{}{
					map[string]interface{}{"foo": 123.0},
					"NOT OBJECT",
				},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewObject(reporter, tc.value).InList(tc.testList...).
					chain.assert(t, failure)

				NewObject(reporter, tc.value).NotInList(tc.testList...).
					chain.assert(t, failure)

			})
		}
	})

	t.Run("canonization", func(t *testing.T) {
		type (
			myMap map[string]interface{}
			myInt int
		)
		reporter := newMockReporter(t)

		NewObject(reporter, map[string]interface{}{"foo": 123, "bar": 456}).
			InList(myMap{"foo": 123.0, "bar": 456.0}).
			chain.assert(t, success)

		NewObject(reporter, map[string]interface{}{"foo": 123, "bar": 456}).
			NotInList(myMap{"foo": 123.0, "bar": 456.0}).
			chain.assert(t, failure)

		NewObject(reporter, map[string]interface{}{"foo": 123, "bar": 456}).
			InList(myMap{"foo": "123", "bar": myInt(456.0)}).
			chain.assert(t, failure)

		NewObject(reporter, map[string]interface{}{"foo": 123, "bar": 456}).
			NotInList(myMap{"foo": "123", "bar": myInt(456.0)}).
			chain.assert(t, success)
	})
}

func TestObject_ContainsKey(t *testing.T) {
	testObj := map[string]interface{}{"foo": 123, "bar": ""}

	cases := []struct {
		name               string
		object             map[string]interface{}
		key                string
		wantContainsKey    chainResult
		wantNotContainsKey chainResult
	}{
		{
			name:               "foo value, correct key value",
			object:             testObj,
			key:                "foo",
			wantContainsKey:    success,
			wantNotContainsKey: failure,
		},
		{
			name:               "bar value, correct key value",
			object:             testObj,
			key:                "bar",
			wantContainsKey:    success,
			wantNotContainsKey: failure,
		},
		{
			name:               "BAR value, wrong key value",
			object:             testObj,
			key:                "BAR",
			wantContainsKey:    failure,
			wantNotContainsKey: success,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewObject(reporter, tc.object).ContainsKey(tc.key).
				chain.assert(t, tc.wantContainsKey)

			NewObject(reporter, tc.object).NotContainsKey(tc.key).
				chain.assert(t, tc.wantNotContainsKey)
		})
	}
}

func TestObject_ContainsValue(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		testObj := map[string]interface{}{"foo": 123, "bar": "xxx"}

		cases := []struct {
			name               string
			object             map[string]interface{}
			value              interface{}
			wantContainsKey    chainResult
			wantNotContainsKey chainResult
		}{
			{
				name:               "123 value, correct value",
				object:             testObj,
				value:              123,
				wantContainsKey:    success,
				wantNotContainsKey: failure,
			},
			{
				name:               "xxx value, correct value",
				object:             testObj,
				value:              "xxx",
				wantContainsKey:    success,
				wantNotContainsKey: failure,
			},
			{
				name:               "XXX value, wrong value",
				object:             testObj,
				value:              "XXX",
				wantContainsKey:    failure,
				wantNotContainsKey: success,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewObject(reporter, tc.object).ContainsValue(tc.value).
					chain.assert(t, tc.wantContainsKey)

				NewObject(reporter, tc.object).NotContainsValue(tc.value).
					chain.assert(t, tc.wantNotContainsKey)
			})
		}
	})

	t.Run("struct", func(t *testing.T) {
		testObj := map[string]interface{}{
			"foo": 123,
			"bar": []interface{}{"456", 789},
			"baz": map[string]interface{}{
				"a": map[string]interface{}{
					"b": 333,
					"c": 444,
				},
			},
		}

		type (
			A struct {
				B int `json:"b"`
				C int `json:"c"`
			}

			Baz struct {
				A A `json:"a"`
			}
		)

		cases := []struct {
			name               string
			object             map[string]interface{}
			value              interface{}
			wantContainsKey    chainResult
			wantNotContainsKey chainResult
		}{
			{
				name:               "correct value, contains slice",
				object:             testObj,
				value:              []interface{}{"456", 789},
				wantContainsKey:    success,
				wantNotContainsKey: failure,
			},
			{
				name:   "correct value, contains nested map",
				object: testObj,
				value: Baz{
					A: A{
						B: 333,
						C: 444,
					},
				},
				wantContainsKey:    success,
				wantNotContainsKey: failure,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewObject(reporter, tc.object).ContainsValue(tc.value).
					chain.assert(t, tc.wantContainsKey)

				NewObject(reporter, tc.object).NotContainsValue(tc.value).
					chain.assert(t, tc.wantNotContainsKey)
			})
		}
	})

	t.Run("canonization", func(t *testing.T) {
		testObj := map[string]interface{}{"foo": 123, "bar": 789}

		type (
			myInt int
		)

		cases := []struct {
			name               string
			object             map[string]interface{}
			value              interface{}
			wantContainsKey    chainResult
			wantNotContainsKey chainResult
		}{
			{
				name:               "correct value, wrapped primitive",
				object:             testObj,
				value:              myInt(789.0),
				wantContainsKey:    success,
				wantNotContainsKey: failure,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewObject(reporter, tc.object).ContainsValue(tc.value).
					chain.assert(t, tc.wantContainsKey)

				NewObject(reporter, tc.object).NotContainsValue(tc.value).
					chain.assert(t, tc.wantNotContainsKey)
			})
		}
	})

	t.Run("invalid argument", func(t *testing.T) {
		testObj := map[string]interface{}{"foo": 123, "bar": "xxx"}

		cases := []struct {
			name               string
			object             map[string]interface{}
			value              interface{}
			wantContainsKey    chainResult
			wantNotContainsKey chainResult
		}{
			{
				name:               "invalid value, channel",
				object:             testObj,
				value:              make(chan int),
				wantContainsKey:    failure,
				wantNotContainsKey: failure,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewObject(reporter, tc.object).ContainsValue(tc.value).
					chain.assert(t, tc.wantContainsKey)

				NewObject(reporter, tc.object).NotContainsValue(tc.value).
					chain.assert(t, tc.wantNotContainsKey)
			})
		}
	})
}

func TestObject_ContainsSubset(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		testObj := map[string]interface{}{
			"foo": 123,
			"bar": []interface{}{"456", 789},
			"baz": map[string]interface{}{
				"a": map[string]interface{}{
					"b": 333,
					"c": 444,
				},
			},
		}

		cases := []struct {
			name               string
			object             map[string]interface{}
			subset             map[string]interface{}
			wantContainsKey    chainResult
			wantNotContainsKey chainResult
		}{
			{
				name:   "partial subset with slices",
				object: testObj,
				subset: map[string]interface{}{
					"foo": 123,
					"bar": []interface{}{"456", 789},
				},
				wantContainsKey:    success,
				wantNotContainsKey: failure,
			},
			{
				name:   "partial subset with nested maps",
				object: testObj,
				subset: map[string]interface{}{
					"bar": []interface{}{"456", 789},
					"baz": map[string]interface{}{
						"a": map[string]interface{}{
							"c": 444,
						},
					},
				},
				wantContainsKey:    success,
				wantNotContainsKey: failure,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewObject(reporter, tc.object).ContainsSubset(tc.subset).
					chain.assert(t, tc.wantContainsKey)

				NewObject(reporter, tc.object).NotContainsSubset(tc.subset).
					chain.assert(t, tc.wantNotContainsKey)
			})
		}
	})

	t.Run("failure", func(t *testing.T) {
		testObj := map[string]interface{}{
			"foo": 123,
			"bar": []interface{}{"456", 789},
			"baz": map[string]interface{}{
				"a": map[string]interface{}{
					"b": 333,
					"c": 444,
				},
			},
		}

		cases := []struct {
			name               string
			object             map[string]interface{}
			subset             map[string]interface{}
			assertion          uint
			wantContainsKey    chainResult
			wantNotContainsKey chainResult
		}{
			{
				name:   "partial subset with wrong key",
				object: testObj,
				subset: map[string]interface{}{
					"foo": 123,
					"qux": 456,
				},
				wantContainsKey:    failure,
				wantNotContainsKey: success,
			},
			{
				name:   "partial subset with nested map",
				object: testObj,
				subset: map[string]interface{}{
					"baz": map[string]interface{}{
						"a": map[string]interface{}{
							"b": "333",
							"c": 444,
						},
					},
				},
				wantContainsKey:    failure,
				wantNotContainsKey: success,
			},
			{
				name:               "nil subset",
				object:             testObj,
				subset:             nil,
				wantContainsKey:    failure,
				wantNotContainsKey: failure,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewObject(reporter, tc.object).ContainsSubset(tc.subset).
					chain.assert(t, tc.wantContainsKey)

				NewObject(reporter, tc.object).NotContainsSubset(tc.subset).
					chain.assert(t, tc.wantNotContainsKey)
			})
		}
	})

	t.Run("struct", func(t *testing.T) {
		testObj := map[string]interface{}{
			"foo": 123,
			"bar": []interface{}{"456", 789},
			"baz": map[string]interface{}{
				"a": map[string]interface{}{
					"b": 333,
					"c": 444,
				},
			},
		}

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

		cases := []struct {
			name               string
			object             map[string]interface{}
			value              interface{}
			wantContainsKey    chainResult
			wantNotContainsKey chainResult
		}{
			{
				name:   "partial subset",
				object: testObj,
				value: S{
					Foo: 123,
					Baz: Baz{
						A: A{
							B: 333,
						},
					},
				},
				wantContainsKey:    success,
				wantNotContainsKey: failure,
			},
			{
				name:               "empty subset",
				object:             testObj,
				value:              S{},
				wantContainsKey:    failure,
				wantNotContainsKey: success,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewObject(reporter, tc.object).ContainsSubset(tc.value).
					chain.assert(t, tc.wantContainsKey)

				NewObject(reporter, tc.object).NotContainsSubset(tc.value).
					chain.assert(t, tc.wantNotContainsKey)
			})
		}
	})

	t.Run("canonization", func(t *testing.T) {
		testObj := map[string]interface{}{
			"foo": 123,
			"bar": []interface{}{"456", 789},
			"baz": map[string]interface{}{
				"a": map[string]interface{}{
					"b": 333,
					"c": 444,
				},
			},
		}

		type (
			myArray []interface{}
			myMap   map[string]interface{}
			myInt   int
		)

		cases := []struct {
			name               string
			object             map[string]interface{}
			value              interface{}
			wantContainsKey    chainResult
			wantNotContainsKey chainResult
		}{
			{
				name:   "correct value, wrapped map",
				object: testObj,
				value: myMap{
					"foo": myInt(123),
					"bar": myArray{"456", myInt(789)},
				},
				wantContainsKey:    success,
				wantNotContainsKey: failure,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewObject(reporter, tc.object).ContainsSubset(tc.value).
					chain.assert(t, tc.wantContainsKey)

				NewObject(reporter, tc.object).NotContainsSubset(tc.value).
					chain.assert(t, tc.wantNotContainsKey)
			})
		}
	})
}

func TestObject_HasValue(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		testObj := map[string]interface{}{
			"foo": 123,
			"bar": []interface{}{"456", 789},
			"baz": map[string]interface{}{
				"a": map[string]interface{}{
					"b": 333,
					"c": 444,
				},
			},
		}

		cases := []struct {
			name               string
			object             map[string]interface{}
			key                string
			value              interface{}
			assertion          uint
			wantContainsKey    chainResult
			wantNotContainsKey chainResult
		}{
			{
				name:               "correct key-value with primitives",
				object:             testObj,
				key:                "foo",
				value:              123,
				wantContainsKey:    success,
				wantNotContainsKey: failure,
			},
			{
				name:               "correct key-value with slices",
				object:             testObj,
				key:                "bar",
				value:              []interface{}{"456", 789},
				wantContainsKey:    success,
				wantNotContainsKey: failure,
			},
			{
				name:               "wrong key-value with maps",
				object:             testObj,
				key:                "baz",
				value:              map[string]interface{}{"a": "b"},
				wantContainsKey:    failure,
				wantNotContainsKey: success,
			},
			{
				name:               "wrong value with empty func",
				object:             testObj,
				key:                "baz",
				value:              func() {},
				wantContainsKey:    failure,
				wantNotContainsKey: failure,
			},
			{
				name:               "wrong key-value with primitive",
				object:             testObj,
				key:                "BAZ",
				value:              777,
				wantContainsKey:    failure,
				wantNotContainsKey: failure,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewObject(reporter, tc.object).HasValue(tc.key, tc.value).
					chain.assert(t, tc.wantContainsKey)

				NewObject(reporter, tc.object).NotHasValue(tc.key, tc.value).
					chain.assert(t, tc.wantNotContainsKey)
			})
		}
	})

	t.Run("struct", func(t *testing.T) {
		testObj := map[string]interface{}{
			"foo": 123,
			"bar": []interface{}{"456", 789},
			"baz": map[string]interface{}{
				"a": map[string]interface{}{
					"b": 333,
					"c": 444,
				},
			},
		}

		type (
			A struct {
				B int `json:"b"`
				C int `json:"c"`
			}

			Baz struct {
				A A `json:"a"`
			}
		)

		cases := []struct {
			name               string
			object             map[string]interface{}
			key                string
			value              interface{}
			wantContainsKey    chainResult
			wantNotContainsKey chainResult
		}{
			{
				name:   "partial subset",
				object: testObj,
				key:    "baz",
				value: Baz{
					A: A{
						B: 333,
						C: 444,
					},
				},
				wantContainsKey:    success,
				wantNotContainsKey: failure,
			},
			{
				name:               "empty subset",
				object:             testObj,
				key:                "baz",
				value:              Baz{},
				wantContainsKey:    failure,
				wantNotContainsKey: success,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewObject(reporter, tc.object).HasValue(tc.key, tc.value).
					chain.assert(t, tc.wantContainsKey)

				NewObject(reporter, tc.object).NotHasValue(tc.key, tc.value).
					chain.assert(t, tc.wantNotContainsKey)
			})
		}
	})

	t.Run("canonization", func(t *testing.T) {
		type (
			myArray []interface{}
			myMap   map[string]interface{}
			myInt   int
		)

		testObj := map[string]interface{}{
			"foo": 123,
			"bar": []interface{}{"456", 789},
			"baz": map[string]interface{}{
				"a": "b",
			},
		}

		cases := []struct {
			name               string
			object             map[string]interface{}
			key                string
			value              interface{}
			wantContainsKey    chainResult
			wantNotContainsKey chainResult
		}{
			{
				name:               "correct value, wrapped array",
				object:             testObj,
				key:                "bar",
				value:              myArray{"456", myInt(789)},
				wantContainsKey:    success,
				wantNotContainsKey: failure,
			},
			{
				name:               "correct value, wrapped map",
				object:             testObj,
				key:                "baz",
				value:              myMap{"a": "b"},
				wantContainsKey:    success,
				wantNotContainsKey: failure,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewObject(reporter, tc.object).HasValue(tc.key, tc.value).
					chain.assert(t, tc.wantContainsKey)

				NewObject(reporter, tc.object).NotHasValue(tc.key, tc.value).
					chain.assert(t, tc.wantNotContainsKey)
			})
		}
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
