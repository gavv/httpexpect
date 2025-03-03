package httpexpect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatch_FailedChain(t *testing.T) {
	chain := newMockChain(t, flagFailed)

	value := newMatch(chain, nil, nil)
	value.chain.assert(t, failure)

	value.Alias("foo")

	value.Length().chain.assert(t, failure)
	value.Submatch(0).chain.assert(t, failure)
	value.NamedSubmatch("").chain.assert(t, failure)

	value.IsEmpty()
	value.NotEmpty()
	value.HasSubmatches("")
	value.NotHasSubmatches("")
}

func TestMatch_Constructors(t *testing.T) {
	matches := []string{"m0", "m1", "m2"}
	names := []string{"", "n1", "n2"}

	t.Run("reporter", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewMatch(reporter, matches, names)
		assert.Equal(t, matches, value.Raw())
		value.chain.assert(t, success)
	})

	t.Run("config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewMatchC(Config{
			Reporter: reporter,
		}, matches, names)
		assert.Equal(t, matches, value.Raw())
		value.chain.assert(t, success)
	})

	t.Run("chain", func(t *testing.T) {
		chain := newMockChain(t)
		value := newMatch(chain, matches, names)
		assert.NotSame(t, value.chain, chain)
		assert.Equal(t, value.chain.context.Path, chain.context.Path)
	})
}

func TestMatch_Raw(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewMatch(reporter, nil, nil)

		assert.NotNil(t, []string{}, value.Raw())
		assert.Equal(t, []string{}, value.Raw())
		value.chain.assert(t, success)
	})

	t.Run("non-nil", func(t *testing.T) {
		reporter := newMockReporter(t)

		data := []string{"foo", "bar"}

		value := NewMatch(reporter, data, nil)

		assert.Equal(t, data, value.Raw())
		assert.NotSame(t, &data[0], &value.Raw()[0])
		value.chain.assert(t, success)
	})
}

func TestMatch_Alias(t *testing.T) {
	reporter := newMockReporter(t)

	matches := []string{"m0", "m1", "m2"}
	names := []string{"", "n1", "n2"}

	value := NewMatch(reporter, matches, names)
	assert.Equal(t, []string{"Match()"}, value.chain.context.Path)
	assert.Equal(t, []string{"Match()"}, value.chain.context.AliasedPath)

	value.Alias("foo")
	assert.Equal(t, []string{"Match()"}, value.chain.context.Path)
	assert.Equal(t, []string{"foo"}, value.chain.context.AliasedPath)

	childValue := value.Submatch(0)
	assert.Equal(t, []string{"Match()", "Submatch(0)"}, childValue.chain.context.Path)
	assert.Equal(t, []string{"foo", "Submatch(0)"}, childValue.chain.context.AliasedPath)
}

func TestMatch_Getters(t *testing.T) {
	matches := []string{"m0", "m1", "m2"}
	names := []string{"", "n1", "n2"}

	t.Run("length", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewMatch(reporter, matches, names)

		innerValue := value.Length()
		assert.Equal(t, 3.0, innerValue.Raw())

		value.chain.assert(t, success)
		innerValue.chain.assert(t, success)
	})

	t.Run("submatch", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewMatch(reporter, matches, names)

		for n := range matches {
			innerValue := value.Submatch(n)
			assert.Equal(t, matches[n], innerValue.Raw())

			value.chain.assert(t, success)
			innerValue.chain.assert(t, success)
		}
	})

	t.Run("submatch out of range", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewMatch(reporter, matches, names)

		innerValue := value.Submatch(3)
		assert.NotNil(t, innerValue)

		value.chain.assert(t, failure)
		innerValue.chain.assert(t, failure)
	})

	t.Run("named submatch", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewMatch(reporter, matches, names)

		for n := range matches {
			if names[n] == "" {
				continue
			}
			innerValue := value.NamedSubmatch(names[n])
			assert.Equal(t, matches[n], innerValue.Raw())

			value.chain.assert(t, success)
			innerValue.chain.assert(t, success)
		}
	})

	t.Run("named submatch bad key", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewMatch(reporter, matches, names)

		innerValue := value.NamedSubmatch("bad_key")
		assert.NotNil(t, innerValue)

		value.chain.assert(t, failure)
		innerValue.chain.assert(t, failure)
	})
}

func TestMatch_IsEmpty(t *testing.T) {
	cases := []struct {
		name      string
		submatch  []string
		wantEmpty chainResult
	}{
		{
			name:      "string",
			submatch:  []string{"m"},
			wantEmpty: failure,
		},
		{
			name:      "empty string slice",
			submatch:  []string{},
			wantEmpty: success,
		},
		{
			name:      "nil",
			submatch:  nil,
			wantEmpty: success,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewMatch(reporter, tc.submatch, nil).IsEmpty().
				chain.assert(t, tc.wantEmpty)

			NewMatch(reporter, tc.submatch, nil).NotEmpty().
				chain.assert(t, !tc.wantEmpty)

			if tc.wantEmpty {
				assert.Equal(t, []string{},
					NewMatch(reporter, tc.submatch, nil).Raw())
			}
		})
	}
}

func TestMatch_Values(t *testing.T) {
	type wantMatch struct {
		target []string
		result chainResult
	}

	cases := []struct {
		name       string
		submatches []string
		wantMatch  []wantMatch
	}{
		{
			name:       "nil match instance",
			submatches: nil,
			wantMatch: []wantMatch{
				{target: nil, result: success},
				{target: []string{""}, result: failure},
			},
		},
		{
			name:       "empty match instance",
			submatches: []string{},
			wantMatch: []wantMatch{
				{target: nil, result: success},
				{target: []string{""}, result: failure},
			},
		},
		{
			name:       "not empty index 0 only",
			submatches: []string{"m0"},
			wantMatch: []wantMatch{
				{target: nil, result: success},
				{target: []string{"m0"}, result: failure},
			},
		},
		{
			name:       "not empty",
			submatches: []string{"m0", "m1", "m2"},
			wantMatch: []wantMatch{
				{target: nil, result: failure},
				{target: []string{"m1"}, result: failure},
				{target: []string{"m2", "m1"}, result: failure},
				{target: []string{"m1", "m2"}, result: success},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			for _, match := range tc.wantMatch {
				NewMatch(reporter, tc.submatches, nil).HasSubmatches(match.target...).
					chain.assert(t, match.result)

				NewMatch(reporter, tc.submatches, nil).NotHasSubmatches(match.target...).
					chain.assert(t, !match.result)
			}
		})
	}
}
