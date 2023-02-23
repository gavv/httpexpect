package httpexpect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatch_FailedChain(t *testing.T) {
	chain := newMockChain(t)
	chain.setFailed()

	value := newMatch(chain, nil, nil)
	value.chain.assertFailed(t)

	value.Alias("foo")

	value.Length().chain.assertFailed(t)
	value.Index(0).chain.assertFailed(t)
	value.Name("").chain.assertFailed(t)

	value.IsEmpty()
	value.NotEmpty()
	value.Values("")
	value.NotValues("")
}

func TestMatch_Constructors(t *testing.T) {
	matches := []string{"m0", "m1", "m2"}
	names := []string{"", "n1", "n2"}

	t.Run("reporter", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewMatch(reporter, matches, names)
		assert.Equal(t, matches, value.Raw())
		value.chain.assertNotFailed(t)
	})

	t.Run("config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewMatchC(Config{
			Reporter: reporter,
		}, matches, names)
		assert.Equal(t, matches, value.Raw())
		value.chain.assertNotFailed(t)
	})

	t.Run("chain", func(t *testing.T) {
		chain := newMockChain(t)
		value := newMatch(chain, matches, names)
		assert.NotSame(t, value.chain, chain)
		assert.Equal(t, value.chain.context.Path, chain.context.Path)
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

	childValue := value.Index(0)
	assert.Equal(t, []string{"Match()", "Index(0)"}, childValue.chain.context.Path)
	assert.Equal(t, []string{"foo", "Index(0)"}, childValue.chain.context.AliasedPath)
}

func TestMatch_Getters(t *testing.T) {
	reporter := newMockReporter(t)

	matches := []string{"m0", "m1", "m2"}
	names := []string{"", "n1", "n2"}

	value := NewMatch(reporter, matches, names)

	assert.Equal(t, matches, value.Raw())

	assert.Equal(t, 3.0, value.Length().Raw())

	assert.Equal(t, "m0", value.Index(0).Raw())
	assert.Equal(t, "m1", value.Index(1).Raw())
	assert.Equal(t, "m2", value.Index(2).Raw())
	value.chain.assertNotFailed(t)

	assert.Equal(t, "m1", value.Name("n1").Raw())
	assert.Equal(t, "m2", value.Name("n2").Raw())
	value.chain.assertNotFailed(t)

	assert.Equal(t, "", value.Index(-1).Raw())
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	assert.Equal(t, "", value.Index(3).Raw())
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	assert.Equal(t, "", value.Name("").Raw())
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	assert.Equal(t, "", value.Name("bad").Raw())
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestMatch_IsEmpty(t *testing.T) {
	cases := map[string]struct {
		submatch    []string
		expectEmpty bool
	}{
		"string":             {submatch: []string{"m"}, expectEmpty: false},
		"empty string slice": {submatch: []string{}, expectEmpty: true},
		"nil":                {submatch: nil, expectEmpty: true},
	}

	for name, instance := range cases {
		t.Run(name, func(t *testing.T) {
			reporter := newMockReporter(t)

			if instance.expectEmpty {
				NewMatch(reporter, instance.submatch, nil).IsEmpty().
					chain.assertNotFailed(t)

				NewMatch(reporter, instance.submatch, nil).NotEmpty().
					chain.assertFailed(t)

				assert.Equal(t, []string{}, NewMatch(reporter, instance.submatch, nil).Raw())
			} else {
				NewMatch(reporter, instance.submatch, nil).NotEmpty().
					chain.assertNotFailed(t)

				NewMatch(reporter, instance.submatch, nil).IsEmpty().
					chain.assertFailed(t)
			}
		})
	}
}

func TestMatch_Values(t *testing.T) {
	type wantMatch struct {
		target []string
		fail   bool
	}

	cases := map[string]struct {
		submatches  []string
		expectMatch []wantMatch
	}{
		"nil match instance": {
			submatches: nil,
			expectMatch: []wantMatch{
				{target: nil, fail: false},
				{target: []string{""}, fail: true},
			},
		},
		"empty match instance": {
			submatches: []string{},
			expectMatch: []wantMatch{
				{target: nil, fail: false},
				{target: []string{""}, fail: true},
			},
		},
		"not empty index 0 only": {
			submatches: []string{"m0"},
			expectMatch: []wantMatch{
				{target: nil, fail: false},
				{target: []string{"m0"}, fail: true},
			},
		},
		"not empty": {
			submatches: []string{"m0", "m1", "m2"},
			expectMatch: []wantMatch{
				{target: nil, fail: true},
				{target: []string{"m1"}, fail: true},
				{target: []string{"m2", "m1"}, fail: true},
				{target: []string{"m1", "m2"}, fail: false},
			},
		},
	}

	for name, instance := range cases {
		t.Run(name, func(t *testing.T) {
			reporter := newMockReporter(t)

			for _, match := range instance.expectMatch {
				if match.fail {
					NewMatch(reporter, instance.submatches, nil).NotValues(match.target...).
						chain.assertNotFailed(t)

					NewMatch(reporter, instance.submatches, nil).Values(match.target...).
						chain.assertFailed(t)

				} else {
					NewMatch(reporter, instance.submatches, nil).Values(match.target...).
						chain.assertNotFailed(t)

					NewMatch(reporter, instance.submatches, nil).NotValues(match.target...).
						chain.assertFailed(t)
				}
			}
		})
	}
}
