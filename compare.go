package httpexpect

import (
	"errors"
	"fmt"
)

func arrayComparator(opChain *chain, array []interface{}) func(x, y *Value) bool {
	if !arrayComparatorValidate(opChain, array) {
		return nil
	}
	return arrayComparatorConstruct(array)
}

func arrayComparatorValidate(opChain *chain, array []interface{}) bool {
	var prev interface{}
	for index, curr := range array {
		switch curr.(type) {
		case bool, float64, string, nil:
			// ok, do nothing

		default:
			opChain.fail(AssertionFailure{
				Type: AssertBelongs,
				Actual: &AssertionValue{
					typeName(fmt.Sprintf("%T", curr)),
				},
				Expected: &AssertionValue{AssertionList{
					typeName("Boolean (bool)"),
					typeName("Number (int*, uint*, float*)"),
					typeName("String (string)"),
					typeName("Null (nil)"),
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
			return false
		}

		if index > 0 && fmt.Sprintf("%T", curr) != fmt.Sprintf("%T", prev) {
			opChain.fail(AssertionFailure{
				Type: AssertEqual,
				Actual: &AssertionValue{
					typeName(fmt.Sprintf("%T (type of element %v)", curr, index)),
				},
				Expected: &AssertionValue{
					typeName(fmt.Sprintf("%T (type of element %v)", prev, index-1)),
				},
				Reference: &AssertionValue{
					array,
				},
				Errors: []error{
					errors.New("expected: types of all elements of reference array are the same"),
					fmt.Errorf("element %v has type %T, but element %v has type %T",
						index-1, prev, index, curr),
				},
			})
			return false
		}

		prev = curr
	}

	return true
}

func arrayComparatorConstruct(array []interface{}) func(x, y *Value) bool {
	if len(array) > 0 {
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
	}

	return nil
}

type typeName string

func (t typeName) String() string {
	return string(t)
}
