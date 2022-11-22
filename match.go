package httpexpect

import (
	"errors"
	"reflect"
)

// Match provides methods to inspect attached regexp match results.
type Match struct {
	chain      *chain
	submatches []string
	names      map[string]int
}

// NewMatch returns a new Match instance.
//
// reporter should not be nil. submatches and names may be nil.
//
// Example:
//
//	s := "http://example.com/users/john"
//	r := regexp.MustCompile(`http://(?P<host>.+)/users/(?P<user>.+)`)
//	m := NewMatch(reporter, r.FindStringSubmatch(s), r.SubexpNames())
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
func NewMatch(reporter Reporter, submatches []string, names []string) *Match {
	return newMatch(newChainWithDefaults("Match()", reporter), submatches, names)
}

func newMatch(parent *chain, matchList []string, nameList []string) *Match {
	m := &Match{parent.clone(), nil, nil}

	if matchList != nil {
		m.submatches = matchList
	} else {
		m.submatches = []string{}
	}

	m.names = map[string]int{}
	for n, name := range nameList {
		if name != "" {
			m.names[name] = n
		}
	}

	return m
}

// Raw returns underlying submatches attached to Match.
// This is the value originally passed to NewMatch.
//
// Example:
//
//	m := NewMatch(t, submatches, names)
//	assert.Equal(t, submatches, m.Raw())
func (m *Match) Raw() []string {
	return m.submatches
}

// Length returns a new Number instance with number of submatches.
//
// Example:
//
//	m := NewMatch(t, submatches, names)
//	m.Length().Equal(len(submatches))
func (m *Match) Length() *Number {
	m.chain.enter("Length()")
	defer m.chain.leave()

	if m.chain.failed() {
		return newNumber(m.chain, 0)
	}

	return newNumber(m.chain, float64(len(m.submatches)))
}

// Index returns a new String instance with submatch for given index.
//
// Note that submatch with index 0 contains the whole match. If index is out
// of bounds, Index reports failure and returns empty (but non-nil) instance.
//
// Example:
//
//	s := "http://example.com/users/john"
//
//	r := regexp.MustCompile(`http://(.+)/users/(.+)`)
//	m := NewMatch(t, r.FindStringSubmatch(s), nil)
//
//	m.Index(0).Equal("http://example.com/users/john")
//	m.Index(1).Equal("example.com")
//	m.Index(2).Equal("john")
func (m *Match) Index(index int) *String {
	m.chain.enter("Index(%d)", index)
	defer m.chain.leave()

	if m.chain.failed() {
		return newString(m.chain, "")
	}

	if index < 0 || index >= len(m.submatches) {
		m.chain.fail(AssertionFailure{
			Type:   AssertInRange,
			Actual: &AssertionValue{index},
			Expected: &AssertionValue{AssertionRange{
				Min: 0,
				Max: len(m.submatches) - 1,
			}},
			Errors: []error{
				errors.New("expected: valid sub-match index"),
			},
		})
		return newString(m.chain, "")
	}

	return newString(m.chain, m.submatches[index])
}

// Name returns a new String instance with submatch for given name.
//
// If there is no submatch with given name, Name reports failure and returns
// empty (but non-nil) instance.
//
// Example:
//
//	s := "http://example.com/users/john"
//
//	r := regexp.MustCompile(`http://(?P<host>.+)/users/(?P<user>.+)`)
//	m := NewMatch(t, r.FindStringSubmatch(s), r.SubexpNames())
//
//	m.Name("host").Equal("example.com")
//	m.Name("user").Equal("john")
func (m *Match) Name(name string) *String {
	m.chain.enter("Name(%q)", name)
	defer m.chain.leave()

	if m.chain.failed() {
		return newString(m.chain, "")
	}

	index, ok := m.names[name]
	if !ok {
		names := make([]interface{}, 0, len(m.names))
		for nm := range m.names {
			names = append(names, nm)
		}
		m.chain.fail(AssertionFailure{
			Type:     AssertBelongs,
			Actual:   &AssertionValue{name},
			Expected: &AssertionValue{AssertionList(names)},
			Errors: []error{
				errors.New("expected: existing sub-match name"),
			},
		})
		return newString(m.chain, "")
	}

	return newString(m.chain, m.submatches[index])
}

// Empty succeeds if submatches array is empty.
//
// Example:
//
//	m := NewMatch(t, submatches, names)
//	m.Empty()
func (m *Match) Empty() *Match {
	m.chain.enter("Empty()")
	defer m.chain.leave()

	if m.chain.failed() {
		return m
	}

	if !(len(m.submatches) == 0) {
		m.chain.fail(AssertionFailure{
			Type:   AssertEmpty,
			Actual: &AssertionValue{m.submatches},
			Errors: []error{
				errors.New("expected: empty sub-match list"),
			},
		})
	}

	return m
}

// NotEmpty succeeds if submatches array is non-empty.
//
// Example:
//
//	m := NewMatch(t, submatches, names)
//	m.NotEmpty()
func (m *Match) NotEmpty() *Match {
	m.chain.enter("NotEmpty()")
	defer m.chain.leave()

	if m.chain.failed() {
		return m
	}

	if !(len(m.submatches) != 0) {
		m.chain.fail(AssertionFailure{
			Type:   AssertNotEmpty,
			Actual: &AssertionValue{m.submatches},
			Errors: []error{
				errors.New("expected: non-empty sub-match list"),
			},
		})
	}

	return m
}

// Values succeeds if submatches array, starting from index 1, is equal to
// given array.
//
// Note that submatch with index 0 contains the whole match and is not
// included into this check.
//
// Example:
//
//	s := "http://example.com/users/john"
//	r := regexp.MustCompile(`http://(.+)/users/(.+)`)
//	m := NewMatch(t, r.FindStringSubmatch(s), nil)
//	m.Values("example.com", "john")
func (m *Match) Values(values ...string) *Match {
	m.chain.enter("Values()")
	defer m.chain.leave()

	if m.chain.failed() {
		return m
	}

	if values == nil {
		values = []string{}
	}

	if !reflect.DeepEqual(values, m.getValues()) {
		m.chain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{m.submatches},
			Expected: &AssertionValue{values},
			Errors: []error{
				errors.New("expected: sub-match lists are equal"),
			},
		})
	}

	return m
}

// NotValues succeeds if submatches array, starting from index 1, is not
// equal to given array.
//
// Note that submatch with index 0 contains the whole match and is not
// included into this check.
//
// Example:
//
//	s := "http://example.com/users/john"
//	r := regexp.MustCompile(`http://(.+)/users/(.+)`)
//	m := NewMatch(t, r.FindStringSubmatch(s), nil)
//	m.NotValues("example.com", "bob")
func (m *Match) NotValues(values ...string) *Match {
	m.chain.enter("NotValues()")
	defer m.chain.leave()

	if values == nil {
		values = []string{}
	}

	if reflect.DeepEqual(values, m.getValues()) {
		m.chain.fail(AssertionFailure{
			Type:     AssertNotEqual,
			Actual:   &AssertionValue{m.submatches},
			Expected: &AssertionValue{values},
			Errors: []error{
				errors.New("expected: sub-match lists are non-equal"),
			},
		})
	}

	return m
}

func (m *Match) getValues() []string {
	if len(m.submatches) > 1 {
		return m.submatches[1:]
	}
	return []string{}
}
