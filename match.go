package httpexpect

import (
	"errors"
	"reflect"
)

// Match provides methods to inspect attached regexp match results.
type Match struct {
	chain          *chain
	submatchValues []string
	submatchNames  map[string]int
}

// NewMatch returns a new Match instance.
//
// If reporter is nil, the function panics.
// Both submatchValues and submatchNames may be nil.
//
// Example:
//
//	s := "http://example.com/users/john"
//	r := regexp.MustCompile(`http://(?P<host>.+)/users/(?P<user>.+)`)
//
//	m := NewMatch(t, r.FindStringSubmatch(s), r.SubexpNames())
//
//	m.NotEmpty()
//	m.Length().IsEqual(3)
//
//	m.Submatch(0).IsEqual("http://example.com/users/john")
//	m.Submatch(1).IsEqual("example.com")
//	m.Submatch(2).IsEqual("john")
//
//	m.NamedSubmatch("host").IsEqual("example.com")
//	m.NamedSubmatch("user").IsEqual("john")
func NewMatch(reporter Reporter, submatchValues []string, submatchNames []string) *Match {
	return newMatch(
		newChainWithDefaults("Match()", reporter), submatchValues, submatchNames)
}

// NewMatchC returns a new Match instance with config.
//
// Requirements for config are same as for WithConfig function.
// Both submatches and names may be nil.
//
// See NewMatch for usage example.
func NewMatchC(config Config, submatchValues []string, submatchNames []string) *Match {
	return newMatch(
		newChainWithConfig("Match()", config.withDefaults()), submatchValues, submatchNames)
}

func newMatch(parent *chain, submatchValues []string, submatchNames []string) *Match {
	m := &Match{parent.clone(), nil, nil}

	if submatchValues != nil {
		m.submatchValues = make([]string, len(submatchValues))
		copy(m.submatchValues, submatchValues)
	} else {
		m.submatchValues = []string{}
	}

	m.submatchNames = map[string]int{}
	for n, name := range submatchNames {
		if name != "" {
			m.submatchNames[name] = n
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
	return m.submatchValues
}

// Alias is similar to Value.Alias.
func (m *Match) Alias(name string) *Match {
	opChain := m.chain.enter("Alias(%q)", name)
	defer opChain.leave()

	m.chain.setAlias(name)
	return m
}

// Length returns a new Number instance with number of submatches.
//
// Example:
//
//	m := NewMatch(t, submatches, names)
//	m.Length().IsEqual(len(submatches))
func (m *Match) Length() *Number {
	opChain := m.chain.enter("Length()")
	defer opChain.leave()

	if opChain.failed() {
		return newNumber(opChain, 0)
	}

	return newNumber(opChain, float64(len(m.submatchValues)))
}

// Submatch returns a new String instance with submatch for given index.
//
// Note that submatch with index 0 contains the whole match. If index is out
// of bounds, Submatch reports failure and returns empty (but non-nil) instance.
//
// Example:
//
//	s := "http://example.com/users/john"
//
//	r := regexp.MustCompile(`http://(.+)/users/(.+)`)
//	m := NewMatch(t, r.FindStringSubmatch(s), nil)
//
//	m.Submatch(0).IsEqual("http://example.com/users/john")
//	m.Submatch(1).IsEqual("example.com")
//	m.Submatch(2).IsEqual("john")
func (m *Match) Submatch(index int) *String {
	opChain := m.chain.enter("Submatch(%d)", index)
	defer opChain.leave()

	if opChain.failed() {
		return newString(opChain, "")
	}

	if index < 0 || index >= len(m.submatchValues) {
		opChain.fail(AssertionFailure{
			Type:   AssertInRange,
			Actual: &AssertionValue{index},
			Expected: &AssertionValue{AssertionRange{
				Min: 0,
				Max: len(m.submatchValues) - 1,
			}},
			Errors: []error{
				errors.New("expected: valid sub-match index"),
			},
		})
		return newString(opChain, "")
	}

	return newString(opChain, m.submatchValues[index])
}

// NamedSubmatch returns a new String instance with submatch for given name.
//
// If there is no submatch with given name, NamedSubmatch reports failure and returns
// empty (but non-nil) instance.
//
// Example:
//
//	s := "http://example.com/users/john"
//
//	r := regexp.MustCompile(`http://(?P<host>.+)/users/(?P<user>.+)`)
//	m := NewMatch(t, r.FindStringSubmatch(s), r.SubexpNames())
//
//	m.NamedSubmatch("host").IsEqual("example.com")
//	m.NamedSubmatch("user").IsEqual("john")
func (m *Match) NamedSubmatch(name string) *String {
	opChain := m.chain.enter("NamedSubmatch(%q)", name)
	defer opChain.leave()

	if opChain.failed() {
		return newString(opChain, "")
	}

	index, ok := m.submatchNames[name]
	if !ok {
		nameList := make([]interface{}, 0, len(m.submatchNames))
		for n := range m.submatchNames {
			nameList = append(nameList, n)
		}

		opChain.fail(AssertionFailure{
			Type:     AssertBelongs,
			Actual:   &AssertionValue{name},
			Expected: &AssertionValue{AssertionList(nameList)},
			Errors: []error{
				errors.New("expected: existing sub-match name"),
			},
		})

		return newString(opChain, "")
	}

	return newString(opChain, m.submatchValues[index])
}

// Deprecated: use Submatch instead.
func (m *Match) Index(index int) *String {
	return m.Submatch(index)
}

// Deprecated: use NamedSubmatch instead.
func (m *Match) Name(name string) *String {
	return m.NamedSubmatch(name)
}

// IsEmpty succeeds if submatches array is empty.
//
// Example:
//
//	m := NewMatch(t, submatches, names)
//	m.IsEmpty()
func (m *Match) IsEmpty() *Match {
	opChain := m.chain.enter("IsEmpty()")
	defer opChain.leave()

	if opChain.failed() {
		return m
	}

	if !(len(m.submatchValues) == 0) {
		opChain.fail(AssertionFailure{
			Type:   AssertEmpty,
			Actual: &AssertionValue{m.submatchValues},
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
	opChain := m.chain.enter("NotEmpty()")
	defer opChain.leave()

	if opChain.failed() {
		return m
	}

	if !(len(m.submatchValues) != 0) {
		opChain.fail(AssertionFailure{
			Type:   AssertNotEmpty,
			Actual: &AssertionValue{m.submatchValues},
			Errors: []error{
				errors.New("expected: non-empty sub-match list"),
			},
		})
	}

	return m
}

// Deprecated: use IsEmpty instead.
func (m *Match) Empty() *Match {
	return m.IsEmpty()
}

// HasSubmatches succeeds if submatches array, starting from index 1, is equal to
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
//	m.HasSubmatches("example.com", "john")
func (m *Match) HasSubmatches(submatchValues ...string) *Match {
	opChain := m.chain.enter("HasSubmatches()")
	defer opChain.leave()

	if opChain.failed() {
		return m
	}

	if submatchValues == nil {
		submatchValues = []string{}
	}

	if !reflect.DeepEqual(submatchValues, m.getValues()) {
		opChain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{m.submatchValues},
			Expected: &AssertionValue{submatchValues},
			Errors: []error{
				errors.New("expected: sub-match lists are equal"),
			},
		})
	}

	return m
}

// NotHasSubmatches succeeds if submatches array, starting from index 1, is not
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
//	m.NotHasSubmatches("example.com", "bob")
func (m *Match) NotHasSubmatches(submatchValues ...string) *Match {
	opChain := m.chain.enter("NotHasSubmatches()")
	defer opChain.leave()

	if submatchValues == nil {
		submatchValues = []string{}
	}

	if reflect.DeepEqual(submatchValues, m.getValues()) {
		opChain.fail(AssertionFailure{
			Type:     AssertNotEqual,
			Actual:   &AssertionValue{m.submatchValues},
			Expected: &AssertionValue{submatchValues},
			Errors: []error{
				errors.New("expected: sub-match lists are non-equal"),
			},
		})
	}

	return m
}

// Deprecated: use HasSubmatches instead.
func (m *Match) Values(submatchValues ...string) *Match {
	return m.HasSubmatches(submatchValues...)
}

// Deprecated: use NotHasSubmatches instead.
func (m *Match) NotValues(submatchValues ...string) *Match {
	return m.NotHasSubmatches(submatchValues...)
}

func (m *Match) getValues() []string {
	if len(m.submatchValues) > 1 {
		return m.submatchValues[1:]
	}
	return []string{}
}
