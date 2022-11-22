package httpexpect

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// String provides methods to inspect attached string value
// (Go representation of JSON string).
type String struct {
	chain *chain
	value string
}

// NewString returns a new String instance.
//
// reporter should not be nil.
//
// Example:
//
//	str := NewString(t, "Hello")
func NewString(reporter Reporter, value string) *String {
	return newString(newChainWithDefaults("String()", reporter), value)
}

func newString(parent *chain, val string) *String {
	return &String{parent.clone(), val}
}

// Raw returns underlying value attached to String.
// This is the value originally passed to NewString.
//
// Example:
//
//	str := NewString(t, "Hello")
//	assert.Equal(t, "Hello", str.Raw())
func (s *String) Raw() string {
	return s.value
}

// Path is similar to Value.Path.
func (s *String) Path(path string) *Value {
	s.chain.enter("Path(%q)", path)
	defer s.chain.leave()

	return jsonPath(s.chain, s.value, path)
}

// Schema is similar to Value.Schema.
func (s *String) Schema(schema interface{}) *String {
	s.chain.enter("Schema()")
	defer s.chain.leave()

	jsonSchema(s.chain, s.value, schema)
	return s
}

// Length returns a new Number instance with string length.
//
// Example:
//
//	str := NewString(t, "Hello")
//	str.Length().Equal(5)
func (s *String) Length() *Number {
	s.chain.enter("Length()")
	defer s.chain.leave()

	if s.chain.failed() {
		return newNumber(s.chain, 0)
	}

	return newNumber(s.chain, float64(len(s.value)))
}

// Number parses float from string and returns a new Number instance
// with result.
//
// Example:
//
//	str := NewString(t, "1234")
//	str.Number()
func (s *String) Number() *Number {
	s.chain.enter("Number()")
	defer s.chain.leave()

	if s.chain.failed() {
		return newNumber(s.chain, 0)
	}

	num, err := strconv.ParseFloat(s.value, 64)

	if err != nil {
		s.chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{s.value},
			Errors: []error{
				errors.New("expected: string can be parsed to number"),
				err,
			},
		})
		return newNumber(s.chain, 0)
	}

	return newNumber(s.chain, num)
}

// DateTime parses date/time from string and returns a new DateTime instance
// with result.
//
// If layout is given, DateTime() uses time.Parse() with given layout.
// Otherwise, it uses http.ParseTime(). If pasing error occurred,
// DateTime reports failure and returns empty (but non-nil) instance.
//
// Example:
//
//	str := NewString(t, "Tue, 15 Nov 1994 08:12:31 GMT")
//	str.DateTime().Lt(time.Now())
//
//	str := NewString(t, "15 Nov 94 08:12 GMT")
//	str.DateTime(time.RFC822).Lt(time.Now())
func (s *String) DateTime(layout ...string) *DateTime {
	if len(layout) != 0 {
		s.chain.enter("DateTime(%q)", layout[0])
	} else {
		s.chain.enter("DateTime()")
	}
	defer s.chain.leave()

	if s.chain.failed() {
		return newDateTime(s.chain, time.Unix(0, 0))
	}

	var (
		tm  time.Time
		err error
	)
	if len(layout) != 0 {
		tm, err = time.Parse(layout[0], s.value)
	} else {
		tm, err = http.ParseTime(s.value)
	}

	if err != nil {
		s.chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{s.value},
			Errors: []error{
				errors.New("expected: string can be parsed to datetime"),
				err,
			},
		})
		return newDateTime(s.chain, time.Unix(0, 0))
	}

	return newDateTime(s.chain, tm)
}

// Empty succeeds if string is empty.
//
// Example:
//
//	str := NewString(t, "")
//	str.Empty()
func (s *String) Empty() *String {
	s.chain.enter("Empty()")
	defer s.chain.leave()

	if s.chain.failed() {
		return s
	}

	if !(s.value == "") {
		s.chain.fail(AssertionFailure{
			Type:   AssertEmpty,
			Actual: &AssertionValue{s.value},
			Errors: []error{
				errors.New("expected: string is empty"),
			},
		})
	}

	return s
}

// NotEmpty succeeds if string is non-empty.
//
// Example:
//
//	str := NewString(t, "Hello")
//	str.NotEmpty()
func (s *String) NotEmpty() *String {
	s.chain.enter("NotEmpty()")
	defer s.chain.leave()

	if s.chain.failed() {
		return s
	}

	if !(s.value != "") {
		s.chain.fail(AssertionFailure{
			Type:   AssertNotEmpty,
			Actual: &AssertionValue{s.value},
			Errors: []error{
				errors.New("expected: string is non-empty"),
			},
		})
	}

	return s
}

// Equal succeeds if string is equal to given Go string.
//
// Example:
//
//	str := NewString(t, "Hello")
//	str.Equal("Hello")
func (s *String) Equal(value string) *String {
	s.chain.enter("Equal()")
	defer s.chain.leave()

	if s.chain.failed() {
		return s
	}

	if !(s.value == value) {
		s.chain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{s.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: strings are equal"),
			},
		})
	}

	return s
}

// NotEqual succeeds if string is not equal to given Go string.
//
// Example:
//
//	str := NewString(t, "Hello")
//	str.NotEqual("Goodbye")
func (s *String) NotEqual(value string) *String {
	s.chain.enter("NotEqual()")
	defer s.chain.leave()

	if s.chain.failed() {
		return s
	}

	if !(s.value != value) {
		s.chain.fail(AssertionFailure{
			Type:     AssertNotEqual,
			Actual:   &AssertionValue{s.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: strings are non-equal"),
			},
		})
	}

	return s
}

// EqualFold succeeds if string is equal to given Go string after applying Unicode
// case-folding (so it's a case-insensitive match).
//
// Example:
//
//	str := NewString(t, "Hello")
//	str.EqualFold("hELLo")
func (s *String) EqualFold(value string) *String {
	s.chain.enter("EqualFold()")
	defer s.chain.leave()

	if s.chain.failed() {
		return s
	}

	if !strings.EqualFold(s.value, value) {
		s.chain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{s.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: strings are equal (if folded)"),
			},
		})
	}

	return s
}

// NotEqualFold succeeds if string is not equal to given Go string after applying
// Unicode case-folding (so it's a case-insensitive match).
//
// Example:
//
//	str := NewString(t, "Hello")
//	str.NotEqualFold("gOODBYe")
func (s *String) NotEqualFold(value string) *String {
	s.chain.enter("NotEqualFold()")
	defer s.chain.leave()

	if s.chain.failed() {
		return s
	}

	if strings.EqualFold(s.value, value) {
		s.chain.fail(AssertionFailure{
			Type:     AssertNotEqual,
			Actual:   &AssertionValue{s.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: strings are non-equal (if folded)"),
			},
		})
	}

	return s
}

// Contains succeeds if string contains given Go string as a substring.
//
// Example:
//
//	str := NewString(t, "Hello")
//	str.Contains("ell")
func (s *String) Contains(value string) *String {
	s.chain.enter("Contains()")
	defer s.chain.leave()

	if s.chain.failed() {
		return s
	}

	if !strings.Contains(s.value, value) {
		s.chain.fail(AssertionFailure{
			Type:     AssertContainsSubset,
			Actual:   &AssertionValue{s.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: string contains sub-string"),
			},
		})
	}

	return s
}

// NotContains succeeds if string doesn't contain Go string as a substring.
//
// Example:
//
//	str := NewString(t, "Hello")
//	str.NotContains("bye")
func (s *String) NotContains(value string) *String {
	s.chain.enter("NotContains()")
	defer s.chain.leave()

	if s.chain.failed() {
		return s
	}

	if strings.Contains(s.value, value) {
		s.chain.fail(AssertionFailure{
			Type:     AssertNotContainsSubset,
			Actual:   &AssertionValue{s.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: string does not contain sub-string"),
			},
		})
	}

	return s
}

// ContainsFold succeeds if string contains given Go string as a substring after
// applying Unicode case-folding (so it's a case-insensitive match).
//
// Example:
//
//	str := NewString(t, "Hello")
//	str.ContainsFold("ELL")
func (s *String) ContainsFold(value string) *String {
	s.chain.enter("ContainsFold()")
	defer s.chain.leave()

	if s.chain.failed() {
		return s
	}

	if !strings.Contains(strings.ToLower(s.value), strings.ToLower(value)) {
		s.chain.fail(AssertionFailure{
			Type:     AssertContainsSubset,
			Actual:   &AssertionValue{s.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: string contains sub-string (if folded)"),
			},
		})
	}

	return s
}

// NotContainsFold succeeds if string doesn't contain given Go string as a substring
// after applying Unicode case-folding (so it's a case-insensitive match).
//
// Example:
//
//	str := NewString(t, "Hello")
//	str.NotContainsFold("BYE")
func (s *String) NotContainsFold(value string) *String {
	s.chain.enter("NotContainsFold()")
	defer s.chain.leave()

	if s.chain.failed() {
		return s
	}

	if strings.Contains(strings.ToLower(s.value), strings.ToLower(value)) {
		s.chain.fail(AssertionFailure{
			Type:     AssertNotContainsSubset,
			Actual:   &AssertionValue{s.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: string does not contain sub-string (if folded)"),
			},
		})
	}

	return s
}

// Match matches the string with given regexp and returns a new Match instance
// with found submatches.
//
// If regexp is invalid or string doesn't match regexp, Match fails and returns
// empty (but non-nil) instance. regexp.Compile is used to construct regexp, and
// Regexp.FindStringSubmatch is used to construct matches.
//
// Example:
//
//	s := NewString(t, "http://example.com/users/john")
//	m := s.Match(`http://(?P<host>.+)/users/(?P<user>.+)`)
//
//	m.NotEmpty()
//	m.Length().Equal(3)
//
//	m.Index(0).Equal("http://example.com/users/john")
//	m.Index(1).Equal("example.com")
//	m.Index(2).Equal("john")
//
//	m.Name("host").Equal("example.com")
//	m.Name("user").Equal("john")
func (s *String) Match(re string) *Match {
	s.chain.enter("Match()")
	defer s.chain.leave()

	if s.chain.failed() {
		return newMatch(s.chain, nil, nil)
	}

	rx, err := regexp.Compile(re)
	if err != nil {
		s.chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{re},
			Errors: []error{
				errors.New("expected: valid regexp"),
				err,
			},
		})
		return newMatch(s.chain, nil, nil)
	}

	match := rx.FindStringSubmatch(s.value)
	if match == nil {
		s.chain.fail(AssertionFailure{
			Type:     AssertMatchRegexp,
			Actual:   &AssertionValue{s.value},
			Expected: &AssertionValue{re},
			Errors: []error{
				errors.New("expected: string matches regexp"),
			},
		})
		return newMatch(s.chain, nil, nil)
	}

	return newMatch(s.chain, match, rx.SubexpNames())
}

// NotMatch succeeds if the string doesn't match to given regexp.
//
// regexp.Compile is used to construct regexp, and Regexp.MatchString
// is used to perform match.
//
// Example:
//
//	s := NewString(t, "a")
//	s.NotMatch(`[^a]`)
func (s *String) NotMatch(re string) *String {
	s.chain.enter("NotMatch()")
	defer s.chain.leave()

	rx, err := regexp.Compile(re)
	if err != nil {
		s.chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{re},
			Errors: []error{
				errors.New("expected: valid regexp"),
				err,
			},
		})
		return s
	}

	if rx.MatchString(s.value) {
		s.chain.fail(AssertionFailure{
			Type:     AssertNotMatchRegexp,
			Actual:   &AssertionValue{s.value},
			Expected: &AssertionValue{re},
			Errors: []error{
				errors.New("expected: string does not match regexp"),
			},
		})
		return s
	}

	return s
}

// MatchAll find all matches in string for given regexp and returns a list
// of found matches.
//
// If regexp is invalid or string doesn't match regexp, MatchAll fails and
// returns empty (but non-nil) slice. regexp.Compile is used to construct
// regexp, and Regexp.FindAllStringSubmatch is used to find matches.
//
// Example:
//
//	s := NewString(t,
//	   "http://example.com/users/john http://example.com/users/bob")
//
//	m := s.MatchAll(`http://(?P<host>\S+)/users/(?P<user>\S+)`)
//
//	m[0].Name("user").Equal("john")
//	m[1].Name("user").Equal("bob")
func (s *String) MatchAll(re string) []Match {
	s.chain.enter("MatchAll()")
	defer s.chain.leave()

	if s.chain.failed() {
		return []Match{}
	}

	rx, err := regexp.Compile(re)
	if err != nil {
		s.chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{re},
			Errors: []error{
				errors.New("expected: valid regexp"),
				err,
			},
		})
		return []Match{}
	}

	matches := rx.FindAllStringSubmatch(s.value, -1)
	if matches == nil {
		s.chain.fail(AssertionFailure{
			Type:     AssertMatchRegexp,
			Actual:   &AssertionValue{s.value},
			Expected: &AssertionValue{re},
			Errors: []error{
				errors.New("expected: string matches regexp"),
			},
		})
		return []Match{}
	}

	ret := []Match{}
	for _, match := range matches {
		ret = append(ret, *newMatch(
			s.chain,
			match,
			rx.SubexpNames()))
	}

	return ret
}
