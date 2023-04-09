package httpexpect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanon_Number(t *testing.T) {

	type myInt int

	for _, tc := range []struct {
		name   string
		in     interface{}
		val    interface{}
		result chainResult
	}{
		{
			name:   "testCanon_Number: input '123' as int",
			in:     123,
			val:    123.0,
			result: true,
		},
		{
			name:   "testCanon_Number: input '123.0' as float",
			in:     123.0,
			val:    123.0,
			result: true,
		},
		{
			name:   "testCanon_Number: input '123' as myInt",
			in:     myInt(123),
			val:    123.0,
			result: true,
		},
		{
			name:   "testCanon_Number: input '123' as string",
			in:     "123",
			result: false,
		},
		{
			name:   "testCanon_Number: input is nil",
			in:     nil,
			result: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chain := newMockChain(t).enter("test")
			defer chain.leave()
			val, ok := canonNumber(chain, tc.in)
			assert := assert.New(t)
			assert.Equal(tc.result, chainResult(ok))
			chain.assert(t, tc.result)
			if tc.result {
				assert.Equal(tc.val, val)
			}
			chain.clear()
		})
	}
}

func TestCannon_Array(t *testing.T) {

	type (
		myInt   int
		myArray []interface{}
	)

	for _, tc := range []struct {
		name   string
		in     interface{}
		val    interface{}
		result chainResult
	}{
		{
			name:   "testCanon_Array: input []interface{}{123.0, 456.0}",
			in:     []interface{}{123.0, 456.0},
			val:    []interface{}{123.0, 456.0},
			result: true,
		},
		{
			name:   "testCanon_Array: input myArray{myInt(123), 456.0}",
			in:     myArray{myInt(123), 456.0},
			val:    []interface{}{123.0, 456.0},
			result: true,
		},
		{
			name:   "testCanon_Array: input '123'",
			in:     "123",
			result: false,
		},
		{
			name:   "testCanon_Array: input empty function",
			in:     func() {},
			result: false,
		},
		{
			name:   "testCanon_Array: input is nil",
			in:     nil,
			result: false,
		},
		{
			name:   "testCanon_Array: input is []interface{}(nil)",
			in:     []interface{}(nil),
			result: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chain := newMockChain(t).enter("test")
			defer chain.leave()
			val, ok := canonArray(chain, tc.in)
			assert := assert.New(t)
			assert.Equal(tc.result, chainResult(ok))
			chain.assert(t, tc.result)
			if tc.result {
				assert.Equal(tc.val, val)
			}
			chain.clear()
		})
	}
}

func TestCanon_Map(t *testing.T) {

	type (
		myInt int
		myMap map[string]interface{}
	)

	for _, tc := range []struct {
		name   string
		in     interface{}
		val    interface{}
		result chainResult
	}{
		{
			name:   "testCanon_Map: input is map[string]interface{}{'foo': 123.0}",
			in:     map[string]interface{}{"foo": 123.0},
			val:    map[string]interface{}{"foo": 123.0},
			result: true,
		},
		{
			name:   "testCanon_Map: input is myMap{'foo': myInt(123)}",
			in:     myMap{"foo": myInt(123)},
			val:    map[string]interface{}{"foo": 123.0},
			result: true,
		},
		{
			name:   "testCanon_Map: input is '123'",
			in:     "123",
			result: false,
		},
		{
			name:   "testCanon_Map: input is func() {}",
			in:     func() {},
			result: false,
		},
		{
			name:   "testCanon_Map: input is nil",
			in:     nil,
			result: false,
		},
		{
			name:   "testCanon_Map: input is map[string]interface{}(nil)",
			in:     map[string]interface{}(nil),
			result: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chain := newMockChain(t).enter("test")
			defer chain.leave()
			val, ok := canonMap(chain, tc.in)
			assert := assert.New(t)
			assert.Equal(tc.result, chainResult(ok))
			chain.assert(t, tc.result)
			if tc.result {
				assert.Equal(tc.val, val)
			}
			chain.clear()
		})
	}
}

func TestCannon_Decode(t *testing.T) {

	type S struct {
		MyFunc func() string
	}
	var (
		target    S
		targetInt int
	)

	for _, tc := range []struct {
		name   string
		value  interface{}
		target interface{}
	}{
		{
			name:   "target is nil",
			value:  123,
			target: nil,
		},
		{
			name:   "value is not marshallable",
			value:  &S{MyFunc: func() string { return "foo" }},
			target: &target,
		},
		{
			name:   "value is not unmarshallable into target",
			value:  true,
			target: targetInt,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chain := newMockChain(t).enter("test")
			defer chain.leave()
			canonDecode(chain, tc.value, tc.target)
			chain.assert(t, failure)
		})
	}
}
