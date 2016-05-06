package httpexpect

import (
	"strings"
)

type String struct {
	checker Checker
	value   string
}

func NewString(checker Checker, value string) *String {
	return &String{checker, value}
}

func (s *String) Raw() string {
	return s.value
}

func (s *String) Empty() *String {
	return s.Equal("")
}

func (s *String) NotEmpty() *String {
	return s.NotEqual("")
}

func (s *String) Equal(v string) *String {
	if !(s.value == v) {
		s.checker.Fail("expected string == \"%s\", got \"%s\"", v, s.value)
	}
	return s
}

func (s *String) NotEqual(v string) *String {
	if !(s.value != v) {
		s.checker.Fail("expected string != \"%s\", got \"%s\"", v, s.value)
	}
	return s
}

func (s *String) EqualFold(v string) *String {
	if !strings.EqualFold(s.value, v) {
		s.checker.Fail(
			"expected string == \"%s\" (case-insensitive), got \"%s\"", v, s.value)
	}
	return s
}

func (s *String) NotEqualFold(v string) *String {
	if strings.EqualFold(s.value, v) {
		s.checker.Fail(
			"expected string != \"%s\" (case-insensitive), got \"%s\"", v, s.value)
	}
	return s
}

func (s *String) Contains(v string) *String {
	if !strings.Contains(s.value, v) {
		s.checker.Fail(
			"expected string containing substring \"%s\", got \"%s\"", v, s.value)
	}
	return s
}

func (s *String) NotContains(v string) *String {
	if strings.Contains(s.value, v) {
		s.checker.Fail(
			"expected string NOT containing substring \"%s\", got \"%s\"", v, s.value)
	}
	return s
}

func (s *String) ContainsFold(v string) *String {
	if !strings.Contains(strings.ToLower(s.value), strings.ToLower(v)) {
		s.checker.Fail(
			"expected string containing substring \"%s\" (case-insensitive), "+
				"got \"%s\"", v, s.value)
	}
	return s
}

func (s *String) NotContainsFold(v string) *String {
	if strings.Contains(strings.ToLower(s.value), strings.ToLower(v)) {
		s.checker.Fail(
			"expected string NOT containing substring \"%s\" (case-insensitive), "+
				"got \"%s\"", v, s.value)
	}
	return s
}
