package httpexpect

import (
	"strings"
)

// String provides methods to inspect attached string value
// (Go representation of JSON string).
type String struct {
	checker Checker
	value   string
}

// NewString returns a new String given a checker used to report failures
// and value to be inspected.
//
// checker should not be nil.
//
// Example:
//  str := NewString(NewAssertChecker(t), "Hello")
func NewString(checker Checker, value string) *String {
	return &String{checker, value}
}

// Raw returns underlying value attached to Str.
// This is the value originally passed to NewStr.
//
// Example:
//  str := NewString(checker, "Hello")
//  assert.Equal(t, "Hello", str.Raw())
func (s *String) Raw() string {
	return s.value
}

// Empty succeedes if string is empty.
//
// Example:
//  str := NewString(checker, "")
//  str.Empty()
func (s *String) Empty() *String {
	return s.Equal("")
}

// NotEmpty succeedes if string is non-empty.
//
// Example:
//  str := NewString(checker, "Hello")
//  str.NotEmpty()
func (s *String) NotEmpty() *String {
	return s.NotEqual("")
}

// Equal succeedes if string is equal to another str.
//
// Example:
//  str := NewString(checker, "Hello")
//  str.Equal("Hello")
func (s *String) Equal(v string) *String {
	if !(s.value == v) {
		s.checker.Fail("expected string == \"%s\", got \"%s\"", v, s.value)
	}
	return s
}

// NotEqual succeedes if string is not equal to another str.
//
// Example:
//  str := NewString(checker, "Hello")
//  str.NotEqual("Goodbye")
func (s *String) NotEqual(v string) *String {
	if !(s.value != v) {
		s.checker.Fail("expected string != \"%s\", got \"%s\"", v, s.value)
	}
	return s
}

// EqualFold succeedes if string is equal to another string under Unicode case-folding
// (case-insensitive match).
//
// Example:
//  str := NewString(checker, "Hello")
//  str.EqualFold("hELLo")
func (s *String) EqualFold(v string) *String {
	if !strings.EqualFold(s.value, v) {
		s.checker.Fail(
			"expected string == \"%s\" (case-insensitive), got \"%s\"", v, s.value)
	}
	return s
}

// NotEqualFold succeedes if string is not equal to another string under Unicode
// case-folding (case-insensitive match).
//
// Example:
//  str := NewString(checker, "Hello")
//  str.NotEqualFold("gOODBYe")
func (s *String) NotEqualFold(v string) *String {
	if strings.EqualFold(s.value, v) {
		s.checker.Fail(
			"expected string != \"%s\" (case-insensitive), got \"%s\"", v, s.value)
	}
	return s
}

// Contains succeedes if string contains given substr.
//
// Example:
//  str := NewString(checker, "Hello")
//  str.Contains("ell")
func (s *String) Contains(v string) *String {
	if !strings.Contains(s.value, v) {
		s.checker.Fail(
			"expected string containing substring \"%s\", got \"%s\"", v, s.value)
	}
	return s
}

// NotContains succeedes if string doesn't contain given substr.
//
// Example:
//  str := NewString(checker, "Hello")
//  str.NotContains("bye")
func (s *String) NotContains(v string) *String {
	if strings.Contains(s.value, v) {
		s.checker.Fail(
			"expected string NOT containing substring \"%s\", got \"%s\"", v, s.value)
	}
	return s
}

// ContainsFold succeedes if string contains given substring under Unicode case-folding
// (case-insensitive match).
//
// Example:
//  str := NewString(checker, "Hello")
//  str.ContainsFold("ELL")
func (s *String) ContainsFold(v string) *String {
	if !strings.Contains(strings.ToLower(s.value), strings.ToLower(v)) {
		s.checker.Fail(
			"expected string containing substring \"%s\" (case-insensitive), "+
				"got \"%s\"", v, s.value)
	}
	return s
}

// NotContainsFold succeedes if string doesn't contain given substring under Unicode
// case-folding (case-insensitive match).
//
// Example:
//  str := NewString(checker, "Hello")
//  str.NotContainsFold("BYE")
func (s *String) NotContainsFold(v string) *String {
	if strings.Contains(strings.ToLower(s.value), strings.ToLower(v)) {
		s.checker.Fail(
			"expected string NOT containing substring \"%s\" (case-insensitive), "+
				"got \"%s\"", v, s.value)
	}
	return s
}
