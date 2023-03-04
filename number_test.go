package httpexpect

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNumber_FailedChain(t *testing.T) {
	chain := newMockChain(t)
	chain.setFailed()

	value := newNumber(chain, 0)
	value.chain.assertFailed(t)

	value.Path("$").chain.assertFailed(t)
	value.Schema("")
	value.Alias("foo")

	var target interface{}
	value.Decode(&target)

	value.IsEqual(0)
	value.NotEqual(0)
	value.InDelta(0, 0)
	value.NotInDelta(0, 0)
	value.InRange(0, 0)
	value.NotInRange(0, 0)
	value.InList(0)
	value.NotInList(0)
	value.Gt(0)
	value.Ge(0)
	value.Lt(0)
	value.Le(0)
	value.IsInt()
	value.NotInt()
	value.IsUint()
	value.NotUint()
	value.IsFinite()
	value.NotFinite()
}

func TestNumber_Constructors(t *testing.T) {
	t.Run("reporter", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewNumber(reporter, 10.3)
		value.IsEqual(10.3)
		value.chain.assertNotFailed(t)
	})

	t.Run("config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewNumberC(Config{
			Reporter: reporter,
		}, 10.3)
		value.IsEqual(10.3)
		value.chain.assertNotFailed(t)
	})

	t.Run("chain", func(t *testing.T) {
		chain := newMockChain(t)
		value := newNumber(chain, 10.3)
		assert.NotSame(t, value.chain, chain)
		assert.Equal(t, value.chain.context.Path, chain.context.Path)
	})
}

func TestNumber_Decode(t *testing.T) {
	t.Run("target is empty interface", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewNumber(reporter, 10.1)

		var target interface{}
		value.Decode(&target)

		value.chain.assertNotFailed(t)
		assert.Equal(t, json.Number("10.1"), target)
	})

	t.Run("target is int", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewNumber(reporter, 10)

		var target int
		value.Decode(&target)

		value.chain.assertNotFailed(t)
		assert.Equal(t, 10, target)
	})

	t.Run("target is float64", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewNumber(reporter, 10.1)

		var target float64
		value.Decode(&target)

		value.chain.assertNotFailed(t)
		assert.Equal(t, 10.1, target)
	})

	t.Run("target is nil", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewNumber(reporter, 10.1)

		value.Decode(nil)

		value.chain.assertFailed(t)
	})

	t.Run("target is unmarshable", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewNumber(reporter, 10.1)

		value.Decode(123)

		value.chain.assertFailed(t)
	})
}

func TestNumber_Alias(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewNumber(reporter, 123)
	assert.Equal(t, []string{"Number()"}, value.chain.context.Path)
	assert.Equal(t, []string{"Number()"}, value.chain.context.AliasedPath)

	value.Alias("foo")
	assert.Equal(t, []string{"Number()"}, value.chain.context.Path)
	assert.Equal(t, []string{"foo"}, value.chain.context.AliasedPath)
}

func TestNumber_Getters(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewNumber(reporter, 123.0)

	assert.Equal(t, 123.0, value.Raw())
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	assert.Equal(t, json.Number("123"), value.Path("$").Raw())
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Schema(`{"type": "number"}`)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Schema(`{"type": "object"}`)
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestNumber_IsEqual(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		cases := []struct {
			name    string
			number  float64
			value   interface{}
			isEqual bool
		}{
			{
				name:    "compare equivalent integers",
				number:  1234,
				value:   1234,
				isEqual: true,
			},
			{
				name:    "compare non-equivalent integers",
				number:  1234,
				value:   4321,
				isEqual: false,
			},
			{
				name:    "compare NaN to float",
				number:  math.NaN(),
				value:   1234.5,
				isEqual: false,
			},
			{
				name:    "compare float to NaN",
				number:  1234.5,
				value:   math.NaN(),
				isEqual: false,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				if tc.isEqual {
					NewNumber(reporter, tc.number).
						IsEqual(tc.value).
						chain.assertNotFailed(t)

					NewNumber(reporter, tc.number).
						NotEqual(tc.value).
						chain.assertFailed(t)
				} else {
					NewNumber(reporter, tc.number).
						NotEqual(tc.value).
						chain.assertNotFailed(t)

					NewNumber(reporter, tc.number).
						IsEqual(tc.value).
						chain.assertFailed(t)
				}
			})
		}
	})

	t.Run("canonization", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewNumber(reporter, 1234).IsEqual(int64(1234)).
			chain.assertNotFailed(t)

		NewNumber(reporter, 1234).IsEqual(float32(1234)).
			chain.assertNotFailed(t)

		NewNumber(reporter, 1234).NotEqual(int64(4321)).
			chain.assertNotFailed(t)

		NewNumber(reporter, 1234).NotEqual(float32(4321)).
			chain.assertNotFailed(t)
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewNumber(reporter, 1234).IsEqual("NOT NUMBER").
			chain.assertFailed(t)

		NewNumber(reporter, 1234).NotEqual("NOT NUMBER").
			chain.assertFailed(t)
	})
}

func TestNumber_InDelta(t *testing.T) {
	cases := []struct {
		name         string
		number       float64
		value        float64
		delta        float64
		isInDelta    bool
		isNotInDelta bool
	}{
		{
			name:         "larger value in delta range",
			number:       1234.5,
			value:        1234.7,
			delta:        0.3,
			isInDelta:    true,
			isNotInDelta: false,
		},
		{
			name:         "smaller value in delta range",
			number:       1234.5,
			value:        1234.3,
			delta:        0.3,
			isInDelta:    true,
			isNotInDelta: false,
		},
		{
			name:         "larger value not in delta range",
			number:       1234.5,
			value:        1234.7,
			delta:        0.1,
			isInDelta:    false,
			isNotInDelta: true,
		},
		{
			name:         "smaller value not in delta range",
			number:       1234.5,
			value:        1234.3,
			delta:        0.1,
			isInDelta:    false,
			isNotInDelta: true,
		},
		{
			name:         "number is NaN",
			number:       math.NaN(),
			value:        1234.0,
			delta:        0.1,
			isInDelta:    false,
			isNotInDelta: false,
		},
		// {
		// 	name:         "value is NaN",
		// 	number:       1234.5,
		// 	value:        math.NaN(),
		// 	delta:        0.1,
		// 	isInDelta:    false,
		// 	isNotInDelta: false,
		// },
		// {
		// 	name:         "delta is NaN",
		// 	number:       1234.5,
		// 	value:        1234.0,
		// 	delta:        math.NaN(),
		// 	isInDelta:    false,
		// 	isNotInDelta: false,
		// },
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			if tc.isInDelta {
				NewNumber(reporter, tc.number).
					InDelta(tc.value, tc.delta).
					chain.assertNotFailed(t)
			} else {
				NewNumber(reporter, tc.number).
					InDelta(tc.value, tc.delta).
					chain.assertFailed(t)
			}

			if tc.isNotInDelta {
				NewNumber(reporter, tc.number).
					NotInDelta(tc.value, tc.delta).
					chain.assertNotFailed(t)
			} else {
				NewNumber(reporter, tc.number).
					NotInDelta(tc.value, tc.delta).
					chain.assertFailed(t)
			}
		})
	}
}

func TestNumber_InRange(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		cases := []struct {
			name         string
			number       float64
			min          interface{}
			max          interface{}
			isInRange    bool
			isNotInRange bool
		}{
			{
				name:         "range includes only number",
				number:       1234,
				min:          1234,
				max:          1234,
				isInRange:    true,
				isNotInRange: false,
			},
			{
				name:         "range includes number and below",
				number:       1234,
				min:          1234 - 1,
				max:          1234,
				isInRange:    true,
				isNotInRange: false,
			},
			{
				name:         "range includes number and above",
				number:       1234,
				min:          1234,
				max:          1234 + 1,
				isInRange:    true,
				isNotInRange: false,
			},
			{
				name:         "range is above number",
				number:       1234,
				min:          1234 + 1,
				max:          1234 + 2,
				isInRange:    false,
				isNotInRange: true,
			},
			{
				name:         "range is below number",
				number:       1234,
				min:          1234 - 2,
				max:          1234 - 1,
				isInRange:    false,
				isNotInRange: true,
			},
			{
				name:         "range min is larger than max",
				number:       1234,
				min:          1234 + 1,
				max:          1234 - 1,
				isInRange:    false,
				isNotInRange: true,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				if tc.isInRange {
					NewNumber(reporter, tc.number).
						InRange(tc.min, tc.max).
						chain.assertNotFailed(t)
				} else {
					NewNumber(reporter, tc.number).
						InRange(tc.min, tc.max).
						chain.assertFailed(t)
				}

				if tc.isNotInRange {
					NewNumber(reporter, tc.number).
						NotInRange(tc.min, tc.max).
						chain.assertNotFailed(t)
				} else {
					NewNumber(reporter, tc.number).
						NotInRange(tc.min, tc.max).
						chain.assertFailed(t)
				}
			})
		}
	})

	t.Run("canonization", func(t *testing.T) {
		cases := []struct {
			number    float64
			min       interface{}
			max       interface{}
			isInRange bool
		}{
			{
				number:    1234,
				min:       int64(1233),
				max:       float32(1235),
				isInRange: true,
			},
			{
				number:    1234,
				min:       1235,
				max:       1236,
				isInRange: false,
			},
		}

		reporter := newMockReporter(t)

		for _, tc := range cases {
			if tc.isInRange {
				NewNumber(reporter, tc.number).
					InRange(tc.min, tc.max).
					chain.assertNotFailed(t)

				NewNumber(reporter, tc.number).
					NotInRange(tc.min, tc.max).
					chain.assertFailed(t)
			} else {
				NewNumber(reporter, tc.number).
					NotInRange(tc.min, tc.max).
					chain.assertNotFailed(t)

				NewNumber(reporter, tc.number).
					InRange(tc.min, tc.max).
					chain.assertFailed(t)
			}
		}
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewNumber(reporter, 1234).InRange(int64(1233), "NOT NUMBER").
			chain.assertFailed(t)

		NewNumber(reporter, 1234).NotInRange(int64(1233), "NOT NUMBER").
			chain.assertFailed(t)

		NewNumber(reporter, 1234).InRange("NOT NUMBER", float32(1235)).
			chain.assertFailed(t)

		NewNumber(reporter, 1234).NotInRange("NOT NUMBER", float32(1235)).
			chain.assertFailed(t)
	})
}

func TestNumber_InList(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		cases := []struct {
			name        string
			number      float64
			list        []interface{}
			isInList    bool
			isNotInList bool
		}{
			{
				name:        "no list",
				number:      1234,
				list:        nil,
				isInList:    false,
				isNotInList: false,
			},
			{
				name:        "empty list",
				number:      1234,
				list:        []interface{}{},
				isInList:    false,
				isNotInList: false,
			},
			{
				name:        "in integer list",
				number:      1234,
				list:        []interface{}{1234, 4567},
				isInList:    true,
				isNotInList: false,
			},
			{
				name:        "in float list",
				number:      1234,
				list:        []interface{}{1234.00, 4567.00},
				isInList:    true,
				isNotInList: false,
			},
			{
				name:        "not in float list",
				number:      1234,
				list:        []interface{}{4567.00, 1234.01},
				isInList:    false,
				isNotInList: true,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				if tc.isInList {
					NewNumber(reporter, tc.number).
						InList(tc.list...).
						chain.assertNotFailed(t)
				} else {
					NewNumber(reporter, tc.number).
						InList(tc.list...).
						chain.assertFailed(t)
				}

				if tc.isNotInList {
					NewNumber(reporter, tc.number).
						NotInList(tc.list...).
						chain.assertNotFailed(t)
				} else {
					NewNumber(reporter, tc.number).
						NotInList(tc.list...).
						chain.assertFailed(t)
				}
			})
		}
	})

	t.Run("canonization", func(t *testing.T) {
		cases := []struct {
			number   float64
			list     []interface{}
			isInList bool
		}{
			{
				number:   111,
				list:     []interface{}{int64(111), float32(222)},
				isInList: true,
			},
			{
				number:   111,
				list:     []interface{}{float32(111), int64(222)},
				isInList: true,
			},
			{
				number:   111,
				list:     []interface{}{222, 333},
				isInList: false,
			},
		}

		reporter := newMockReporter(t)

		for _, tc := range cases {
			if tc.isInList {
				NewNumber(reporter, tc.number).
					InList(tc.list...).
					chain.assertNotFailed(t)

				NewNumber(reporter, tc.number).
					NotInList(tc.list...).
					chain.assertFailed(t)
			} else {
				NewNumber(reporter, tc.number).
					NotInList(tc.list...).
					chain.assertNotFailed(t)

				NewNumber(reporter, tc.number).
					InList(tc.list...).
					chain.assertFailed(t)
			}
		}
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewNumber(reporter, 111).InList(222, "NOT NUMBER").
			chain.assertFailed(t)

		NewNumber(reporter, 111).NotInList(222, "NOT NUMBER").
			chain.assertFailed(t)

		NewNumber(reporter, 111).InList(111, "NOT NUMBER").
			chain.assertFailed(t)

		NewNumber(reporter, 111).NotInList(111, "NOT NUMBER").
			chain.assertFailed(t)
	})
}

func TestNumber_IsGreater(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		cases := []struct {
			name   string
			number float64
			value  interface{}
			isGt   bool
			isGe   bool
		}{
			{
				name:   "number is lesser",
				number: 1234,
				value:  1234 + 1,
				isGt:   false,
				isGe:   false,
			},
			{
				name:   "number is equal",
				number: 1234,
				value:  1234,
				isGt:   false,
				isGe:   true,
			},
			{
				name:   "number is greater",
				number: 1234,
				value:  1234 - 1,
				isGt:   true,
				isGe:   true,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				if tc.isGt {
					NewNumber(reporter, tc.number).
						Gt(tc.value).
						chain.assertNotFailed(t)
				} else {
					NewNumber(reporter, tc.number).
						Gt(tc.value).
						chain.assertFailed(t)
				}

				if tc.isGe {
					NewNumber(reporter, tc.number).
						Ge(tc.value).
						chain.assertNotFailed(t)
				} else {
					NewNumber(reporter, tc.number).
						Ge(tc.value).
						chain.assertFailed(t)
				}
			})
		}
	})

	t.Run("canonization", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewNumber(reporter, 1234).Gt(int64(1233)).
			chain.assertNotFailed(t)

		NewNumber(reporter, 1234).Gt(float32(1233)).
			chain.assertNotFailed(t)

		NewNumber(reporter, 1234).Ge(int64(1233)).
			chain.assertNotFailed(t)

		NewNumber(reporter, 1234).Ge(float32(1233)).
			chain.assertNotFailed(t)
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewNumber(reporter, 1234).Gt("NOT NUMBER").
			chain.assertFailed(t)

		NewNumber(reporter, 1234).Ge("NOT NUMBER").
			chain.assertFailed(t)
	})
}

func TestNumber_IsLesser(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		cases := []struct {
			name   string
			number float64
			value  interface{}
			isLt   bool
			isLe   bool
		}{
			{
				name:   "number is lesser",
				number: 1234,
				value:  1234 + 1,
				isLt:   true,
				isLe:   true,
			},
			{
				name:   "number is equal",
				number: 1234,
				value:  1234,
				isLt:   false,
				isLe:   true,
			},
			{
				name:   "number is greater",
				number: 1234,
				value:  1234 - 1,
				isLt:   false,
				isLe:   false,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				if tc.isLt {
					NewNumber(reporter, tc.number).
						Lt(tc.value).
						chain.assertNotFailed(t)
				} else {
					NewNumber(reporter, tc.number).
						Lt(tc.value).
						chain.assertFailed(t)
				}

				if tc.isLe {
					NewNumber(reporter, tc.number).
						Le(tc.value).
						chain.assertNotFailed(t)
				} else {
					NewNumber(reporter, tc.number).
						Le(tc.value).
						chain.assertFailed(t)
				}
			})
		}
	})

	t.Run("canonization", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewNumber(reporter, 1234).Lt(int64(1235)).
			chain.assertNotFailed(t)

		NewNumber(reporter, 1234).Lt(float32(1235)).
			chain.assertNotFailed(t)

		NewNumber(reporter, 1234).Le(int64(1235)).
			chain.assertNotFailed(t)

		NewNumber(reporter, 1234).Le(float32(1235)).
			chain.assertNotFailed(t)
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewNumber(reporter, 1234).Lt("NOT NUMBER").
			chain.assertFailed(t)

		NewNumber(reporter, 1234).Le("NOT NUMBER").
			chain.assertFailed(t)
	})
}

func TestNumber_IsInt(t *testing.T) {
	t.Run("values", func(t *testing.T) {
		tests := []struct {
			name    string
			value   float64
			isInt16 bool
			isInt32 bool
			isInt   bool
		}{
			{
				name:    "0",
				value:   0,
				isInt16: true,
				isInt32: true,
				isInt:   true,
			},
			{
				name:    "1",
				value:   1,
				isInt16: true,
				isInt32: true,
				isInt:   true,
			},
			{
				name:    "0.5",
				value:   0.5,
				isInt16: false,
				isInt32: false,
				isInt:   false,
			},
			{
				name:    "NaN",
				value:   math.NaN(),
				isInt16: false,
				isInt32: false,
				isInt:   false,
			},
			{
				name:    "-Inf",
				value:   math.Inf(-1),
				isInt16: false,
				isInt32: false,
				isInt:   false,
			},
			{
				name:    "+Inf",
				value:   math.Inf(+1),
				isInt16: false,
				isInt32: false,
				isInt:   false,
			},
			{
				name:    "MinInt16-1",
				value:   math.MinInt16 - 1,
				isInt16: false,
				isInt32: true,
				isInt:   true,
			},
			{
				name:    "MinInt16",
				value:   math.MinInt16,
				isInt16: true,
				isInt32: true,
				isInt:   true,
			},
			{
				name:    "MaxInt16",
				value:   math.MaxInt16,
				isInt16: true,
				isInt32: true,
				isInt:   true,
			},
			{
				name:    "MaxInt16+1",
				value:   math.MaxInt16 + 1,
				isInt16: false,
				isInt32: true,
				isInt:   true,
			},
			{
				name:    "MinInt32-1",
				value:   math.MinInt32 - 1,
				isInt16: false,
				isInt32: false,
				isInt:   true,
			},
			{
				name:    "MinInt32",
				value:   math.MinInt32,
				isInt16: false,
				isInt32: true,
				isInt:   true,
			},
			{
				name:    "MaxInt32",
				value:   math.MaxInt32,
				isInt16: false,
				isInt32: true,
				isInt:   true,
			},
			{
				name:    "MaxInt32+1",
				value:   math.MaxInt32 + 1,
				isInt16: false,
				isInt32: false,
				isInt:   true,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				if tc.isInt {
					NewNumber(reporter, tc.value).IsInt().chain.assertNotFailed(t)
					NewNumber(reporter, tc.value).NotInt().chain.assertFailed(t)
				} else {
					NewNumber(reporter, tc.value).IsInt().chain.assertFailed(t)
					NewNumber(reporter, tc.value).NotInt().chain.assertNotFailed(t)
				}

				if tc.isInt32 {
					NewNumber(reporter, tc.value).IsInt(32).chain.assertNotFailed(t)
					NewNumber(reporter, tc.value).NotInt(32).chain.assertFailed(t)
				} else {
					NewNumber(reporter, tc.value).IsInt(32).chain.assertFailed(t)
					NewNumber(reporter, tc.value).NotInt(32).chain.assertNotFailed(t)
				}

				if tc.isInt16 {
					NewNumber(reporter, tc.value).IsInt(16).chain.assertNotFailed(t)
					NewNumber(reporter, tc.value).NotInt(16).chain.assertFailed(t)
				} else {
					NewNumber(reporter, tc.value).IsInt(16).chain.assertFailed(t)
					NewNumber(reporter, tc.value).NotInt(16).chain.assertNotFailed(t)
				}
			})
		}
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewNumber(reporter, 1234).IsInt(16, 32).
			chain.assertFailed(t)

		NewNumber(reporter, 1234).NotInt(16, 32).
			chain.assertFailed(t)

		NewNumber(reporter, 1234).IsInt(0).
			chain.assertFailed(t)

		NewNumber(reporter, 1234).NotInt(0).
			chain.assertFailed(t)

		NewNumber(reporter, 1234).IsInt(-16).
			chain.assertFailed(t)

		NewNumber(reporter, 1234).NotInt(-16).
			chain.assertFailed(t)
	})
}

func TestNumber_IsUint(t *testing.T) {
	t.Run("values", func(t *testing.T) {
		tests := []struct {
			name     string
			value    float64
			isUint16 bool
			isUint32 bool
			isUint   bool
		}{
			{
				name:     "0",
				value:    0,
				isUint16: true,
				isUint32: true,
				isUint:   true,
			},
			{
				name:     "1",
				value:    1,
				isUint16: true,
				isUint32: true,
				isUint:   true,
			},
			{
				name:     "-1",
				value:    -1,
				isUint16: false,
				isUint32: false,
				isUint:   false,
			},
			{
				name:     "0.5",
				value:    0.5,
				isUint16: false,
				isUint32: false,
				isUint:   false,
			},
			{
				name:     "NaN",
				value:    math.NaN(),
				isUint16: false,
				isUint32: false,
				isUint:   false,
			},
			{
				name:     "-Inf",
				value:    math.Inf(-1),
				isUint16: false,
				isUint32: false,
				isUint:   false,
			},
			{
				name:     "+Inf",
				value:    math.Inf(+1),
				isUint16: false,
				isUint32: false,
				isUint:   false,
			},
			{
				name:     "MaxUint16",
				value:    math.MaxUint16,
				isUint16: true,
				isUint32: true,
				isUint:   true,
			},
			{
				name:     "MaxUint16+1",
				value:    math.MaxUint16 + 1,
				isUint16: false,
				isUint32: true,
				isUint:   true,
			},
			{
				name:     "MaxUint32",
				value:    math.MaxUint32,
				isUint16: false,
				isUint32: true,
				isUint:   true,
			},
			{
				name:     "MaxUint32+1",
				value:    math.MaxUint32 + 1,
				isUint16: false,
				isUint32: false,
				isUint:   true,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				if tc.isUint {
					NewNumber(reporter, tc.value).IsUint().chain.assertNotFailed(t)
					NewNumber(reporter, tc.value).NotUint().chain.assertFailed(t)
				} else {
					NewNumber(reporter, tc.value).IsUint().chain.assertFailed(t)
					NewNumber(reporter, tc.value).NotUint().chain.assertNotFailed(t)
				}

				if tc.isUint32 {
					NewNumber(reporter, tc.value).IsUint(32).chain.assertNotFailed(t)
					NewNumber(reporter, tc.value).NotUint(32).chain.assertFailed(t)
				} else {
					NewNumber(reporter, tc.value).IsUint(32).chain.assertFailed(t)
					NewNumber(reporter, tc.value).NotUint(32).chain.assertNotFailed(t)
				}

				if tc.isUint16 {
					NewNumber(reporter, tc.value).IsUint(16).chain.assertNotFailed(t)
					NewNumber(reporter, tc.value).NotUint(16).chain.assertFailed(t)
				} else {
					NewNumber(reporter, tc.value).IsUint(16).chain.assertFailed(t)
					NewNumber(reporter, tc.value).NotUint(16).chain.assertNotFailed(t)
				}
			})
		}
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewNumber(reporter, 1234).IsUint(16, 32).
			chain.assertFailed(t)

		NewNumber(reporter, 1234).NotUint(16, 32).
			chain.assertFailed(t)

		NewNumber(reporter, 1234).IsUint(0).
			chain.assertFailed(t)

		NewNumber(reporter, 1234).NotUint(0).
			chain.assertFailed(t)

		NewNumber(reporter, 1234).IsUint(-16).
			chain.assertFailed(t)

		NewNumber(reporter, 1234).NotUint(-16).
			chain.assertFailed(t)
	})
}

func TestNumber_IsFinite(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		isFinite bool
	}{
		{
			name:     "0",
			value:    0,
			isFinite: true,
		},
		{
			name:     "1",
			value:    1,
			isFinite: true,
		},
		{
			name:     "-1",
			value:    -1,
			isFinite: true,
		},
		{
			name:     "0.5",
			value:    0.5,
			isFinite: true,
		},
		{
			name:     "NaN",
			value:    math.NaN(),
			isFinite: false,
		},
		{
			name:     "-Inf",
			value:    math.Inf(-1),
			isFinite: false,
		},
		{
			name:     "+Inf",
			value:    math.Inf(+1),
			isFinite: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			if tc.isFinite {
				NewNumber(reporter, tc.value).IsFinite().chain.assertNotFailed(t)
				NewNumber(reporter, tc.value).NotFinite().chain.assertFailed(t)
			} else {
				NewNumber(reporter, tc.value).IsFinite().chain.assertFailed(t)
				NewNumber(reporter, tc.value).NotFinite().chain.assertNotFailed(t)
			}
		})
	}
}
