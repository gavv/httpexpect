package httpexpect

import (
	"errors"
	"fmt"
)

func arrayComparator(opChain *chain, array []interface{}) func(x, y *Value) bool {
	var prev interface{}
	for index, curr := range array {
		switch curr.(type) {
		case bool, float64, string, nil:
			// ok, do nothing

		default:
			opChain.fail(AssertionFailure{
				Type: AssertBelongs,
				Actual: &AssertionValue{
					unquotedType(fmt.Sprintf("%T", curr)),
				},
				Expected: &AssertionValue{AssertionList{
					unquotedType("Boolean (bool)"),
					unquotedType("Number (int*, uint*, float*)"),
					unquotedType("String (string)"),
					unquotedType("Null (nil)"),
				}},
				Reference: &AssertionValue{
					array,
				},
				Errors: []error{
					errors.New("expected: type of each element of reference array" +
						" belongs to allowed list"),
					fmt.Errorf("element %v has disallowed type %T", index, curr),
				},
			})
			return nil
		}

		if index > 0 && fmt.Sprintf("%T", curr) != fmt.Sprintf("%T", prev) {
			opChain.fail(AssertionFailure{
				Type: AssertEqual,
				Actual: &AssertionValue{
					unquotedType(fmt.Sprintf("%T (type of element %v)", curr, index)),
				},
				Expected: &AssertionValue{
					unquotedType(fmt.Sprintf("%T (type of element %v)", prev, index-1)),
				},
				Reference: &AssertionValue{
					array,
				},
				Errors: []error{
					errors.New("expected:" +
						" types of all elements of reference array are the same"),
					fmt.Errorf("element %v has type %T, but element %v has type %T",
						index-1, prev, index, curr),
				},
			})
			return nil
		}

		prev = curr
	}

	switch array[0].(type) {
	case bool:
		return func(x, y *Value) bool {
			xVal := x.Raw().(bool)
			yVal := y.Raw().(bool)
			return (!xVal && yVal)
		}
	case float64:
		return func(x, y *Value) bool {
			xVal := x.Raw().(float64)
			yVal := y.Raw().(float64)
			return xVal < yVal
		}
	case string:
		return func(x, y *Value) bool {
			xVal := x.Raw().(string)
			yVal := y.Raw().(string)
			return xVal < yVal
		}
	case nil:
		return func(x, y *Value) bool {
			// `nil` is never less than `nil`
			return false
		}
	}

	return nil
}

type unquotedType string

func (t unquotedType) String() string {
	return string(t)
}
