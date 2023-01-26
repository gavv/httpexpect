package httpexpect

import (
	"errors"
	"fmt"
	"reflect"
)

// Array provides methods to inspect attached []interface{} object
// (Go representation of JSON array).
type Array struct {
	noCopy noCopy
	chain  *chain
	value  []interface{}
}

// NewArray returns a new Array instance.
//
// If reporter is nil, the function panics.
// If value is nil, failure is reported.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
func NewArray(reporter Reporter, value []interface{}) *Array {
	return newArray(newChainWithDefaults("Array()", reporter), value)
}

// NewArrayC returns a new Array instance with config.
//
// Requirements for config are same as for WithConfig function.
// If value is nil, failure is reported.
//
// Example:
//
//	array := NewArrayC(config, []interface{}{"foo", 123})
func NewArrayC(config Config, value []interface{}) *Array {
	return newArray(newChainWithConfig("Array()", config.withDefaults()), value)
}

func newArray(parent *chain, val []interface{}) *Array {
	a := &Array{chain: parent.clone(), value: nil}

	opChain := a.chain.enter("")
	defer opChain.leave()

	if val == nil {
		opChain.fail(AssertionFailure{
			Type:   AssertNotNil,
			Actual: &AssertionValue{val},
			Errors: []error{
				errors.New("expected: non-nil array"),
			},
		})
	} else {
		a.value, _ = canonArray(opChain, val)
	}

	return a
}

// Raw returns underlying value attached to Array.
// This is the value originally passed to NewArray, converted to canonical form.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	assert.Equal(t, []interface{}{"foo", 123.0}, array.Raw())
func (a *Array) Raw() []interface{} {
	return a.value
}

// Decode unmarshals the underlying value attached to the Array to a target variable.
// target should be one of these:
//
// - pointer to an empty interface
// - pointer to a slice of any type
//
// Example:
//
//	type S struct{
//		Foo int `json:foo`
//	}
//	value := []interface{}{
//		map[string]interface{}{
//			"foo": 123,
//		},
//		map[string]interface{}{
//			"foo": 456,
//		},
//	}
//	array := NewArray(t, value)
//
//	var target []S
//	arr.Decode(&target)
//
//	assert.Equal(t, []S{{123}, {456}}, target)
func (a *Array) Decode(target interface{}) *Array {
	opChain := a.chain.enter("Decode()")
	defer opChain.leave()

	if opChain.failed() {
		return a
	}

	canonDecode(opChain, a.value, target)
	return a
}

// Alias is similar to Value.Alias.
func (a *Array) Alias(name string) *Array {
	opChain := a.chain.enter("Alias(%q)", name)
	defer opChain.leave()

	a.chain.setAlias(name)
	return a
}

// Path is similar to Value.Path.
func (a *Array) Path(path string) *Value {
	opChain := a.chain.enter("Path(%q)", path)
	defer opChain.leave()

	return jsonPath(opChain, a.value, path)
}

// Schema is similar to Value.Schema.
func (a *Array) Schema(schema interface{}) *Array {
	opChain := a.chain.enter("Schema()")
	defer opChain.leave()

	jsonSchema(opChain, a.value, schema)
	return a
}

// Length returns a new Number instance with array length.
//
// Example:
//
//	array := NewArray(t, []interface{}{1, 2, 3})
//	array.Length().Equal(3)
func (a *Array) Length() *Number {
	opChain := a.chain.enter("Length()")
	defer opChain.leave()

	if opChain.failed() {
		return newNumber(opChain, 0)
	}

	return newNumber(opChain, float64(len(a.value)))
}

// Element returns a new Value instance with array element for given index.
//
// If index is out of array bounds, Element reports failure and returns empty
// (but non-nil) instance.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.Element(0).String().Equal("foo")
//	array.Element(1).Number().Equal(123)
func (a *Array) Element(index int) *Value {
	opChain := a.chain.enter("Element(%d)", index)
	defer opChain.leave()

	if opChain.failed() {
		return newValue(opChain, nil)
	}

	if index < 0 || index >= len(a.value) {
		opChain.fail(AssertionFailure{
			Type:   AssertInRange,
			Actual: &AssertionValue{index},
			Expected: &AssertionValue{AssertionRange{
				Min: 0,
				Max: len(a.value) - 1,
			}},
			Errors: []error{
				errors.New("expected: valid element index"),
			},
		})
		return newValue(opChain, nil)
	}

	return newValue(opChain, a.value[index])
}

// First returns a new Value instance for the first element of array.
//
// If given array is empty, First reports failure and returns empty
// (but non-nil) instance.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.First().String().Equal("foo")
func (a *Array) First() *Value {
	opChain := a.chain.enter("First()")
	defer opChain.leave()

	if opChain.failed() {
		return newValue(opChain, nil)
	}

	if len(a.value) == 0 {
		opChain.fail(AssertionFailure{
			Type:   AssertNotEmpty,
			Actual: &AssertionValue{a.value},
			Errors: []error{
				errors.New("expected: non-empty array"),
			},
		})
		return newValue(opChain, nil)
	}

	return newValue(opChain, a.value[0])
}

// Last returns a new Value instance for the last element of array.
//
// If given array is empty, Last reports failure and returns empty
// (but non-nil) instance.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.Last().Number().Equal(123)
func (a *Array) Last() *Value {
	opChain := a.chain.enter("Last()")
	defer opChain.leave()

	if opChain.failed() {
		return newValue(opChain, nil)
	}

	if len(a.value) == 0 {
		opChain.fail(AssertionFailure{
			Type:   AssertNotEmpty,
			Actual: &AssertionValue{a.value},
			Errors: []error{
				errors.New("expected: non-empty array"),
			},
		})
		return newValue(opChain, nil)
	}

	return newValue(opChain, a.value[len(a.value)-1])
}

// Iter returns a new slice of Values attached to array elements.
//
// Example:
//
//	strings := []interface{}{"foo", "bar"}
//	array := NewArray(t, strings)
//
//	for index, value := range array.Iter() {
//	    value.String().Equal(strings[index])
//	}
func (a *Array) Iter() []Value {
	opChain := a.chain.enter("Iter()")
	defer opChain.leave()

	if opChain.failed() {
		return []Value{}
	}

	ret := []Value{}

	for index, element := range a.value {
		func() {
			valueChain := opChain.replace("Iter[%v]", index)
			defer valueChain.leave()

			ret = append(ret, *newValue(valueChain, element))
		}()
	}

	return ret
}

// Every runs the passed function on all the elements in the array.
//
// If assertion inside function fails, the original Array is marked failed.
//
// Every will execute the function for all values in the array irrespective
// of assertion failures for some values in the array.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", "bar"})
//
//	array.Every(func(index int, value *httpexpect.Value) {
//		value.String().NotEmpty()
//	})
func (a *Array) Every(fn func(index int, value *Value)) *Array {
	opChain := a.chain.enter("Every()")
	defer opChain.leave()

	if opChain.failed() {
		return a
	}

	if fn == nil {
		opChain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected nil function argument"),
			},
		})
		return a
	}

	for index, element := range a.value {
		func() {
			valueChain := opChain.replace("Every[%v]", index)
			defer valueChain.leave()

			fn(index, newValue(valueChain, element))
		}()
	}

	return a
}

// Filter accepts a function that returns a boolean. The function is ran
// over the array elements. If the function returns true, the element passes
// the filter and is added to the new array of filtered elements. If false,
// the element is skipped (or in other words filtered out). After iterating
// through all the elements of the original array, the new filtered array
// is returned.
//
// If there are any failed assertions in the filtering function, the
// element is omitted without causing test failure.
//
// Example:
//
//	array := NewArray(t, []interface{}{1, 2, "foo", "bar"})
//	filteredArray := array.Filter(func(index int, value *httpexpect.Value) bool {
//		value.String().NotEmpty()		//fails on 1 and 2
//		return value.Raw() != "bar"		//fails on "bar"
//	})
//	filteredArray.Equal([]interface{}{"foo"})	//succeeds
func (a *Array) Filter(fn func(index int, value *Value) bool) *Array {
	opChain := a.chain.enter("Filter()")
	defer opChain.leave()

	if opChain.failed() {
		return newArray(opChain, nil)
	}

	if fn == nil {
		opChain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected nil function argument"),
			},
		})
		return newArray(opChain, nil)
	}

	filteredArray := []interface{}{}

	for index, element := range a.value {
		func() {
			valueChain := opChain.replace("Filter[%v]", index)
			defer valueChain.leave()

			valueChain.setRoot()
			valueChain.setSeverity(SeverityLog)

			if fn(index, newValue(valueChain, element)) && !valueChain.treeFailed() {
				filteredArray = append(filteredArray, element)
			}
		}()
	}

	return newArray(opChain, filteredArray)
}

// Transform runs the passed function on all the elements in the array
// and returns a new array without effeecting original array.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", "bar"})
//	transformedArray := array.Transform(
//		func(index int, value interface{}) interface{} {
//			return strings.ToUpper(value.(string))
//		})
//	transformedArray.Equals([]interface{}{"FOO", "BAR"})
func (a *Array) Transform(fn func(index int, value interface{}) interface{}) *Array {
	opChain := a.chain.enter("Transform()")
	defer opChain.leave()

	if opChain.failed() {
		return newArray(opChain, nil)
	}

	if fn == nil {
		opChain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected nil function argument"),
			},
		})
		return newArray(opChain, nil)
	}

	transformedArray := []interface{}{}

	for index, element := range a.value {
		transformedArray = append(transformedArray, fn(index, element))
	}

	return newArray(opChain, transformedArray)
}

// Find accepts a function that returns a boolean, runs it over the array
// elements, and returns the first element on which it returned true.
//
// If there are any failed assertions in the predicate function, the
// element is skipped without causing test failure.
//
// If no elements were found, a failure is reported.
//
// Example:
//
//	array := NewArray(t, []interface{}{1, "foo", 101, "bar", 201})
//	foundValue := array.Find(func(index int, value *httpexpect.Value) bool {
//		num := value.Number()    // skip if element is not a number
//		return num.Raw() > 100   // check element value
//	})
//	foundValue.Equal(101) // succeeds
func (a *Array) Find(fn func(index int, value *Value) bool) *Value {
	opChain := a.chain.enter("Find()")
	defer opChain.leave()

	if opChain.failed() {
		return newValue(opChain, nil)
	}

	if fn == nil {
		opChain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected nil function argument"),
			},
		})
		return newValue(opChain, nil)
	}

	for index, element := range a.value {
		found := false

		func() {
			valueChain := opChain.replace("Find[%v]", index)
			defer valueChain.leave()

			valueChain.setRoot()
			valueChain.setSeverity(SeverityLog)

			if fn(index, newValue(valueChain, element)) && !valueChain.treeFailed() {
				found = true
			}
		}()

		if found {
			return newValue(opChain, element)
		}
	}

	opChain.fail(AssertionFailure{
		Type:   AssertValid,
		Actual: &AssertionValue{a.value},
		Errors: []error{
			errors.New("expected: at least one array element matches predicate"),
		},
	})

	return newValue(opChain, nil)
}

// FindAll accepts a function that returns a boolean, runs it over the array
// elements, and returns all the elements on which it returned true.
//
// If there are any failed assertions in the predicate function, the
// element is skipped without causing test failure.
//
// If no elements were found, empty slice is returned without reporting error.
//
// Example:
//
//	array := NewArray(t, []interface{}{1, "foo", 101, "bar", 201})
//	foundValues := array.FindAll(func(index int, value *httpexpect.Value) bool {
//		num := value.Number()   // skip if element is not a number
//		return num.Raw() > 100  // check element value
//	})
//
//	assert.Equal(t, len(foundValues), 2)
//	foundValues[0].Equal(101)
//	foundValues[1].Equal(201)
func (a *Array) FindAll(fn func(index int, value *Value) bool) []*Value {
	opChain := a.chain.enter("FindAll()")
	defer opChain.leave()

	if opChain.failed() {
		return []*Value{}
	}

	if fn == nil {
		opChain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected nil function argument"),
			},
		})
		return []*Value{}
	}

	foundValues := make([]*Value, 0, len(a.value))

	for index, element := range a.value {
		func() {
			valueChain := opChain.replace("FindAll[%v]", index)
			defer valueChain.leave()

			valueChain.setRoot()
			valueChain.setSeverity(SeverityLog)

			if fn(index, newValue(valueChain, element)) && !valueChain.treeFailed() {
				foundValues = append(foundValues, newValue(opChain, element))
			}
		}()
	}

	return foundValues
}

// NotFind accepts a function that returns a boolean, runs it over the array
// elelements, and checks that it does not return true for any of the elements.
//
// If there are any failed assertions in the predicate function, the
// element is skipped without causing test failure.
//
// If the predicate function did not fail and returned true for at least
// one element, a failure is reported.
//
// Example:
//
//	array := NewArray(t, []interface{}{1, "foo", 2, "bar"})
//	array.NotFind(func(index int, value *httpexpect.Value) bool {
//		num := value.Number()    // skip if element is not a number
//		return num.Raw() > 100   // check element value
//	}) // succeeds
func (a *Array) NotFind(fn func(index int, value *Value) bool) *Array {
	opChain := a.chain.enter("NotFind()")
	defer opChain.leave()

	if opChain.failed() {
		return a
	}

	if fn == nil {
		opChain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected nil function argument"),
			},
		})
		return a
	}

	for index, element := range a.value {
		found := false

		func() {
			valueChain := opChain.replace("NotFind[%v]", index)
			defer valueChain.leave()

			valueChain.setRoot()
			valueChain.setSeverity(SeverityLog)

			if fn(index, newValue(valueChain, element)) && !valueChain.treeFailed() {
				found = true
			}
		}()

		if found {
			opChain.fail(AssertionFailure{
				Type:     AssertNotContainsElement,
				Expected: &AssertionValue{element},
				Actual:   &AssertionValue{a.value},
				Errors: []error{
					errors.New("expected: none of the array elements match predicate"),
					fmt.Errorf("element with index %d matches predicate", index),
				},
			})
			return a
		}
	}

	return a
}

// Empty succeeds if array is empty.
//
// Example:
//
//	array := NewArray(t, []interface{}{})
//	array.Empty()
func (a *Array) Empty() *Array {
	opChain := a.chain.enter("Empty()")
	defer opChain.leave()

	if opChain.failed() {
		return a
	}

	if !(len(a.value) == 0) {
		opChain.fail(AssertionFailure{
			Type:   AssertEmpty,
			Actual: &AssertionValue{a.value},
			Errors: []error{
				errors.New("expected: empty array"),
			},
		})
	}

	return a
}

// NotEmpty succeeds if array is non-empty.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.NotEmpty()
func (a *Array) NotEmpty() *Array {
	opChain := a.chain.enter("NotEmpty()")
	defer opChain.leave()

	if opChain.failed() {
		return a
	}

	if len(a.value) == 0 {
		opChain.fail(AssertionFailure{
			Type:   AssertNotEmpty,
			Actual: &AssertionValue{a.value},
			Errors: []error{
				errors.New("expected: non-empty array"),
			},
		})
	}

	return a
}

// Equal succeeds if array is equal to given value.
// Before comparison, both array and value are converted to canonical form.
//
// value should be a slice of any type.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.Equal([]interface{}{"foo", 123})
//
//	array := NewArray(t, []interface{}{"foo", "bar"})
//	array.Equal([]string{}{"foo", "bar"})
//
//	array := NewArray(t, []interface{}{123, 456})
//	array.Equal([]int{}{123, 456})
func (a *Array) Equal(value interface{}) *Array {
	opChain := a.chain.enter("Equal()")
	defer opChain.leave()

	if opChain.failed() {
		return a
	}

	expected, ok := canonArray(opChain, value)
	if !ok {
		return a
	}

	if !reflect.DeepEqual(expected, a.value) {
		opChain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{a.value},
			Expected: &AssertionValue{expected},
			Errors: []error{
				errors.New("expected: arrays are equal"),
			},
		})
	}

	return a
}

// NotEqual succeeds if array is not equal to given value.
// Before comparison, both array and value are converted to canonical form.
//
// value should be a slice of any type.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.NotEqual([]interface{}{123, "foo"})
func (a *Array) NotEqual(value interface{}) *Array {
	opChain := a.chain.enter("NotEqual()")
	defer opChain.leave()

	if opChain.failed() {
		return a
	}

	expected, ok := canonArray(opChain, value)
	if !ok {
		return a
	}

	if reflect.DeepEqual(expected, a.value) {
		opChain.fail(AssertionFailure{
			Type:     AssertNotEqual,
			Actual:   &AssertionValue{a.value},
			Expected: &AssertionValue{expected},
			Errors: []error{
				errors.New("expected: arrays are non-equal"),
			},
		})
	}

	return a
}

// EqualUnordered succeeds if array is equal to another array, ignoring element
// order. Before comparison, both arrays are converted to canonical form.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.EqualUnordered([]interface{}{123, "foo"})
func (a *Array) EqualUnordered(value interface{}) *Array {
	opChain := a.chain.enter("EqualUnordered()")
	defer opChain.leave()

	if opChain.failed() {
		return a
	}

	expected, ok := canonArray(opChain, value)
	if !ok {
		return a
	}

	for _, element := range expected {
		expectedCount := countElement(expected, element)
		actualCount := countElement(a.value, element)

		if actualCount != expectedCount {
			if expectedCount == 1 && actualCount == 0 {
				opChain.fail(AssertionFailure{
					Type:      AssertContainsElement,
					Actual:    &AssertionValue{a.value},
					Expected:  &AssertionValue{element},
					Reference: &AssertionValue{value},
					Errors: []error{
						errors.New("expected: array contains element from reference array"),
					},
				})
			} else {
				opChain.fail(AssertionFailure{
					Type:      AssertNotContainsElement,
					Actual:    &AssertionValue{a.value},
					Expected:  &AssertionValue{element},
					Reference: &AssertionValue{value},
					Errors: []error{
						fmt.Errorf(
							"expected: element occurs %d time(s), as in reference array,"+
								" but it occurs %d time(s)",
							expectedCount,
							actualCount),
					},
				})
			}
			return a
		}
	}

	for _, element := range a.value {
		expectedCount := countElement(expected, element)
		actualCount := countElement(a.value, element)

		if actualCount != expectedCount {
			if expectedCount == 0 && actualCount == 1 {
				opChain.fail(AssertionFailure{
					Type:      AssertNotContainsElement,
					Actual:    &AssertionValue{a.value},
					Expected:  &AssertionValue{element},
					Reference: &AssertionValue{value},
					Errors: []error{
						errors.New("expected: array does not contain elements" +
							" that are not present in reference array"),
					},
				})
			} else {
				opChain.fail(AssertionFailure{
					Type:      AssertNotContainsElement,
					Actual:    &AssertionValue{a.value},
					Expected:  &AssertionValue{element},
					Reference: &AssertionValue{value},
					Errors: []error{
						fmt.Errorf(
							"expected: element occurs %d time(s), as in reference array,"+
								" but it occurs %d time(s)",
							expectedCount,
							actualCount),
					},
				})
			}
			return a
		}
	}

	return a
}

// NotEqualUnordered succeeds if array is not equal to another array, ignoring
// element order. Before comparison, both arrays are converted to canonical form.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.NotEqualUnordered([]interface{}{123, "foo", "bar"})
func (a *Array) NotEqualUnordered(value interface{}) *Array {
	opChain := a.chain.enter("NotEqualUnordered()")
	defer opChain.leave()

	if opChain.failed() {
		return a
	}

	expected, ok := canonArray(opChain, value)
	if !ok {
		return a
	}

	different := false

	for _, element := range expected {
		expectedCount := countElement(expected, element)
		actualCount := countElement(a.value, element)

		if actualCount != expectedCount {
			different = true
			break
		}
	}

	for _, element := range a.value {
		expectedCount := countElement(expected, element)
		actualCount := countElement(a.value, element)

		if actualCount != expectedCount {
			different = true
			break
		}
	}

	if !different {
		opChain.fail(AssertionFailure{
			Type:      AssertNotEqual,
			Actual:    &AssertionValue{a.value},
			Expected:  &AssertionValue{value},
			Reference: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: arrays are non-equal (ignoring order)"),
			},
		})
	}

	return a
}

// ConsistsOf succeeds if array contains all given elements, in given order, and only
// them. Before comparison, array and all elements are converted to canonical form.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.ConsistsOf("foo", 123)
//
// These calls are equivalent:
//
//	array.ConsistsOf("a", "b")
//	array.Equal([]interface{}{"a", "b"})
func (a *Array) ConsistsOf(values ...interface{}) *Array {
	opChain := a.chain.enter("ConsistsOf()")
	defer opChain.leave()

	if opChain.failed() {
		return a
	}

	expected, ok := canonArray(opChain, values)
	if !ok {
		return a
	}

	if !reflect.DeepEqual(expected, a.value) {
		opChain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{a.value},
			Expected: &AssertionValue{expected},
			Errors: []error{
				errors.New("expected: array consists of given elements"),
			},
		})
	}

	return a
}

// NotConsistsOf is opposite to ConsistsOf.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.NotConsistsOf("foo")
//	array.NotConsistsOf("foo", 123, 456)
//	array.NotConsistsOf(123, "foo")
//
// These calls are equivalent:
//
//	array.NotConsistsOf("a", "b")
//	array.NotEqual([]interface{}{"a", "b"})
func (a *Array) NotConsistsOf(values ...interface{}) *Array {
	opChain := a.chain.enter("NotConsistsOf()")
	defer opChain.leave()

	if opChain.failed() {
		return a
	}

	expected, ok := canonArray(opChain, values)
	if !ok {
		return a
	}

	if reflect.DeepEqual(expected, a.value) {
		opChain.fail(AssertionFailure{
			Type:     AssertNotEqual,
			Actual:   &AssertionValue{a.value},
			Expected: &AssertionValue{expected},
			Errors: []error{
				errors.New("expected: arrays does not consist of given elements"),
			},
		})
	}

	return a
}

// Deprecated: use ConsistsOf instead.
func (a *Array) Elements(values ...interface{}) *Array {
	return a.ConsistsOf(values...)
}

// Deprecated: use NotConsistsOf instead.
func (a *Array) NotElements(values ...interface{}) *Array {
	return a.NotConsistsOf(values...)
}

// Deprecated: use ContainsAll or ContainsAny instead.
func (a *Array) Contains(values ...interface{}) *Array {
	opChain := a.chain.enter("Contains()")
	defer opChain.leave()

	if opChain.failed() {
		return a
	}

	elements, ok := canonArray(opChain, values)
	if !ok {
		return a
	}

	for _, expected := range elements {
		if countElement(a.value, expected) == 0 {
			opChain.fail(AssertionFailure{
				Type:      AssertContainsElement,
				Actual:    &AssertionValue{a.value},
				Expected:  &AssertionValue{expected},
				Reference: &AssertionValue{values},
				Errors: []error{
					errors.New("expected: array contains element from reference array"),
				},
			})
			break
		}
	}

	return a
}

// Deprecated: use NotContainsAll or NotContainsAny instead.
func (a *Array) NotContains(values ...interface{}) *Array {
	opChain := a.chain.enter("NotContains()")
	defer opChain.leave()

	if opChain.failed() {
		return a
	}

	elements, ok := canonArray(opChain, values)
	if !ok {
		return a
	}

	for _, expected := range elements {
		if !(countElement(a.value, expected) == 0) {
			opChain.fail(AssertionFailure{
				Type:      AssertNotContainsElement,
				Actual:    &AssertionValue{a.value},
				Expected:  &AssertionValue{expected},
				Reference: &AssertionValue{values},
				Errors: []error{
					errors.New("expected:" +
						" array does not contain any elements from reference array"),
				},
			})
			break
		}
	}

	return a
}

// ContainsAll succeeds if array contains all given elements (in any order).
// Before comparison, array and all elements are converted to canonical form.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.ContainsAll(123, "foo")
func (a *Array) ContainsAll(values ...interface{}) *Array {
	opChain := a.chain.enter("ContainsAll()")
	defer opChain.leave()

	if opChain.failed() {
		return a
	}

	elements, ok := canonArray(opChain, values)
	if !ok {
		return a
	}

	for _, expected := range elements {
		if countElement(a.value, expected) == 0 {
			opChain.fail(AssertionFailure{
				Type:      AssertContainsElement,
				Actual:    &AssertionValue{a.value},
				Expected:  &AssertionValue{expected},
				Reference: &AssertionValue{values},
				Errors: []error{
					errors.New("expected: array contains element from reference array"),
				},
			})
			break
		}
	}

	return a
}

// NotContainsAll succeeds if array does not contain at least one of the elements.
// Before comparison, array and all elements are converted to canonical form.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.NotContainsAll("bar")         // success
//	array.NotContainsAll(123, "foo")    // failure
func (a *Array) NotContainsAll(values ...interface{}) *Array {
	opChain := a.chain.enter("NotContainsAll()")
	defer opChain.leave()

	if opChain.failed() {
		return a
	}

	elements, ok := canonArray(opChain, values)
	if !ok {
		return a
	}

	haveMissing := false

	for _, expected := range elements {
		if countElement(a.value, expected) == 0 {
			haveMissing = true
			break
		}
	}

	if !haveMissing {
		opChain.fail(AssertionFailure{
			Type:      AssertNotContainsElement,
			Actual:    &AssertionValue{a.value},
			Reference: &AssertionValue{values},
			Errors: []error{
				errors.New("expected:" +
					" array does not contain at least one element from reference array"),
			},
		})
	}

	return a
}

// ContainsAny succeeds if array contains at least one element from the given elements.
// Before comparison, array and all elements are converted to canonical form.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123, 123})
//	array.ContainsAny(123, "foo", "FOO") // success
//	array.ContainsAny("FOO") // failure
func (a *Array) ContainsAny(values ...interface{}) *Array {
	opChain := a.chain.enter("ContainsAny()")
	defer opChain.leave()

	if opChain.failed() {
		return a
	}

	elements, ok := canonArray(opChain, values)
	if !ok {
		return a
	}

	foundAny := false

	for _, expected := range elements {
		if countElement(a.value, expected) != 0 {
			foundAny = true
			break
		}
	}

	if !foundAny {
		opChain.fail(AssertionFailure{
			Type:      AssertContainsElement,
			Actual:    &AssertionValue{a.value},
			Reference: &AssertionValue{values},
			Errors: []error{
				errors.New("expected:" +
					" array contains at least one element from reference array"),
			},
		})
	}

	return a
}

// NotContainsAny succeeds if none of the given elements are in the array.
// Before comparison, array and all elements are converted to canonical form.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.NotContainsAny("bar", 124) // success
//	array.NotContainsAny(123) // failure
func (a *Array) NotContainsAny(values ...interface{}) *Array {
	opChain := a.chain.enter("NotContainsAny()")
	defer opChain.leave()

	if opChain.failed() {
		return a
	}

	elements, ok := canonArray(opChain, values)
	if !ok {
		return a
	}

	for _, expected := range elements {
		if countElement(a.value, expected) != 0 {
			opChain.fail(AssertionFailure{
				Type:      AssertNotContainsElement,
				Actual:    &AssertionValue{a.value},
				Expected:  &AssertionValue{expected},
				Reference: &AssertionValue{values},
				Errors: []error{
					errors.New("expected:" +
						" array does not contain any elements from reference array"),
				},
			})
			return a
		}
	}

	return a
}

// ContainsOnly succeeds if array contains all given elements, in any order, and only
// them, ignoring duplicates. Before comparison, array and all elements are converted
// to canonical form.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123, 123})
//	array.ContainsOnly(123, "foo")
//
// These calls are equivalent:
//
//	array.ContainsOnly("a", "b")
//	array.ContainsOnly("b", "a")
func (a *Array) ContainsOnly(values ...interface{}) *Array {
	opChain := a.chain.enter("ContainsOnly()")
	defer opChain.leave()

	if opChain.failed() {
		return a
	}

	elements, ok := canonArray(opChain, values)
	if !ok {
		return a
	}

	for _, element := range elements {
		if countElement(a.value, element) == 0 {
			opChain.fail(AssertionFailure{
				Type:      AssertContainsElement,
				Actual:    &AssertionValue{a.value},
				Expected:  &AssertionValue{element},
				Reference: &AssertionValue{values},
				Errors: []error{
					errors.New("expected: array contains element from reference array"),
				},
			})
			return a
		}
	}

	for _, element := range a.value {
		if countElement(elements, element) == 0 {
			opChain.fail(AssertionFailure{
				Type:      AssertNotContainsElement,
				Actual:    &AssertionValue{a.value},
				Expected:  &AssertionValue{element},
				Reference: &AssertionValue{values},
				Errors: []error{
					errors.New("expected: array does not contain elements" +
						" that are not present in reference array"),
				},
			})
			return a
		}
	}

	return a
}

// NotContainsOnly is opposite to ContainsOnly.
//
// Example:
//
//	array := NewArray(t, []interface{}{"foo", 123})
//	array.NotContainsOnly(123)
//	array.NotContainsOnly(123, "foo", "bar")
//
// These calls are equivalent:
//
//	array.NotContainsOnly("a", "b")
//	array.NotContainsOnly("b", "a")
func (a *Array) NotContainsOnly(values ...interface{}) *Array {
	opChain := a.chain.enter("NotContainsOnly()")
	defer opChain.leave()

	if opChain.failed() {
		return a
	}

	elements, ok := canonArray(opChain, values)
	if !ok {
		return a
	}

	different := false

	for _, element := range elements {
		if countElement(a.value, element) == 0 {
			different = true
			break
		}
	}

	for _, element := range a.value {
		if countElement(elements, element) == 0 {
			different = true
			break
		}
	}

	if !different {
		opChain.fail(AssertionFailure{
			Type:      AssertNotEqual,
			Actual:    &AssertionValue{a.value},
			Expected:  &AssertionValue{values},
			Reference: &AssertionValue{values},
			Errors: []error{
				errors.New("expected:" +
					" array does not contain only elements from reference array" +
					" (at least one distinguishing element needed)"),
			},
		})
	}

	return a
}

// IsOrdered succeeds if every element is not less than the previous element
// as defined on the given `less` comparator function.
// For default, it will use built-in comparator function for each data type.
// Built-in comparator requires all elements in the array to have same data type.
// Array with 0 or 1 element will always succeed
//
// Example:
//
//	array := NewArray(t, []interface{}{100, 101, 102})
//	array.IsOrdered() // succeeds
//	array.IsOrdered(func(x, y *httpexpect.Value) bool {
//		return x.Number().Raw() < y.Number().Raw()
//	}) // succeeds
func (a *Array) IsOrdered(less ...func(x, y *Value) bool) *Array {
	opChain := a.chain.enter("IsOrdered()")
	defer opChain.leave()

	if opChain.failed() {
		return a
	}

	if len(less) > 1 {
		opChain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected multiple less arguments"),
			},
		})
		return a
	}

	if len(a.value) <= 1 {
		return a
	}

	var lessFn func(x, y *Value) bool
	if len(less) == 1 {
		lessFn = less[0]
	} else {
		lessFn = arrayComparator(opChain, a.value)
		if lessFn == nil {
			return a
		}
	}

	for i := 0; i < len(a.value)-1; i++ {
		var unordered bool

		func() {
			xChain := opChain.replace("IsOrdered[%d]", i)
			defer xChain.leave()

			yChain := opChain.replace("IsOrdered[%d]", i+1)
			defer yChain.leave()

			x := newValue(xChain, a.value[i])
			y := newValue(yChain, a.value[i+1])

			unordered = lessFn(y, x)
		}()

		if opChain.failed() {
			return a
		}

		if unordered {
			opChain.fail(AssertionFailure{
				Type:      AssertLt,
				Actual:    &AssertionValue{a.value[i]},
				Expected:  &AssertionValue{a.value[i+1]},
				Reference: &AssertionValue{a.value},
				Errors: []error{
					errors.New("expected: reference array is ordered"),
					fmt.Errorf("element %v must not be less than element %v",
						i+1, i),
				},
			})
			return a
		}
	}

	return a
}

// NotOrdered succeeds if at least one element is less than the previous element
// as defined on the given `less` comparator function.
// For default, it will use built-in comparator function for each data type.
// Built-in comparator requires all elements in the array to have same data type.
// Array with 0 or 1 element will always succeed
//
// Example:
//
//	array := NewArray(t, []interface{}{102, 101, 100})
//	array.NotOrdered() // succeeds
//	array.NotOrdered(func(x, y *httpexpect.Value) bool {
//		return x.Number().Raw() < y.Number().Raw()
//	}) // succeeds
func (a *Array) NotOrdered(less ...func(x, y *Value) bool) *Array {
	opChain := a.chain.enter("NotOrdered()")
	defer opChain.leave()

	if opChain.failed() {
		return a
	}

	if len(less) > 1 {
		opChain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected multiple less arguments"),
			},
		})
		return a
	}

	if len(a.value) <= 1 {
		return a
	}

	var lessFn func(x, y *Value) bool
	if len(less) == 1 {
		lessFn = less[0]
	} else {
		lessFn = arrayComparator(opChain, a.value)
		if lessFn == nil {
			return a
		}
	}

	ordered := true

	for i := 0; i < len(a.value)-1; i++ {
		func() {
			xChain := opChain.replace("IsOrdered[%d]", i)
			defer xChain.leave()

			yChain := opChain.replace("IsOrdered[%d]", i+1)
			defer yChain.leave()

			x := newValue(xChain, a.value[i])
			y := newValue(yChain, a.value[i+1])

			if lessFn(y, x) {
				ordered = false
			}
		}()

		if opChain.failed() {
			return a
		}

		if !ordered {
			break
		}
	}

	if ordered {
		opChain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{a.value},
			Errors: []error{
				errors.New("expected: array is not ordered, but it is"),
			},
		})
	}

	return a
}

func countElement(array []interface{}, element interface{}) int {
	count := 0
	for _, e := range array {
		if reflect.DeepEqual(element, e) {
			count++
		}
	}
	return count
}
