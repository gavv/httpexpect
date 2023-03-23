package httpexpect

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNumber_FailedChain(t *testing.T) {
	chain := newMockChain(t, flagFailed)

	value := newNumber(chain, 0)
	value.chain.assert(t, failure)

	value.Path("$").chain.assert(t, failure)
	value.Schema("")
	value.Alias("foo")

	var target interface{}
	value.Decode(&target)

	value.IsEqual(0)
	value.NotEqual(0)
	value.InDelta(0, 0)
	value.NotInDelta(0, 0)
	value.InDeltaRelative(0, 0)
	value.NotInDeltaRelative(0, 0)
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
		value.chain.assert(t, success)
	})

	t.Run("config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewNumberC(Config{
			Reporter: reporter,
		}, 10.3)
		value.IsEqual(10.3)
		value.chain.assert(t, success)
	})

	t.Run("chain", func(t *testing.T) {
		chain := newMockChain(t)
		value := newNumber(chain, 10.3)
		assert.NotSame(t, value.chain, chain)
		assert.Equal(t, value.chain.context.Path, chain.context.Path)
	})
}

func TestNumber_Raw(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewNumber(reporter, 123.0)

	assert.Equal(t, 123.0, value.Raw())
	value.chain.assert(t, success)
}

func TestNumber_Decode(t *testing.T) {
	t.Run("target is empty interface", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewNumber(reporter, 10.1)

		var target interface{}
		value.Decode(&target)

		value.chain.assert(t, success)
		assert.Equal(t, 10.1, target)
	})

	t.Run("target is int", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewNumber(reporter, 10)

		var target int
		value.Decode(&target)

		value.chain.assert(t, success)
		assert.Equal(t, 10, target)
	})

	t.Run("target is float64", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewNumber(reporter, 10.1)

		var target float64
		value.Decode(&target)

		value.chain.assert(t, success)
		assert.Equal(t, 10.1, target)
	})

	t.Run("target is nil", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewNumber(reporter, 10.1)

		value.Decode(nil)

		value.chain.assert(t, failure)
	})

	t.Run("target is unmarshable", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewNumber(reporter, 10.1)

		value.Decode(123)

		value.chain.assert(t, failure)
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

func TestNumber_Path(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewNumber(reporter, 123.0)

	assert.Equal(t, 123.0, value.Path("$").Raw())
	value.chain.assert(t, success)
}

func TestNumber_Schema(t *testing.T) {
	reporter := newMockReporter(t)

	NewNumber(reporter, 123.0).Schema(`{"type": "number"}`).
		chain.assert(t, success)

	NewNumber(reporter, 123.0).Schema(`{"type": "object"}`).
		chain.assert(t, failure)
}

func TestNumber_IsEqual(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		cases := []struct {
			name      string
			number    float64
			value     interface{}
			wantEqual chainResult
		}{
			{
				name:      "compare equivalent integers",
				number:    1234,
				value:     1234,
				wantEqual: success,
			},
			{
				name:      "compare non-equivalent integers",
				number:    1234,
				value:     4321,
				wantEqual: failure,
			},
			{
				name:      "compare NaN to float",
				number:    math.NaN(),
				value:     1234.5,
				wantEqual: failure,
			},
			{
				name:      "compare float to NaN",
				number:    1234.5,
				value:     math.NaN(),
				wantEqual: failure,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewNumber(reporter, tc.number).IsEqual(tc.value).
					chain.assert(t, tc.wantEqual)

				NewNumber(reporter, tc.number).NotEqual(tc.value).
					chain.assert(t, !tc.wantEqual)
			})
		}
	})

	t.Run("canonization", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewNumber(reporter, 1234).IsEqual(int64(1234)).
			chain.assert(t, success)

		NewNumber(reporter, 1234).IsEqual(float32(1234)).
			chain.assert(t, success)

		NewNumber(reporter, 1234).NotEqual(int64(4321)).
			chain.assert(t, success)

		NewNumber(reporter, 1234).NotEqual(float32(4321)).
			chain.assert(t, success)
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewNumber(reporter, 1234).IsEqual("NOT NUMBER").
			chain.assert(t, failure)

		NewNumber(reporter, 1234).NotEqual("NOT NUMBER").
			chain.assert(t, failure)
	})
}

func TestNumber_InDelta(t *testing.T) {
	cases := []struct {
		name           string
		number         float64
		value          float64
		delta          float64
		wantInDelta    chainResult
		wantNotInDelta chainResult
	}{
		{
			name:           "larger value in delta range",
			number:         1234.5,
			value:          1234.7,
			delta:          0.3,
			wantInDelta:    success,
			wantNotInDelta: failure,
		},
		{
			name:           "smaller value in delta range",
			number:         1234.5,
			value:          1234.3,
			delta:          0.3,
			wantInDelta:    success,
			wantNotInDelta: failure,
		},
		{
			name:           "larger value not in delta range",
			number:         1234.5,
			value:          1234.7,
			delta:          0.1,
			wantInDelta:    failure,
			wantNotInDelta: success,
		},
		{
			name:           "smaller value not in delta range",
			number:         1234.5,
			value:          1234.3,
			delta:          0.1,
			wantInDelta:    failure,
			wantNotInDelta: success,
		},
		{
			name:           "number is NaN",
			number:         math.NaN(),
			value:          1234.0,
			delta:          0.1,
			wantInDelta:    failure,
			wantNotInDelta: failure,
		},
		{
			name:           "value is NaN",
			number:         1234.5,
			value:          math.NaN(),
			delta:          0.1,
			wantInDelta:    failure,
			wantNotInDelta: failure,
		},
		{
			name:           "delta is NaN",
			number:         1234.5,
			value:          1234.0,
			delta:          math.NaN(),
			wantInDelta:    failure,
			wantNotInDelta: failure,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewNumber(reporter, tc.number).InDelta(tc.value, tc.delta).
				chain.assert(t, tc.wantInDelta)

			NewNumber(reporter, tc.number).NotInDelta(tc.value, tc.delta).
				chain.assert(t, tc.wantNotInDelta)
		})
	}
}

func TestNumber_InDeltaRelative(t *testing.T) {
	cases := []struct {
		name             string
		number           float64
		value            float64
		delta            float64
		expectInDelta    bool
		expectNotInDelta bool
	}{
		{
			name:             "larger value in delta range",
			number:           1234.5,
			value:            1271.5,
			delta:            0.03,
			expectInDelta:    true,
			expectNotInDelta: false,
		},
		{
			name:             "smaller value in delta range",
			number:           1234.5,
			value:            1221.1,
			delta:            0.03,
			expectInDelta:    true,
			expectNotInDelta: false,
		},
		{
			name:             "larger value not in delta range",
			number:           1234.5,
			value:            1259.1,
			delta:            0.01,
			expectInDelta:    false,
			expectNotInDelta: true,
		},
		{
			name:             "smaller value not in delta range",
			number:           1234.5,
			value:            1209.8,
			delta:            0.01,
			expectInDelta:    false,
			expectNotInDelta: true,
		},
		{
			name:             "delta is negative",
			number:           1234.5,
			value:            1234.0,
			delta:            -0.01,
			expectInDelta:    false,
			expectNotInDelta: false,
		},
		{
			name:             "target is NaN",
			number:           math.NaN(),
			value:            1234.0,
			delta:            0.01,
			expectInDelta:    false,
			expectNotInDelta: false,
		},
		{
			name:             "value is NaN",
			number:           1234.5,
			value:            math.NaN(),
			delta:            0.01,
			expectInDelta:    false,
			expectNotInDelta: false,
		},
		{
			name:             "delta is NaN",
			number:           1234.5,
			value:            1234.0,
			delta:            math.NaN(),
			expectInDelta:    false,
			expectNotInDelta: false,
		},
		{
			name:             "+Inf delta in range",
			number:           1234.5,
			value:            1234.0,
			delta:            math.Inf(1),
			expectInDelta:    true,
			expectNotInDelta: false,
		},
		{
			name:             "+Inf target",
			number:           math.Inf(1),
			value:            1234.0,
			delta:            0,
			expectInDelta:    false,
			expectNotInDelta: false,
		},
		{
			name:             "-Inf value",
			number:           1234.5,
			value:            math.Inf(-1),
			delta:            0.01,
			expectInDelta:    false,
			expectNotInDelta: true,
		},
		{
			name:             "target is 0 in delta range",
			number:           0,
			value:            0,
			delta:            0,
			expectInDelta:    true,
			expectNotInDelta: false,
		},
		{
			name:             "value is 0 in delta range",
			number:           0.05,
			value:            0,
			delta:            1.0,
			expectInDelta:    true,
			expectNotInDelta: false,
		},
		{
			name:             "value is 0 not in delta range",
			number:           0.01,
			value:            0,
			delta:            0.01,
			expectInDelta:    false,
			expectNotInDelta: true,
		},
	}

	for name, instance := range cases {
		t.Run(name, func(t *testing.T) {
			reporter := newMockReporter(t)

			if instance.expectInDelta {
				NewNumber(reporter, instance.number).
					InDeltaRelative(instance.value, instance.delta).
					chain.assertNotFailed(t)
			} else {
				NewNumber(reporter, instance.number).
					InDeltaRelative(instance.value, instance.delta).
					chain.assertFailed(t)
			}

			if instance.expectNotInDelta {
				NewNumber(reporter, instance.number).
					NotInDeltaRelative(instance.value, instance.delta).
					chain.assertNotFailed(t)
			} else {
				NewNumber(reporter, instance.number).
					NotInDeltaRelative(instance.value, instance.delta).
					chain.assertFailed(t)
			}
		})
	}
}

func TestNumber_InRange(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		cases := []struct {
			name           string
			number         float64
			min            interface{}
			max            interface{}
			wantInRange    chainResult
			wantNotInRange chainResult
		}{
			{
				name:           "range includes only number",
				number:         1234,
				min:            1234,
				max:            1234,
				wantInRange:    success,
				wantNotInRange: failure,
			},
			{
				name:           "range includes number and below",
				number:         1234,
				min:            1234 - 1,
				max:            1234,
				wantInRange:    success,
				wantNotInRange: failure,
			},
			{
				name:           "range includes number and above",
				number:         1234,
				min:            1234,
				max:            1234 + 1,
				wantInRange:    success,
				wantNotInRange: failure,
			},
			{
				name:           "range is above number",
				number:         1234,
				min:            1234 + 1,
				max:            1234 + 2,
				wantInRange:    failure,
				wantNotInRange: success,
			},
			{
				name:           "range is below number",
				number:         1234,
				min:            1234 - 2,
				max:            1234 - 1,
				wantInRange:    failure,
				wantNotInRange: success,
			},
			{
				name:           "range min is larger than max",
				number:         1234,
				min:            1234 + 1,
				max:            1234 - 1,
				wantInRange:    failure,
				wantNotInRange: success,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewNumber(reporter, tc.number).InRange(tc.min, tc.max).
					chain.assert(t, tc.wantInRange)

				NewNumber(reporter, tc.number).NotInRange(tc.min, tc.max).
					chain.assert(t, tc.wantNotInRange)
			})
		}
	})

	t.Run("canonization", func(t *testing.T) {
		cases := []struct {
			number      float64
			min         interface{}
			max         interface{}
			wantInRange chainResult
		}{
			{
				number:      1234,
				min:         int64(1233),
				max:         float32(1235),
				wantInRange: success,
			},
			{
				number:      1234,
				min:         1235,
				max:         1236,
				wantInRange: failure,
			},
		}

		reporter := newMockReporter(t)

		for _, tc := range cases {
			NewNumber(reporter, tc.number).InRange(tc.min, tc.max).
				chain.assert(t, tc.wantInRange)

			NewNumber(reporter, tc.number).NotInRange(tc.min, tc.max).
				chain.assert(t, !tc.wantInRange)
		}
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewNumber(reporter, 1234).InRange(int64(1233), "NOT NUMBER").
			chain.assert(t, failure)

		NewNumber(reporter, 1234).NotInRange(int64(1233), "NOT NUMBER").
			chain.assert(t, failure)

		NewNumber(reporter, 1234).InRange("NOT NUMBER", float32(1235)).
			chain.assert(t, failure)

		NewNumber(reporter, 1234).NotInRange("NOT NUMBER", float32(1235)).
			chain.assert(t, failure)
	})
}

func TestNumber_InList(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		cases := []struct {
			name          string
			number        float64
			list          []interface{}
			wantInList    chainResult
			wantNotInList chainResult
		}{
			{
				name:          "no list",
				number:        1234,
				list:          nil,
				wantInList:    failure,
				wantNotInList: failure,
			},
			{
				name:          "empty list",
				number:        1234,
				list:          []interface{}{},
				wantInList:    failure,
				wantNotInList: failure,
			},
			{
				name:          "in integer list",
				number:        1234,
				list:          []interface{}{1234, 4567},
				wantInList:    success,
				wantNotInList: failure,
			},
			{
				name:          "in float list",
				number:        1234,
				list:          []interface{}{1234.00, 4567.00},
				wantInList:    success,
				wantNotInList: failure,
			},
			{
				name:          "not in float list",
				number:        1234,
				list:          []interface{}{4567.00, 1234.01},
				wantInList:    failure,
				wantNotInList: success,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewNumber(reporter, tc.number).InList(tc.list...).
					chain.assert(t, tc.wantInList)

				NewNumber(reporter, tc.number).NotInList(tc.list...).
					chain.assert(t, tc.wantNotInList)
			})
		}
	})

	t.Run("canonization", func(t *testing.T) {
		cases := []struct {
			number     float64
			list       []interface{}
			wantInList chainResult
		}{
			{
				number:     111,
				list:       []interface{}{int64(111), float32(222)},
				wantInList: success,
			},
			{
				number:     111,
				list:       []interface{}{float32(111), int64(222)},
				wantInList: success,
			},
			{
				number:     111,
				list:       []interface{}{222, 333},
				wantInList: failure,
			},
		}

		reporter := newMockReporter(t)

		for _, tc := range cases {
			NewNumber(reporter, tc.number).InList(tc.list...).
				chain.assert(t, tc.wantInList)

			NewNumber(reporter, tc.number).NotInList(tc.list...).
				chain.assert(t, !tc.wantInList)
		}
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewNumber(reporter, 111).InList(222, "NOT NUMBER").
			chain.assert(t, failure)

		NewNumber(reporter, 111).NotInList(222, "NOT NUMBER").
			chain.assert(t, failure)

		NewNumber(reporter, 111).InList(111, "NOT NUMBER").
			chain.assert(t, failure)

		NewNumber(reporter, 111).NotInList(111, "NOT NUMBER").
			chain.assert(t, failure)
	})
}

func TestNumber_IsGreater(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		cases := []struct {
			name   string
			number float64
			value  interface{}
			wantGt chainResult
			wantGe chainResult
		}{
			{
				name:   "number is lesser",
				number: 1234,
				value:  1234 + 1,
				wantGt: failure,
				wantGe: failure,
			},
			{
				name:   "number is equal",
				number: 1234,
				value:  1234,
				wantGt: failure,
				wantGe: success,
			},
			{
				name:   "number is greater",
				number: 1234,
				value:  1234 - 1,
				wantGt: success,
				wantGe: success,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewNumber(reporter, tc.number).Gt(tc.value).
					chain.assert(t, tc.wantGt)

				NewNumber(reporter, tc.number).Ge(tc.value).
					chain.assert(t, tc.wantGe)
			})
		}
	})

	t.Run("canonization", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewNumber(reporter, 1234).Gt(int64(1233)).
			chain.assert(t, success)

		NewNumber(reporter, 1234).Gt(float32(1233)).
			chain.assert(t, success)

		NewNumber(reporter, 1234).Ge(int64(1233)).
			chain.assert(t, success)

		NewNumber(reporter, 1234).Ge(float32(1233)).
			chain.assert(t, success)
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewNumber(reporter, 1234).Gt("NOT NUMBER").
			chain.assert(t, failure)

		NewNumber(reporter, 1234).Ge("NOT NUMBER").
			chain.assert(t, failure)
	})
}

func TestNumber_IsLesser(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		cases := []struct {
			name   string
			number float64
			value  interface{}
			wantLt chainResult
			wantLe chainResult
		}{
			{
				name:   "number is lesser",
				number: 1234,
				value:  1234 + 1,
				wantLt: success,
				wantLe: success,
			},
			{
				name:   "number is equal",
				number: 1234,
				value:  1234,
				wantLt: failure,
				wantLe: success,
			},
			{
				name:   "number is greater",
				number: 1234,
				value:  1234 - 1,
				wantLt: failure,
				wantLe: failure,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewNumber(reporter, tc.number).Lt(tc.value).
					chain.assert(t, tc.wantLt)

				NewNumber(reporter, tc.number).Le(tc.value).
					chain.assert(t, tc.wantLe)
			})
		}
	})

	t.Run("canonization", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewNumber(reporter, 1234).Lt(int64(1235)).
			chain.assert(t, success)

		NewNumber(reporter, 1234).Lt(float32(1235)).
			chain.assert(t, success)

		NewNumber(reporter, 1234).Le(int64(1235)).
			chain.assert(t, success)

		NewNumber(reporter, 1234).Le(float32(1235)).
			chain.assert(t, success)
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewNumber(reporter, 1234).Lt("NOT NUMBER").
			chain.assert(t, failure)

		NewNumber(reporter, 1234).Le("NOT NUMBER").
			chain.assert(t, failure)
	})
}

func TestNumber_IsInt(t *testing.T) {
	t.Run("values", func(t *testing.T) {
		cases := []struct {
			name      string
			value     float64
			wantInt16 chainResult
			wantInt32 chainResult
			wantInt   chainResult
		}{
			{
				name:      "0",
				value:     0,
				wantInt16: success,
				wantInt32: success,
				wantInt:   success,
			},
			{
				name:      "1",
				value:     1,
				wantInt16: success,
				wantInt32: success,
				wantInt:   success,
			},
			{
				name:      "0.5",
				value:     0.5,
				wantInt16: failure,
				wantInt32: failure,
				wantInt:   failure,
			},
			{
				name:      "NaN",
				value:     math.NaN(),
				wantInt16: failure,
				wantInt32: failure,
				wantInt:   failure,
			},
			{
				name:      "-Inf",
				value:     math.Inf(-1),
				wantInt16: failure,
				wantInt32: failure,
				wantInt:   failure,
			},
			{
				name:      "+Inf",
				value:     math.Inf(+1),
				wantInt16: failure,
				wantInt32: failure,
				wantInt:   failure,
			},
			{
				name:      "MinInt16-1",
				value:     math.MinInt16 - 1,
				wantInt16: failure,
				wantInt32: success,
				wantInt:   success,
			},
			{
				name:      "MinInt16",
				value:     math.MinInt16,
				wantInt16: success,
				wantInt32: success,
				wantInt:   success,
			},
			{
				name:      "MaxInt16",
				value:     math.MaxInt16,
				wantInt16: success,
				wantInt32: success,
				wantInt:   success,
			},
			{
				name:      "MaxInt16+1",
				value:     math.MaxInt16 + 1,
				wantInt16: failure,
				wantInt32: success,
				wantInt:   success,
			},
			{
				name:      "MinInt32-1",
				value:     math.MinInt32 - 1,
				wantInt16: failure,
				wantInt32: failure,
				wantInt:   success,
			},
			{
				name:      "MinInt32",
				value:     math.MinInt32,
				wantInt16: failure,
				wantInt32: success,
				wantInt:   success,
			},
			{
				name:      "MaxInt32",
				value:     math.MaxInt32,
				wantInt16: failure,
				wantInt32: success,
				wantInt:   success,
			},
			{
				name:      "MaxInt32+1",
				value:     math.MaxInt32 + 1,
				wantInt16: failure,
				wantInt32: failure,
				wantInt:   success,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewNumber(reporter, tc.value).IsInt().
					chain.assert(t, tc.wantInt)
				NewNumber(reporter, tc.value).NotInt().
					chain.assert(t, !tc.wantInt)

				NewNumber(reporter, tc.value).IsInt(32).
					chain.assert(t, tc.wantInt32)
				NewNumber(reporter, tc.value).NotInt(32).
					chain.assert(t, !tc.wantInt32)

				NewNumber(reporter, tc.value).IsInt(16).
					chain.assert(t, tc.wantInt16)
				NewNumber(reporter, tc.value).NotInt(16).
					chain.assert(t, !tc.wantInt16)
			})
		}
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewNumber(reporter, 1234).IsInt(16, 32).
			chain.assert(t, failure)
		NewNumber(reporter, 1234).NotInt(16, 32).
			chain.assert(t, failure)

		NewNumber(reporter, 1234).IsInt(0).
			chain.assert(t, failure)
		NewNumber(reporter, 1234).NotInt(0).
			chain.assert(t, failure)

		NewNumber(reporter, 1234).IsInt(-16).
			chain.assert(t, failure)
		NewNumber(reporter, 1234).NotInt(-16).
			chain.assert(t, failure)
	})
}

func TestNumber_IsUint(t *testing.T) {
	t.Run("values", func(t *testing.T) {
		cases := []struct {
			name       string
			value      float64
			wantUint16 chainResult
			wantUint32 chainResult
			wantUint   chainResult
		}{
			{
				name:       "0",
				value:      0,
				wantUint16: success,
				wantUint32: success,
				wantUint:   success,
			},
			{
				name:       "1",
				value:      1,
				wantUint16: success,
				wantUint32: success,
				wantUint:   success,
			},
			{
				name:       "-1",
				value:      -1,
				wantUint16: failure,
				wantUint32: failure,
				wantUint:   failure,
			},
			{
				name:       "0.5",
				value:      0.5,
				wantUint16: failure,
				wantUint32: failure,
				wantUint:   failure,
			},
			{
				name:       "NaN",
				value:      math.NaN(),
				wantUint16: failure,
				wantUint32: failure,
				wantUint:   failure,
			},
			{
				name:       "-Inf",
				value:      math.Inf(-1),
				wantUint16: failure,
				wantUint32: failure,
				wantUint:   failure,
			},
			{
				name:       "+Inf",
				value:      math.Inf(+1),
				wantUint16: failure,
				wantUint32: failure,
				wantUint:   failure,
			},
			{
				name:       "MaxUint16",
				value:      math.MaxUint16,
				wantUint16: success,
				wantUint32: success,
				wantUint:   success,
			},
			{
				name:       "MaxUint16+1",
				value:      math.MaxUint16 + 1,
				wantUint16: failure,
				wantUint32: success,
				wantUint:   success,
			},
			{
				name:       "MaxUint32",
				value:      math.MaxUint32,
				wantUint16: failure,
				wantUint32: success,
				wantUint:   success,
			},
			{
				name:       "MaxUint32+1",
				value:      math.MaxUint32 + 1,
				wantUint16: failure,
				wantUint32: failure,
				wantUint:   success,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				NewNumber(reporter, tc.value).IsUint().
					chain.assert(t, tc.wantUint)
				NewNumber(reporter, tc.value).NotUint().
					chain.assert(t, !tc.wantUint)

				NewNumber(reporter, tc.value).IsUint(32).
					chain.assert(t, tc.wantUint32)
				NewNumber(reporter, tc.value).NotUint(32).
					chain.assert(t, !tc.wantUint32)

				NewNumber(reporter, tc.value).IsUint(16).
					chain.assert(t, tc.wantUint16)
				NewNumber(reporter, tc.value).NotUint(16).
					chain.assert(t, !tc.wantUint16)
			})
		}
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewNumber(reporter, 1234).IsUint(16, 32).
			chain.assert(t, failure)
		NewNumber(reporter, 1234).NotUint(16, 32).
			chain.assert(t, failure)

		NewNumber(reporter, 1234).IsUint(0).
			chain.assert(t, failure)
		NewNumber(reporter, 1234).NotUint(0).
			chain.assert(t, failure)

		NewNumber(reporter, 1234).IsUint(-16).
			chain.assert(t, failure)
		NewNumber(reporter, 1234).NotUint(-16).
			chain.assert(t, failure)
	})
}

func TestNumber_IsFinite(t *testing.T) {
	cases := []struct {
		name       string
		value      float64
		wantFinite chainResult
	}{
		{
			name:       "0",
			value:      0,
			wantFinite: success,
		},
		{
			name:       "1",
			value:      1,
			wantFinite: success,
		},
		{
			name:       "-1",
			value:      -1,
			wantFinite: success,
		},
		{
			name:       "0.5",
			value:      0.5,
			wantFinite: success,
		},
		{
			name:       "NaN",
			value:      math.NaN(),
			wantFinite: failure,
		},
		{
			name:       "-Inf",
			value:      math.Inf(-1),
			wantFinite: failure,
		},
		{
			name:       "+Inf",
			value:      math.Inf(+1),
			wantFinite: failure,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewNumber(reporter, tc.value).IsFinite().
				chain.assert(t, tc.wantFinite)

			NewNumber(reporter, tc.value).NotFinite().
				chain.assert(t, !tc.wantFinite)
		})
	}
}
