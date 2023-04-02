package httpexpect

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValue_FailedChain(t *testing.T) {
	chain := newMockChain(t, flagFailed)

	value := newValue(chain, nil)
	value.chain.assert(t, failure)

	value.Path("$").chain.assert(t, failure)
	value.Schema("")
	value.Alias("foo")

	var target interface{}
	value.Decode(target)

	value.Object().chain.assert(t, failure)
	value.Array().chain.assert(t, failure)
	value.String().chain.assert(t, failure)
	value.Number().chain.assert(t, failure)
	value.Boolean().chain.assert(t, failure)

	value.IsNull()
	value.NotNull()
	value.IsObject()
	value.NotObject()
	value.IsArray()
	value.NotArray()
	value.IsString()
	value.NotString()
	value.IsNumber()
	value.NotNumber()
	value.IsBoolean()
	value.NotBoolean()
	value.IsEqual(nil)
	value.NotEqual(nil)
	value.InList(nil)
	value.NotInList(nil)
}

func TestValue_Constructors(t *testing.T) {
	t.Run("reporter", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewValue(reporter, "Test")
		value.IsEqual("Test")
		value.chain.assert(t, success)
		value.String().chain.assert(t, success)
	})

	t.Run("config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewValueC(Config{
			Reporter: reporter,
		}, "Test")
		value.IsEqual("Test")
		value.chain.assert(t, success)
		value.String().chain.assert(t, success)
	})

	t.Run("chain", func(t *testing.T) {
		chain := newMockChain(t)
		value := newValue(chain, "Test")
		assert.NotSame(t, value.chain, chain)
		assert.Equal(t, value.chain.context.Path, chain.context.Path)
	})
}

func TestValue_Decode(t *testing.T) {
	t.Run("target is empty interface", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewValue(reporter, 123.0)

		var target interface{}
		value.Decode(&target)

		value.chain.assert(t, success)
		assert.Equal(t, 123.0, target)
	})

	t.Run("target is struct", func(t *testing.T) {
		reporter := newMockReporter(t)

		type S struct {
			Foo int             `json:"foo"`
			Bar []interface{}   `json:"bar"`
			Baz struct{ A int } `json:"baz"`
		}

		m := map[string]interface{}{
			"foo": 123,
			"bar": []interface{}{"123", 456.0},
			"baz": struct{ A int }{123},
		}

		value := NewValue(reporter, m)

		actualStruct := S{
			123,
			[]interface{}{"123", 456.0},
			struct{ A int }{123},
		}

		var target S
		value.Decode(&target)

		value.chain.assert(t, success)
		assert.Equal(t, target, actualStruct)
	})

	t.Run("target is nil", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewValue(reporter, 123)

		value.Decode(nil)

		value.chain.failed()
	})

	t.Run("target is unmarshable", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewValue(reporter, 123)

		value.Decode(123)

		value.chain.failed()
	})
}

func TestValue_Alias(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewValue(reporter, 123)
	assert.Equal(t, []string{"Value()"}, value.chain.context.Path)
	assert.Equal(t, []string{"Value()"}, value.chain.context.AliasedPath)

	value.Alias("foo")
	assert.Equal(t, []string{"Value()"}, value.chain.context.Path)
	assert.Equal(t, []string{"foo"}, value.chain.context.AliasedPath)

	childValue := value.Number()
	assert.Equal(t, []string{"Value()", "Number()"}, childValue.chain.context.Path)
	assert.Equal(t, []string{"foo", "Number()"}, childValue.chain.context.AliasedPath)
}

func TestValue_Getters(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("null", func(t *testing.T) {
		var data interface{}

		NewValue(reporter, data).Object().chain.assert(t, failure)
		NewValue(reporter, data).Array().chain.assert(t, failure)
		NewValue(reporter, data).String().chain.assert(t, failure)
		NewValue(reporter, data).Number().chain.assert(t, failure)
		NewValue(reporter, data).Boolean().chain.assert(t, failure)
		NewValue(reporter, data).NotNull().chain.assert(t, failure)
		NewValue(reporter, data).IsNull().chain.assert(t, success)
	})

	t.Run("indirect null", func(t *testing.T) {
		var data []interface{}

		NewValue(reporter, data).Object().chain.assert(t, failure)
		NewValue(reporter, data).Array().chain.assert(t, failure)
		NewValue(reporter, data).String().chain.assert(t, failure)
		NewValue(reporter, data).Number().chain.assert(t, failure)
		NewValue(reporter, data).Boolean().chain.assert(t, failure)
		NewValue(reporter, data).NotNull().chain.assert(t, failure)
		NewValue(reporter, data).IsNull().chain.assert(t, success)
	})

	t.Run("bad", func(t *testing.T) {
		data := func() {}

		NewValue(reporter, data).Object().chain.assert(t, failure)
		NewValue(reporter, data).Array().chain.assert(t, failure)
		NewValue(reporter, data).String().chain.assert(t, failure)
		NewValue(reporter, data).Number().chain.assert(t, failure)
		NewValue(reporter, data).Boolean().chain.assert(t, failure)
		NewValue(reporter, data).NotNull().chain.assert(t, failure)
		NewValue(reporter, data).IsNull().chain.assert(t, failure)
	})

	t.Run("object", func(t *testing.T) {
		data := map[string]interface{}{}

		NewValue(reporter, data).Object().chain.assert(t, success)
		NewValue(reporter, data).Array().chain.assert(t, failure)
		NewValue(reporter, data).String().chain.assert(t, failure)
		NewValue(reporter, data).Number().chain.assert(t, failure)
		NewValue(reporter, data).Boolean().chain.assert(t, failure)
		NewValue(reporter, data).NotNull().chain.assert(t, success)
		NewValue(reporter, data).IsNull().chain.assert(t, failure)
	})

	t.Run("array", func(t *testing.T) {
		data := []interface{}{}

		NewValue(reporter, data).Object().chain.assert(t, failure)
		NewValue(reporter, data).Array().chain.assert(t, success)
		NewValue(reporter, data).String().chain.assert(t, failure)
		NewValue(reporter, data).Number().chain.assert(t, failure)
		NewValue(reporter, data).Boolean().chain.assert(t, failure)
		NewValue(reporter, data).NotNull().chain.assert(t, success)
		NewValue(reporter, data).IsNull().chain.assert(t, failure)
	})

	t.Run("string", func(t *testing.T) {
		data := ""

		NewValue(reporter, data).Object().chain.assert(t, failure)
		NewValue(reporter, data).Array().chain.assert(t, failure)
		NewValue(reporter, data).String().chain.assert(t, success)
		NewValue(reporter, data).Number().chain.assert(t, failure)
		NewValue(reporter, data).Boolean().chain.assert(t, failure)
		NewValue(reporter, data).NotNull().chain.assert(t, success)
		NewValue(reporter, data).IsNull().chain.assert(t, failure)
	})

	t.Run("number", func(t *testing.T) {
		data := 0.0

		NewValue(reporter, data).Object().chain.assert(t, failure)
		NewValue(reporter, data).Array().chain.assert(t, failure)
		NewValue(reporter, data).String().chain.assert(t, failure)
		NewValue(reporter, data).Number().chain.assert(t, success)
		NewValue(reporter, data).Boolean().chain.assert(t, failure)
		NewValue(reporter, data).NotNull().chain.assert(t, success)
		NewValue(reporter, data).IsNull().chain.assert(t, failure)
	})

	t.Run("boolean", func(t *testing.T) {
		data := false

		NewValue(reporter, data).Object().chain.assert(t, failure)
		NewValue(reporter, data).Array().chain.assert(t, failure)
		NewValue(reporter, data).String().chain.assert(t, failure)
		NewValue(reporter, data).Number().chain.assert(t, failure)
		NewValue(reporter, data).Boolean().chain.assert(t, success)
		NewValue(reporter, data).NotNull().chain.assert(t, success)
		NewValue(reporter, data).IsNull().chain.assert(t, failure)
	})
}

func TestValue_GetObject(t *testing.T) {
	type myMap map[string]interface{}

	cases := []struct {
		name           string
		data           interface{}
		result         chainResult
		expectedObject map[string]interface{}
	}{
		{
			name:           "map",
			data:           map[string]interface{}{"foo": 123.0},
			result:         success,
			expectedObject: map[string]interface{}{"foo": 123.0},
		},
		{
			name:           "myMap",
			data:           myMap{"foo": 123.0},
			result:         success,
			expectedObject: map[string]interface{}(myMap{"foo": 123.0}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			value := NewValue(reporter, tc.data)
			inner := value.Object()

			inner.chain.assert(t, tc.result)

			if tc.result {
				assert.Equal(t, tc.expectedObject, inner.Raw())
			}
		})
	}
}

func TestValue_GetArray(t *testing.T) {
	type myArray []interface{}

	cases := []struct {
		name          string
		data          interface{}
		result        chainResult
		expectedArray []interface{}
	}{
		{
			name:          "array",
			data:          []interface{}{"foo", 123.0},
			result:        success,
			expectedArray: []interface{}{"foo", 123.0},
		},
		{
			name:          "myArray",
			data:          myArray{"foo", 123.0},
			result:        success,
			expectedArray: []interface{}(myArray{"foo", 123.0}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			value := NewValue(reporter, tc.data)
			inner := value.Array()

			value.chain.assert(t, tc.result)
			inner.chain.assert(t, tc.result)

			if tc.result {
				assert.Equal(t, tc.expectedArray, inner.Raw())
			}
		})
	}
}

func TestValue_GetString(t *testing.T) {
	type myString string

	cases := []struct {
		name           string
		data           interface{}
		result         chainResult
		expectedString string
	}{
		{
			name:           "string",
			data:           "foo",
			result:         success,
			expectedString: "foo",
		},
		{
			name:           "myString",
			data:           myString("foo"),
			result:         success,
			expectedString: "foo",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			value := NewValue(reporter, tc.data)
			inner := value.String()

			value.chain.assert(t, tc.result)
			inner.chain.assert(t, tc.result)

			if tc.result {
				assert.Equal(t, tc.expectedString, inner.Raw())
			}
		})
	}
}

func TestValue_GetNumber(t *testing.T) {
	type myInt int

	cases := []struct {
		name        string
		data        interface{}
		result      chainResult
		expectedNum float64
	}{
		{name: "float", data: 123.0, result: success, expectedNum: float64(123.0)},
		{name: "integer", data: 123, result: success, expectedNum: float64(123)},
		{name: "myInt", data: myInt(123), result: success, expectedNum: float64(myInt(123))},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			value := NewValue(reporter, tc.data)
			inner := value.Number()

			value.chain.assert(t, tc.result)
			inner.chain.assert(t, tc.result)

			if tc.result {
				assert.Equal(t, tc.expectedNum, inner.Raw())
			}
		})
	}
}

func TestValue_GetBoolean(t *testing.T) {
	type myBool bool

	cases := []struct {
		name         string
		data         interface{}
		result       chainResult
		expectedBool bool
	}{
		{name: "false", data: false, result: success, expectedBool: false},
		{name: "true", data: true, result: success, expectedBool: true},
		{name: "myTrue", data: myBool(true), result: success, expectedBool: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			value := NewValue(reporter, tc.data)
			inner := value.Boolean()

			value.chain.assert(t, tc.result)
			inner.chain.assert(t, tc.result)

			if tc.result {
				assert.Equal(t, tc.expectedBool, inner.Raw())
			}
		})
	}
}

func TestValue_IsObject(t *testing.T) {
	cases := []struct {
		name       string
		data       interface{}
		wantObject chainResult
	}{
		{name: "object", data: map[string]interface{}{"foo": 123.0}, wantObject: success},
		{name: "string", data: "foo", wantObject: failure},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewValue(reporter, tc.data).IsObject().
				chain.assert(t, tc.wantObject)

			NewValue(reporter, tc.data).NotObject().
				chain.assert(t, !tc.wantObject)
		})
	}
}

func TestValue_IsArray(t *testing.T) {
	cases := []struct {
		name      string
		data      interface{}
		wantArray chainResult
	}{
		{name: "array", data: []interface{}{"foo", "123"}, wantArray: success},
		{name: "string", data: "foo", wantArray: failure},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewValue(reporter, tc.data).IsArray().
				chain.assert(t, tc.wantArray)

			NewValue(reporter, tc.data).NotArray().
				chain.assert(t, !tc.wantArray)
		})
	}
}

func TestValue_IsString(t *testing.T) {
	cases := []struct {
		name       string
		data       interface{}
		wantString chainResult
	}{
		{name: "string", data: "foo", wantString: success},
		{name: "integer", data: 123, wantString: failure},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewValue(reporter, tc.data).IsString().
				chain.assert(t, tc.wantString)

			NewValue(reporter, tc.data).NotString().
				chain.assert(t, !tc.wantString)
		})
	}
}

func TestValue_IsNumber(t *testing.T) {
	cases := []struct {
		name       string
		data       interface{}
		wantNumber chainResult
	}{
		{name: "integer", data: 123, wantNumber: success},
		{name: "string", data: "foo", wantNumber: failure},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewValue(reporter, tc.data).IsNumber().
				chain.assert(t, tc.wantNumber)

			NewValue(reporter, tc.data).NotNumber().
				chain.assert(t, !tc.wantNumber)
		})
	}
}

func TestValue_IsBoolean(t *testing.T) {
	cases := []struct {
		name     string
		data     interface{}
		wantBool chainResult
	}{
		{name: "bool", data: true, wantBool: success},
		{name: "string", data: "foo", wantBool: failure},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewValue(reporter, tc.data).IsBoolean().
				chain.assert(t, tc.wantBool)

			NewValue(reporter, tc.data).NotBoolean().
				chain.assert(t, !tc.wantBool)
		})
	}
}

func TestValue_IsEqual(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		cases := []struct {
			name      string
			value1    interface{}
			value2    interface{}
			wantEqual chainResult
		}{
			{
				name:      "compare equivalent values (strings)",
				value1:    "baz",
				value2:    "baz",
				wantEqual: success,
			},
			{
				name:      "compare equivalent values (maps)",
				value1:    map[string]interface{}{"foo": "bar"},
				value2:    map[string]interface{}{"foo": "bar"},
				wantEqual: success,
			},
			{
				name:      "compare non-equivalent values",
				value1:    map[string]interface{}{"foo": "bar"},
				value2:    "baz",
				wantEqual: failure,
			},
			{
				name:      "compare nil values",
				value1:    nil,
				value2:    nil,
				wantEqual: success,
			},
			{
				name:      "compare nil and nil-value",
				value1:    nil,
				value2:    map[string]interface{}(nil),
				wantEqual: success,
			},
			{
				name:      "compare nil and value",
				value1:    nil,
				value2:    map[string]interface{}{},
				wantEqual: failure,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewValue(reporter, tc.value1).IsEqual(tc.value2).
					chain.assert(t, tc.wantEqual)

				NewValue(reporter, tc.value1).NotEqual(tc.value2).
					chain.assert(t, !tc.wantEqual)
			})
		}
	})

	t.Run("invalid argument", func(t *testing.T) {
		cases := []struct {
			name         string
			value1       interface{}
			value2       interface{}
			wantEqual    chainResult
			wantNotEqual chainResult
		}{
			{
				name:         "compare value and func",
				value1:       map[string]interface{}{"foo": "bar"},
				value2:       func() {},
				wantEqual:    failure,
				wantNotEqual: failure,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewValue(reporter, tc.value1).IsEqual(tc.value2).
					chain.assert(t, tc.wantEqual)

				NewValue(reporter, tc.value1).NotEqual(tc.value2).
					chain.assert(t, tc.wantNotEqual)
			})
		}
	})
}

func TestValue_InList(t *testing.T) {
	type dataStruct struct {
		Data []int `json:"data"`
	}

	t.Run("basic", func(t *testing.T) {
		cases := []struct {
			name       string
			value      interface{}
			list       []interface{}
			wantInList chainResult
		}{
			{
				name:  "in list",
				value: map[string]interface{}{"foo": "bar"},
				list: []interface{}{map[string]interface{}{"foo": "bar"}, dataStruct{
					Data: []int{1, 2, 3, 4},
				}},
				wantInList: success,
			},
			{
				name:  "not in list",
				value: "baz",
				list: []interface{}{map[string]interface{}{"foo": "bar"}, dataStruct{
					Data: []int{1, 2, 3, 4},
				}},
				wantInList: failure,
			},
			{
				name:       "map not in list of string",
				value:      map[string]interface{}{"foo": "bar"},
				list:       []interface{}{"baz"},
				wantInList: failure,
			},
			{
				name:       "map in list of map",
				value:      map[string]interface{}{"foo": "bar"},
				list:       []interface{}{map[string]interface{}{"foo": "bar"}},
				wantInList: success,
			},
			{
				name:       "nil in list of nil-map",
				value:      nil,
				list:       []interface{}{map[string]interface{}(nil)},
				wantInList: success,
			},
			{
				name:       "nil not in list of empty map",
				value:      nil,
				list:       []interface{}{map[string]interface{}{}},
				wantInList: failure,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewValue(reporter, tc.value).InList(tc.list...).
					chain.assert(t, tc.wantInList)

				NewValue(reporter, tc.value).NotInList(tc.list...).
					chain.assert(t, !tc.wantInList)
			})
		}
	})

	t.Run("invalid argument", func(t *testing.T) {
		cases := []struct {
			name          string
			value         interface{}
			list          []interface{}
			wantInList    chainResult
			wantNotInList chainResult
		}{
			{
				name:          "nil list",
				value:         map[string]interface{}{"foo": "bar"},
				list:          nil,
				wantInList:    failure,
				wantNotInList: failure,
			},
			{
				name:          "empty list",
				value:         map[string]interface{}{"foo": "bar"},
				list:          []interface{}{},
				wantInList:    failure,
				wantNotInList: failure,
			},
			{
				name:          "list of a func",
				value:         map[string]interface{}{"foo": "bar"},
				list:          []interface{}{func() {}},
				wantInList:    failure,
				wantNotInList: failure,
			},
			{
				name:          "list of map and func",
				value:         map[string]interface{}{"foo": "bar"},
				list:          []interface{}{map[string]interface{}{"foo": "bar"}, func() {}},
				wantInList:    failure,
				wantNotInList: failure,
			},
			{
				name:          "list of string and func",
				value:         map[string]interface{}{"foo": "bar"},
				list:          []interface{}{"baz", func() {}},
				wantInList:    failure,
				wantNotInList: failure,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewValue(reporter, tc.value).InList(tc.list...).
					chain.assert(t, tc.wantInList)

				NewValue(reporter, tc.value).NotInList(tc.list...).
					chain.assert(t, tc.wantNotInList)
			})
		}
	})
}

func TestValue_PathTypes(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("object", func(t *testing.T) {
		user0 := map[string]interface{}{"name": "john"}
		user1 := map[string]interface{}{"name": "bob"}

		data := map[string]interface{}{
			"users": []interface{}{
				user0,
				user1,
			},
		}

		t.Run("queries", func(t *testing.T) {
			value := NewValue(reporter, data)

			assert.Equal(t, data, value.Path("$").Raw())
			assert.Equal(t, data["users"], value.Path("$.users").Raw())
			assert.Equal(t, user0, value.Path("$.users[0]").Raw())
			assert.Equal(t, "john", value.Path("$.users[0].name").Raw())
			assert.Equal(t, []interface{}{"john", "bob"}, value.Path("$.users[*].name").Raw())
			assert.Equal(t, []interface{}{"john", "bob"}, value.Path("$..name").Raw())
			value.chain.assert(t, success)

			names := value.Path("$..name").Array().Iter()
			names[0].String().IsEqual("john").chain.assert(t, success)
			names[1].String().IsEqual("bob").chain.assert(t, success)
			value.chain.assert(t, success)
		})

		t.Run("bad key", func(t *testing.T) {
			value := NewValue(reporter, data)

			bad := value.Path("$.bad")
			assert.True(t, bad != nil)
			assert.True(t, bad.Raw() == nil)
			value.chain.assert(t, failure)
		})

		t.Run("invalid query", func(t *testing.T) {
			value := NewValue(reporter, data)

			bad := value.Path("!")
			assert.True(t, bad != nil)
			assert.True(t, bad.Raw() == nil)
			value.chain.assert(t, failure)
		})
	})

	t.Run("array", func(t *testing.T) {
		user0 := map[string]interface{}{"name": "john"}
		user1 := map[string]interface{}{"name": "bob"}

		data := []interface{}{
			user0,
			user1,
		}

		value := NewValue(reporter, data)

		assert.Equal(t, data, value.Path("$").Raw())
		assert.Equal(t, user0, value.Path("$[0]").Raw())
		assert.Equal(t, "john", value.Path("$[0].name").Raw())
		assert.Equal(t, []interface{}{"john", "bob"}, value.Path("$[*].name").Raw())
		assert.Equal(t, []interface{}{"john", "bob"}, value.Path("$..name").Raw())
		value.chain.assert(t, success)
	})

	t.Run("string", func(t *testing.T) {
		data := "foo"

		value := NewValue(reporter, data)

		assert.Equal(t, data, value.Path("$").Raw())
		value.chain.assert(t, success)
	})

	t.Run("number", func(t *testing.T) {
		data := 123

		value := NewValue(reporter, data)

		assert.Equal(t, float64(data), value.Path("$").Raw())
		value.chain.assert(t, success)
	})

	t.Run("boolean", func(t *testing.T) {
		data := true

		value := NewValue(reporter, data)

		assert.Equal(t, data, value.Path("$").Raw())
		value.chain.assert(t, success)
	})

	t.Run("null", func(t *testing.T) {
		value := NewValue(reporter, nil)

		assert.Equal(t, nil, value.Path("$").Raw())
		value.chain.assert(t, success)
	})

	t.Run("error", func(t *testing.T) {
		data := "foo"

		value := NewValue(reporter, data)

		for _, key := range []string{"$.bad", "!"} {
			bad := value.Path(key)
			assert.True(t, bad != nil)
			assert.True(t, bad.Raw() == nil)
			value.chain.assert(t, failure)
		}
	})

	t.Run("int float", func(t *testing.T) {
		data := map[string]interface{}{
			"A": 123,
			"B": 123.0,
		}

		value := NewValue(reporter, data)
		value.chain.assert(t, success)

		a := value.Path(`$["A"]`)
		a.chain.assert(t, success)
		assert.Equal(t, 123.0, a.Raw())

		b := value.Path(`$["B"]`)
		b.chain.assert(t, success)
		assert.Equal(t, 123.0, b.Raw())
	})
}

// based on github.com/yalp/jsonpath
func TestValue_PathExpressions(t *testing.T) {
	data := map[string]interface{}{
		"A": []interface{}{
			"string",
			23.3,
			3.0,
			true,
			false,
			nil,
		},
		"B": "value",
		"C": 3.14,
		"D": map[string]interface{}{
			"C": 3.1415,
			"V": []interface{}{
				"string2a",
				"string2b",
				map[string]interface{}{
					"C": 3.141592,
				},
			},
		},
		"E": map[string]interface{}{
			"A": []interface{}{"string3"},
			"D": map[string]interface{}{
				"V": map[string]interface{}{
					"C": 3.14159265,
				},
			},
		},
		"F": map[string]interface{}{
			"V": []interface{}{
				"string4a",
				"string4b",
				map[string]interface{}{
					"CC": 3.1415926535,
				},
				map[string]interface{}{
					"CC": "hello",
				},
				[]interface{}{
					"string5a",
					"string5b",
				},
				[]interface{}{
					"string6a",
					"string6b",
				},
			},
		},
	}

	reporter := newMockReporter(t)

	runTests := func(tests map[string]interface{}) {
		value := NewValue(reporter, data)
		value.chain.assert(t, success)

		for path, expected := range tests {
			actual := value.Path(path)
			actual.chain.assert(t, success)

			assert.Equal(t, expected, actual.Raw())
		}
	}

	t.Run("pick", func(t *testing.T) {
		runTests(map[string]interface{}{
			"$":         data,
			"$.A[0]":    "string",
			`$["A"][0]`: "string",
			"$.A":       []interface{}{"string", 23.3, 3.0, true, false, nil},
			"$.A[*]":    []interface{}{"string", 23.3, 3.0, true, false, nil},
			"$.A.*":     []interface{}{"string", 23.3, 3.0, true, false, nil},
			"$.A.*.a":   []interface{}{},
		})
	})

	t.Run("slice", func(t *testing.T) {
		runTests(map[string]interface{}{
			"$.A[1,4,2]":      []interface{}{23.3, false, 3.0},
			`$["B","C"]`:      []interface{}{"value", 3.14},
			`$["C","B"]`:      []interface{}{3.14, "value"},
			"$.A[1:4]":        []interface{}{23.3, 3.0, true},
			"$.A[::2]":        []interface{}{"string", 3.0, false},
			"$.A[-2:]":        []interface{}{false, nil},
			"$.A[:-1]":        []interface{}{"string", 23.3, 3.0, true, false},
			"$.A[::-1]":       []interface{}{nil, false, true, 3.0, 23.3, "string"},
			"$.F.V[4:5][0,1]": []interface{}{"string5a", "string5b"},
			"$.F.V[4:6][1]":   []interface{}{"string5b", "string6b"},
			"$.F.V[4:6][0,1]": []interface{}{"string5a", "string5b", "string6a", "string6b"},
			"$.F.V[4,5][0:2]": []interface{}{"string5a", "string5b", "string6a", "string6b"},
			"$.F.V[4:6]": []interface{}{
				[]interface{}{
					"string5a",
					"string5b",
				},
				[]interface{}{
					"string6a",
					"string6b",
				},
			},
		})
	})

	t.Run("quote", func(t *testing.T) {
		runTests(map[string]interface{}{
			`$[A][0]`:    "string",
			`$["A"][0]`:  "string",
			`$[B,C]`:     []interface{}{"value", 3.14},
			`$["B","C"]`: []interface{}{"value", 3.14},
		})
	})

	t.Run("search", func(t *testing.T) {
		runTests(map[string]interface{}{
			"$..C":       []interface{}{3.14, 3.1415, 3.141592, 3.14159265},
			`$..["C"]`:   []interface{}{3.14, 3.1415, 3.141592, 3.14159265},
			"$.D.V..C":   []interface{}{3.141592},
			"$.D.V.*.C":  []interface{}{3.141592},
			"$.D.V..*.C": []interface{}{3.141592},
			"$.D.*..C":   []interface{}{3.141592},
			"$.*.V..C":   []interface{}{3.141592},
			"$.*.D.V.C":  []interface{}{3.14159265},
			"$.*.D..C":   []interface{}{3.14159265},
			"$.*.D.V..*": []interface{}{3.14159265},
			"$..D..V..C": []interface{}{3.141592, 3.14159265},
			"$.*.*.*.C":  []interface{}{3.141592, 3.14159265},
			"$..V..C":    []interface{}{3.141592, 3.14159265},
			"$.D.V..*": []interface{}{
				"string2a",
				"string2b",
				map[string]interface{}{
					"C": 3.141592,
				},
				3.141592,
			},
			"$..A": []interface{}{
				[]interface{}{"string", 23.3, 3.0, true, false, nil},
				[]interface{}{"string3"},
			},
			"$..A..*":      []interface{}{"string", 23.3, 3.0, true, false, nil, "string3"},
			"$.A..*":       []interface{}{"string", 23.3, 3.0, true, false, nil},
			"$.A.*":        []interface{}{"string", 23.3, 3.0, true, false, nil},
			"$..A[0,1]":    []interface{}{"string", 23.3},
			"$..A[0]":      []interface{}{"string", "string3"},
			"$.*.V[0]":     []interface{}{"string2a", "string4a"},
			"$.*.V[1]":     []interface{}{"string2b", "string4b"},
			"$.*.V[0,1]":   []interface{}{"string2a", "string2b", "string4a", "string4b"},
			"$.*.V[0:2]":   []interface{}{"string2a", "string2b", "string4a", "string4b"},
			"$.*.V[2].C":   []interface{}{3.141592},
			"$..V[2].C":    []interface{}{3.141592},
			"$..V[*].C":    []interface{}{3.141592},
			"$.*.V[2].*":   []interface{}{3.141592, 3.1415926535},
			"$.*.V[2:3].*": []interface{}{3.141592, 3.1415926535},
			"$.*.V[2:4].*": []interface{}{3.141592, 3.1415926535, "hello"},
			"$..V[2,3].CC": []interface{}{3.1415926535, "hello"},
			"$..V[2:4].CC": []interface{}{3.1415926535, "hello"},
			"$..V[*].*": []interface{}{
				3.141592,
				3.1415926535,
				"hello",
				"string5a",
				"string5b",
				"string6a",
				"string6b",
			},
			"$..[0]": []interface{}{
				"string",
				"string2a",
				"string3",
				"string4a",
				"string5a",
				"string6a",
			},
			"$..ZZ": []interface{}{},
		})
	})
}

func TestValue_Schema(t *testing.T) {
	reporter := newMockReporter(t)

	schema := `{
		"type": "object",
		"properties": {
			"foo": {
				"type": "string"
			},
			"bar": {
				"type": "integer"
			}
		},
		"require": ["foo", "bar"]
	}`

	data1 := map[string]interface{}{
		"foo": "a",
		"bar": 1,
	}

	data2 := map[string]interface{}{
		"foo": "a",
		"bar": "b",
	}

	NewValue(reporter, data1).Schema(schema).chain.assert(t, success)
	NewValue(reporter, data2).Schema(schema).chain.assert(t, failure)

	NewValue(reporter, data1).Schema([]byte(schema)).chain.assert(t, success)
	NewValue(reporter, data2).Schema([]byte(schema)).chain.assert(t, failure)

	var b interface{}
	err := json.Unmarshal([]byte(schema), &b)
	require.Nil(t, err)

	NewValue(reporter, data1).Schema(b).chain.assert(t, success)
	NewValue(reporter, data2).Schema(b).chain.assert(t, failure)

	tmp, _ := ioutil.TempFile("", "httpexpect")
	defer os.Remove(tmp.Name())

	_, err = tmp.Write([]byte(schema))
	require.Nil(t, err)

	err = tmp.Close()
	require.Nil(t, err)

	url := "file://" + tmp.Name()

	NewValue(reporter, data1).Schema(url).chain.assert(t, success)
	NewValue(reporter, data2).Schema(url).chain.assert(t, failure)

	NewValue(reporter, data1).Schema("file:///bad/path").chain.assert(t, failure)
	NewValue(reporter, data1).Schema("{ bad json").chain.assert(t, failure)
}
