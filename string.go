package httpexpect

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
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

// HasPrefix succeeds if string has given Go string as prefix
//
// Example:
//
//	str := NewString(t, "Hello World")
//	str.HasPrefix("Hello")
func (s *String) HasPrefix(value string) *String {
	s.chain.enter("HasPrefix()")
	defer s.chain.leave()

	if s.chain.failed() {
		return s
	}

	if !strings.HasPrefix(s.value, value) {
		s.chain.fail(AssertionFailure{
			Type:     AssertContainsSubset,
			Actual:   &AssertionValue{s.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: string has prefix"),
			},
		})
	}

	return s
}

// NotHasPrefix succeeds if string doesn't have given Go string as prefix
//
// Example:
//
//	str := NewString(t, "Hello World")
//	str.NotHasPrefix("Bye")
func (s *String) NotHasPrefix(value string) *String {
	s.chain.enter("NotHasPrefix()")
	defer s.chain.leave()

	if s.chain.failed() {
		return s
	}

	if strings.HasPrefix(s.value, value) {
		s.chain.fail(AssertionFailure{
			Type:     AssertNotContainsSubset,
			Actual:   &AssertionValue{s.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: string doesn't have prefix"),
			},
		})
	}

	return s
}

// HasSuffix succeeds if string has given Go string as suffix
//
// Example:
//
//	str := NewString(t, "Hello World")
//	str.HasSuffix("World")
func (s *String) HasSuffix(value string) *String {
	s.chain.enter("HasSuffix()")
	defer s.chain.leave()

	if s.chain.failed() {
		return s
	}

	if !strings.HasSuffix(s.value, value) {
		s.chain.fail(AssertionFailure{
			Type:     AssertContainsSubset,
			Actual:   &AssertionValue{s.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: string has suffix"),
			},
		})
	}

	return s
}

// NotHasSuffix succeeds if string doesn't have given Go string as suffix
//
// Example:
//
//	str := NewString(t, "Hello World")
//	str.NotHasSuffix("Hello")
func (s *String) NotHasSuffix(value string) *String {
	s.chain.enter("NotHasSuffix()")
	defer s.chain.leave()

	if s.chain.failed() {
		return s
	}

	if strings.HasSuffix(s.value, value) {
		s.chain.fail(AssertionFailure{
			Type:     AssertNotContainsSubset,
			Actual:   &AssertionValue{s.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: string doesn't have suffix"),
			},
		})
	}

	return s
}

// HasPrefixFold succeeds if string has given Go string as prefix
// after applying Unicode case-folding (so it's a case-insensitive match).
//
// Example:
//
//	str := NewString(t, "Hello World")
//	str.HasPrefixFold("hello")
func (s *String) HasPrefixFold(value string) *String {
	s.chain.enter("HasPrefixFold()")
	defer s.chain.leave()

	if s.chain.failed() {
		return s
	}

	if !strings.HasPrefix(strings.ToLower(s.value), strings.ToLower(value)) {
		s.chain.fail(AssertionFailure{
			Type:     AssertContainsSubset,
			Actual:   &AssertionValue{s.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: string has prefix (if folded)"),
			},
		})
	}

	return s
}

// NotHasPrefixFold succeeds if string doesn't have given Go string as prefix
// after applying Unicode case-folding (so it's a case-insensitive match).
//
// Example:
//
//	str := NewString(t, "Hello World")
//	str.NotHasPrefixFold("Bye")
func (s *String) NotHasPrefixFold(value string) *String {
	s.chain.enter("NotHasPrefixFold()")
	defer s.chain.leave()

	if s.chain.failed() {
		return s
	}

	if strings.HasPrefix(strings.ToLower(s.value), strings.ToLower(value)) {
		s.chain.fail(AssertionFailure{
			Type:     AssertNotContainsSubset,
			Actual:   &AssertionValue{s.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: string doesn't have prefix (if folded)"),
			},
		})
	}

	return s
}

// HasSuffixFold succeeds if string has given Go string as suffix
// after applying Unicode case-folding (so it's a case-insensitive match).
//
// Example:
//
//	str := NewString(t, "Hello World")
//	str.HasSuffixFold("world")
func (s *String) HasSuffixFold(value string) *String {
	s.chain.enter("HasSuffixFold()")
	defer s.chain.leave()

	if s.chain.failed() {
		return s
	}

	if !strings.HasSuffix(strings.ToLower(s.value), strings.ToLower(value)) {
		s.chain.fail(AssertionFailure{
			Type:     AssertContainsSubset,
			Actual:   &AssertionValue{s.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: string has suffix (if folded)"),
			},
		})
	}

	return s
}

// NotHasSuffixFold succeeds if string doesn't have given Go string as suffix
// after applying Unicode case-folding (so it's a case-insensitive match).
//
// Example:
//
//	str := NewString(t, "Hello World")
//	str.NotHasSuffixFold("Bye")
func (s *String) NotHasSuffixFold(value string) *String {
	s.chain.enter("NotHasSuffix()")
	defer s.chain.leave()

	if s.chain.failed() {
		return s
	}

	if strings.HasSuffix(strings.ToLower(s.value), strings.ToLower(value)) {
		s.chain.fail(AssertionFailure{
			Type:     AssertNotContainsSubset,
			Actual:   &AssertionValue{s.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: string doesn't have suffix (if folded)"),
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

// IsASCII succeeds if all string characters belongs to ASCII.
//
// Example:
//
//	str := NewString(t, "Hello")
//	str.IsASCII()
func (s *String) IsASCII() *String {
	s.chain.enter("IsASCII()")
	defer s.chain.leave()

	if s.chain.failed() {
		return s
	}

	isASCII := true
	for _, c := range s.value {
		if c > unicode.MaxASCII {
			isASCII = false
			break
		}
	}

	if !isASCII {
		s.chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{s.value},
			Errors: []error{
				errors.New("expected: all string characters are ascii"),
			},
		})
	}

	return s
}

// NotIsASCII succeeds if at least one string character does not belong to ASCII.
//
// Example:
//
//	str := NewString(t, "こんにちは")
//	str.NotIsASCII()
func (s *String) NotIsASCII() *String {
	s.chain.enter("NotIsASCII()")
	defer s.chain.leave()

	if s.chain.failed() {
		return s
	}

	isASCII := true
	for _, c := range s.value {
		if c > unicode.MaxASCII {
			isASCII = false
			break
		}
	}

	if isASCII {
		s.chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{s.value},
			Errors: []error{
				errors.New("expected: at least one string character is not ascii"),
			},
		})
	}

	return s
}

// AsNumber parses float from string and returns a new Number instance
// with result.
//
// If base is 10 or omitted, uses strconv.ParseFloat.
// Otherwise, uses strconv.ParseInt or strconv.ParseUint with given base.
//
// Example:
//
//	str := NewString(t, "100")
//	str.AsNumber().Equal(100)
//
// Specifying base:
//
//	str.AsNumber(10).Equal(100)
//	str.AsNumber(16).Equal(256)
func (s *String) AsNumber(base ...int) *Number {
	s.chain.enter("AsNumber()")
	defer s.chain.leave()

	if s.chain.failed() {
		return newNumber(s.chain, 0)
	}

	if len(base) > 1 {
		s.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected multiple base arguments"),
			},
		})
		return newNumber(s.chain, 0)
	}

	b := 10
	if len(base) != 0 {
		b = base[0]
	}

	var fnum float64
	var inum int64
	var unum uint64
	var err error

	inum, err = strconv.ParseInt(s.value, b, 64)
	fnum = float64(inum)

	if err == nil && int64(fnum) != inum {
		s.chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{s.value},
			Errors: []error{
				errors.New("expected:" +
					" number can be represented as float64 without precision loss"),
			},
		})
		return newNumber(s.chain, 0)
	}

	if err != nil && errors.Is(err, strconv.ErrRange) {
		unum, err = strconv.ParseUint(s.value, b, 64)
		fnum = float64(unum)

		if err == nil && uint64(fnum) != unum {
			s.chain.fail(AssertionFailure{
				Type:   AssertValid,
				Actual: &AssertionValue{s.value},
				Errors: []error{
					errors.New("expected:" +
						" number can be represented as float64 without precision loss"),
				},
			})
			return newNumber(s.chain, 0)
		}
	}

	if err != nil && b == 10 {
		fnum, err = strconv.ParseFloat(s.value, 64)
	}

	if err != nil {
		if b == 10 {
			s.chain.fail(AssertionFailure{
				Type:   AssertValid,
				Actual: &AssertionValue{s.value},
				Errors: []error{
					errors.New("expected: string can be parsed to integer or float"),
					err,
				},
			})
		} else {
			s.chain.fail(AssertionFailure{
				Type:   AssertValid,
				Actual: &AssertionValue{s.value},
				Errors: []error{
					fmt.Errorf(
						"expected: string can be parsed to integer with base %d",
						base[0]),
					err,
				},
			})
		}
		return newNumber(s.chain, 0)
	}

	return newNumber(s.chain, fnum)
}

// AsBoolean parses true/false value string and returns a new Boolean instance
// with result.
//
// Accepts string values "true", "True", "false", "False".
//
// Example:
//
//	str := NewString(t, "true")
//	str.AsBoolean().True()
func (s *String) AsBoolean() *Boolean {
	s.chain.enter("AsBoolean()")
	defer s.chain.leave()

	if s.chain.failed() {
		return newBoolean(s.chain, false)
	}

	switch s.value {
	case "true", "True":
		return newBoolean(s.chain, true)

	case "false", "False":
		return newBoolean(s.chain, false)
	}

	s.chain.fail(AssertionFailure{
		Type:   AssertValid,
		Actual: &AssertionValue{s.value},
		Errors: []error{
			errors.New("expected: string can be parsed to boolean"),
		},
	})

	return newBoolean(s.chain, false)
}

// AsDateTime parses date/time from string and returns a new DateTime instance
// with result.
//
// If format is given, AsDateTime() uses time.Parse() with every given format.
// Otherwise, it uses the list of predefined common formats.
//
// If the string can't be parsed with any format, AsDateTime reports failure
// and returns empty (but non-nil) instance.
//
// Example:
//
//	str := NewString(t, "Tue, 15 Nov 1994 08:12:31 GMT")
//	str.AsDateTime().Lt(time.Now())
//
//	str := NewString(t, "15 Nov 94 08:12 GMT")
//	str.AsDateTime(time.RFC822).Lt(time.Now())
func (s *String) AsDateTime(format ...string) *DateTime {
	s.chain.enter("AsDateTime()")
	defer s.chain.leave()

	if s.chain.failed() {
		return newDateTime(s.chain, time.Unix(0, 0))
	}

	var formatList []datetimeFormat

	if len(format) != 0 {
		for _, f := range format {
			formatList = append(formatList, datetimeFormat{layout: f})
		}
	} else {
		formatList = []datetimeFormat{
			{http.TimeFormat, "RFC1123+GMT"},

			{time.RFC850, "RFC850"},

			{time.ANSIC, "ANSIC"},
			{time.UnixDate, "Unix"},
			{time.RubyDate, "Ruby"},

			{time.RFC1123, "RFC1123"},
			{time.RFC1123Z, "RFC1123Z"},
			{time.RFC822, "RFC822"},
			{time.RFC822Z, "RFC822Z"},
			{time.RFC3339, "RFC3339"},
			{time.RFC3339Nano, "RFC3339+nano"},
		}
	}

	var (
		tm  time.Time
		err error
	)
	for _, f := range formatList {
		tm, err = time.Parse(f.layout, s.value)
		if err == nil {
			break
		}
	}

	if err != nil {
		if len(formatList) == 1 {
			s.chain.fail(AssertionFailure{
				Type:     AssertMatchFormat,
				Actual:   &AssertionValue{s.value},
				Expected: &AssertionValue{formatList[0]},
				Errors: []error{
					errors.New("expected: string can be parsed to datetime" +
						" with given format"),
				},
			})
		} else {
			var expectedFormats []interface{}
			for _, f := range formatList {
				expectedFormats = append(expectedFormats, f)
			}
			s.chain.fail(AssertionFailure{
				Type:     AssertMatchFormat,
				Actual:   &AssertionValue{s.value},
				Expected: &AssertionValue{AssertionList(expectedFormats)},
				Errors: []error{
					errors.New("expected: string can be parsed to datetime" +
						" with one of the formats from list"),
				},
			})
		}
		return newDateTime(s.chain, time.Unix(0, 0))
	}

	return newDateTime(s.chain, tm)
}

type datetimeFormat struct {
	layout string
	name   string
}

func (f datetimeFormat) String() string {
	if f.name != "" {
		return fmt.Sprintf("%q (%s)", f.layout, f.name)
	} else {
		return fmt.Sprintf("%q", f.layout)
	}
}

// Deprecated: use AsNumber instead.
func (s *String) Number() *Number {
	return s.AsNumber()
}

// Deprecated: use AsDateTime instead.
func (s *String) DateTime(layout ...string) *DateTime {
	return s.AsDateTime(layout...)
}
