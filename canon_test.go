package httpexpect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type myInt int
type myArray []interface{}

type inputStruct struct {
	chain *chain
	in    interface{}
}
type outputStruct struct {
	val interface{}
	ok  bool
}

type TestCase struct {
	description string
	input       inputStruct
	output      outputStruct
}

func TestCanon_Number(t *testing.T) {

	chain := newMockChain(t).enter("test")
	defer chain.leave()

	for _, testCase := range []TestCase{
		{
			description: "testCanon_Number: input '123' as int",
			input:       inputStruct{chain: chain, in: 123},
			output:      outputStruct{val: 123.0, ok: true},
		},
		{
			description: "testCanon_Number: input '123.0' as float",
			input:       inputStruct{chain: chain, in: 123.0},
			output:      outputStruct{val: 123.0, ok: true},
		},
		{
			description: "testCanon_Number: input '123' as myInt",
			input:       inputStruct{chain: chain, in: myInt(123)},
			output:      outputStruct{val: 123.0, ok: true},
		},
		{
			description: "testCanon_Number: input '123' as string",
			input:       inputStruct{chain: chain, in: "123"},
			output:      outputStruct{ok: false},
		},
		{
			description: "testCanon_Number: input is nil",
			input:       inputStruct{chain: chain, in: nil},
			output:      outputStruct{ok: false},
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			val, ok := canonNumber(testCase.input.chain, testCase.input.in)
			assert := assert.New(t)
			assert.Equal(testCase.output.ok, ok)
			if testCase.output.ok {
				assert.Equal(testCase.output.val, val)
				chain.assert(t, success)
			} else {
				chain.assert(t, failure)
			}
			chain.clear()
		})
	}
}

func TestCannon_Array(t *testing.T) {
	chain := newMockChain(t).enter("test")
	defer chain.leave()

	for _, testCase := range []TestCase{
		{
			description: "testCanon_Array: input []interface{}{123.0, 456.0}",
			input:       inputStruct{chain: chain, in: []interface{}{123.0, 456.0}},
			output:      outputStruct{val: []interface{}{123.0, 456.0}, ok: true},
		},
		{
			description: "testCanon_Array: input myArray{myInt(123), 456.0}",
			input:       inputStruct{chain: chain, in: myArray{myInt(123), 456.0}},
			output:      outputStruct{val: []interface{}{123.0, 456.0}, ok: true},
		},
		{
			description: "testCanon_Array: input '123'",
			input:       inputStruct{chain: chain, in: "123"},
			output:      outputStruct{ok: false},
		},
		{
			description: "testCanon_Array: input empty function",
			input:       inputStruct{chain: chain, in: func() {}},
			output:      outputStruct{ok: false},
		},
		{
			description: "testCanon_Array: input is nil",
			input:       inputStruct{chain: chain, in: nil},
			output:      outputStruct{ok: false},
		},
		{
			description: "testCanon_Array: input is []interface{}(nil)",
			input:       inputStruct{chain: chain, in: nil},
			output:      outputStruct{ok: false},
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			val, ok := canonArray(testCase.input.chain, testCase.input.in)
			assert := assert.New(t)
			assert.Equal(testCase.output.ok, ok)
			if testCase.output.ok {
				assert.Equal(testCase.output.val, val)
				chain.assert(t, success)
			} else {
				chain.assert(t, failure)
			}
			chain.clear()
		})
	}
}

func TestCanon_Map(t *testing.T) {
	chain := newMockChain(t).enter("test")
	defer chain.leave()

	for _, testCase := range []TestCase{
		{
			description: "testCanon_Map: input is map[string]interface{}{'foo': 123.0}",
			input:       inputStruct{chain: chain, in: map[string]interface{}{"foo": 123.0}},
			output:      outputStruct{val: map[string]interface{}{"foo": 123.0}, ok: true},
		},
		{
			description: "testCanon_Map: input is myMap{'foo': myInt(123)}",
			input:       inputStruct{chain: chain, in: map[string]interface{}{"foo": myInt(123)}},
			output:      outputStruct{val: map[string]interface{}{"foo": 123.0}, ok: true},
		},
		{
			description: "testCanon_Map: input is '123'",
			input:       inputStruct{chain: chain, in: "123"},
			output:      outputStruct{ok: false},
		},
		{
			description: "testCanon_Map: input is func() {}",
			input:       inputStruct{chain: chain, in: func() {}},
			output:      outputStruct{ok: false},
		},
		{
			description: "testCanon_Map: input is nil",
			input:       inputStruct{chain: chain, in: nil},
			output:      outputStruct{ok: false},
		},
		{
			description: "testCanon_Map: input is map[string]interface{}(nil)",
			input:       inputStruct{chain: chain, in: map[string]interface{}(nil)},
			output:      outputStruct{ok: false},
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			val, ok := canonMap(testCase.input.chain, testCase.input.in)
			assert := assert.New(t)
			assert.Equal(testCase.output.ok, ok)
			if testCase.output.ok {
				assert.Equal(testCase.output.val, val)
				chain.assert(t, success)
			} else {
				chain.assert(t, failure)
			}
			chain.clear()
		})
	}
}

func TestCannon_Decode(t *testing.T) {
	type decodeTestCase struct {
		chain  *chain
		value  interface{}
		target interface{}
	}

	chain := newMockChain(t).enter("test")
	defer chain.leave()

	type s struct {
		MyFunc func() string
	}

	var target s
	var targetInt int
	for _, testCase := range []decodeTestCase{
		{
			chain: chain, value: 123, target: nil,
		},
		{
			chain: chain, value: &s{MyFunc: func() string { return "foo" }}, target: &target,
		},
		{
			chain: chain, value: true, target: targetInt,
		},
	} {
		t.Run("value is not unmarshallable into target", func(t *testing.T) {
			canonDecode(testCase.chain, testCase.value, testCase.target)
		})
		chain.assert(t, failure)
	}
}
