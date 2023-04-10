package httpexpect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanon_Number(t *testing.T) {
	type myInt int

	cases := []struct {
		name   string
		in     interface{}
		out    interface{}
		result chainResult
	}{
		{
			name:   "input is int",
			in:     123,
			out:    123.0,
			result: success,
		},
		{
			name:   "input is float",
			in:     123.0,
			out:    123.0,
			result: success,
		},
		{
			name:   "input is myInt",
			in:     myInt(123),
			out:    123.0,
			result: success,
		},
		{
			name:   "input is string",
			in:     "123",
			result: failure,
		},
		{
			name:   "input is nil",
			in:     nil,
			result: failure,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			chain := newMockChain(t).enter("test")
			defer chain.leave()

			val, ok := canonNumber(chain, tc.in)
			assert.Equal(t, tc.result, chainResult(ok))

			chain.assert(t, tc.result)
			if tc.result {
				assert.Equal(t, tc.out, val)
			}
		})
	}
}

func TestCannon_Array(t *testing.T) {
	type (
		myInt   int
		myArray []interface{}
	)

	cases := []struct {
		name   string
		in     interface{}
		out    interface{}
		result chainResult
	}{
		{
			name:   "input is []interface{}",
			in:     []interface{}{123.0, 456.0},
			out:    []interface{}{123.0, 456.0},
			result: success,
		},
		{
			name:   "input is myArray{}",
			in:     myArray{myInt(123), 456.0},
			out:    []interface{}{123.0, 456.0},
			result: success,
		},
		{
			name:   "input is string",
			in:     "123",
			result: failure,
		},
		{
			name:   "input is func() {}",
			in:     func() {},
			result: failure,
		},
		{
			name:   "input is nil",
			in:     nil,
			result: failure,
		},
		{
			name:   "input is []interface{}(nil)",
			in:     []interface{}(nil),
			result: failure,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			chain := newMockChain(t).enter("test")
			defer chain.leave()

			val, ok := canonArray(chain, tc.in)
			assert.Equal(t, tc.result, chainResult(ok))

			chain.assert(t, tc.result)
			if tc.result {
				assert.Equal(t, tc.out, val)
			}
		})
	}
}

func TestCanon_Map(t *testing.T) {
	type (
		myInt int
		myMap map[string]interface{}
	)

	cases := []struct {
		name   string
		in     interface{}
		out    interface{}
		result chainResult
	}{
		{
			name:   "input is map[string]interface{}{}",
			in:     map[string]interface{}{"foo": 123.0},
			out:    map[string]interface{}{"foo": 123.0},
			result: success,
		},
		{
			name:   "input is myMap{}",
			in:     myMap{"foo": myInt(123)},
			out:    map[string]interface{}{"foo": 123.0},
			result: success,
		},
		{
			name:   "input is string",
			in:     "123",
			result: failure,
		},
		{
			name:   "input is func() {}",
			in:     func() {},
			result: failure,
		},
		{
			name:   "input is nil",
			in:     nil,
			result: failure,
		},
		{
			name:   "input is map[string]interface{}(nil)",
			in:     map[string]interface{}(nil),
			result: failure,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			chain := newMockChain(t).enter("test")
			defer chain.leave()

			val, ok := canonMap(chain, tc.in)
			assert.Equal(t, tc.result, chainResult(ok))

			chain.assert(t, tc.result)
			if tc.result {
				assert.Equal(t, tc.out, val)
			}
		})
	}
}

func TestCannon_Decode(t *testing.T) {
	cases := []struct {
		name   string
		value  interface{}
		target interface{}
		result chainResult
	}{
		{
			name:   "string",
			value:  "foo",
			target: new(string),
			result: success,
		},
		{
			name:   "int",
			value:  123,
			target: new(int),
			result: success,
		},
		{
			name:   "any",
			value:  123,
			target: new(interface{}),
			result: success,
		},
		{
			name:   "target is nil",
			value:  123,
			target: nil,
			result: failure,
		},
		{
			name:   "unmarshallable value",
			value:  func() {},
			target: new(interface{}),
			result: failure,
		},
		{
			name:   "incompatible value",
			value:  "foo",
			target: new(int),
			result: failure,
		},
		{
			name:   "target is not a pointer",
			value:  123,
			target: 123,
			result: failure,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			chain := newMockChain(t).enter("test")
			defer chain.leave()

			canonDecode(chain, tc.value, tc.target)
			chain.assert(t, tc.result)
		})
	}
}
