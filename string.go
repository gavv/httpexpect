package httpexpect

import (
	"strings"
)

// String provides methods to inspect attached string value
// (Go representation of JSON string).
type String struct {
	chain chain
	value string
}

// NewString returns a new String given a reporter used to report failures
// and value to be inspected.
//
// reporter should not be nil.
//
// Example:
//  str := NewString(t, "Hello")
func NewString(reporter Reporter, value string) *String {
	return &String{makeChain(reporter), value}
}

// Raw returns underlying value attached to Str.
// This is the value originally passed to NewStr.
//
// Example:
//  str := NewString(t, "Hello")
//  assert.Equal(t, "Hello", str.Raw())
func (s *String) Raw() string {
	return s.value
}

// Empty succeedes if string is empty.
//
// Example:
//  str := NewString(t, "")
//  str.Empty()
func (s *String) Empty() *String {
	return s.Equal("")
}

// NotEmpty succeedes if string is non-empty.
//
// Example:
//  str := NewString(t, "Hello")
//  str.NotEmpty()
func (s *String) NotEmpty() *String {
	return s.NotEqual("")
}

// Equal succeedes if string is equal to another str.
//
// Example:
//  str := NewString(t, "Hello")
//  str.Equal("Hello")
func (s *String) Equal(value string) *String {
	if !(s.value == value) {
		s.chain.fail("expected string == \"%s\", but got \"%s\"", value, s.value)
	}
	return s
}

// NotEqual succeedes if string is not equal to another str.
//
// Example:
//  str := NewString(t, "Hello")
//  str.NotEqual("Goodbye")
func (s *String) NotEqual(value string) *String {
	if !(s.value != value) {
		s.chain.fail("expected string != \"%s\", but got \"%s\"", value, s.value)
	}
	return s
}

// EqualFold succeedes if string is equal to another string under Unicode case-folding
// (case-insensitive match).
//
// Example:
//  str := NewString(t, "Hello")
//  str.EqualFold("hELLo")
func (s *String) EqualFold(value string) *String {
	if !strings.EqualFold(s.value, value) {
		s.chain.fail(
			"expected string == \"%s\" (case-insensitive), but got \"%s\"", value, s.value)
	}
	return s
}

// NotEqualFold succeedes if string is not equal to another string under Unicode
// case-folding (case-insensitive match).
//
// Example:
//  str := NewString(t, "Hello")
//  str.NotEqualFold("gOODBYe")
func (s *String) NotEqualFold(value string) *String {
	if strings.EqualFold(s.value, value) {
		s.chain.fail(
			"expected string != \"%s\" (case-insensitive), but got \"%s\"", value, s.value)
	}
	return s
}

// Contains succeedes if string contains given substr.
//
// Example:
//  str := NewString(t, "Hello")
//  str.Contains("ell")
func (s *String) Contains(value string) *String {
	if !strings.Contains(s.value, value) {
		s.chain.fail(
			"expected string containing substring \"%s\", but got \"%s\"", value, s.value)
	}
	return s
}

// NotContains succeedes if string doesn't contain given substr.
//
// Example:
//  str := NewString(t, "Hello")
//  str.NotContains("bye")
func (s *String) NotContains(value string) *String {
	if strings.Contains(s.value, value) {
		s.chain.fail(
			"expected string NOT containing substring \"%s\", but got \"%s\"",
			value, s.value)
	}
	return s
}

// ContainsFold succeedes if string contains given substring under Unicode case-folding
// (case-insensitive match).
//
// Example:
//  str := NewString(t, "Hello")
//  str.ContainsFold("ELL")
func (s *String) ContainsFold(value string) *String {
	if !strings.Contains(strings.ToLower(s.value), strings.ToLower(value)) {
		s.chain.fail(
			"expected string containing substring \"%s\" (case-insensitive), "+
				"but got \"%s\"", value, s.value)
	}
	return s
}

// NotContainsFold succeedes if string doesn't contain given substring under Unicode
// case-folding (case-insensitive match).
//
// Example:
//  str := NewString(t, "Hello")
//  str.NotContainsFold("BYE")
func (s *String) NotContainsFold(value string) *String {
	if strings.Contains(strings.ToLower(s.value), strings.ToLower(value)) {
		s.chain.fail(
			"expected string NOT containing substring \"%s\" (case-insensitive), "+
				"but got \"%s\"", value, s.value)
	}
	return s
}
