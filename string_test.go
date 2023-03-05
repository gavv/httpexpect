package httpexpect

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestString_FailedChain(t *testing.T) {
	chain := newMockChain(t)
	chain.setFailed()

	value := newString(chain, "")
	value.chain.assertFailed(t)

	value.Path("$").chain.assertFailed(t)
	value.Schema("")
	value.Alias("foo")

	var target interface{}
	value.Decode(target)

	value.Length().chain.assertFailed(t)

	value.IsEmpty()
	value.NotEmpty()
	value.IsEqual("")
	value.NotEqual("")
	value.IsEqualFold("")
	value.NotEqualFold("")
	value.InList("")
	value.NotInList("")
	value.InListFold("")
	value.NotInListFold("")
	value.Contains("")
	value.NotContains("")
	value.ContainsFold("")
	value.NotContainsFold("")
	value.HasPrefix("")
	value.NotHasPrefix("")
	value.HasSuffix("")
	value.NotHasSuffix("")
	value.HasPrefixFold("")
	value.NotHasPrefixFold("")
	value.HasSuffixFold("")
	value.NotHasSuffixFold("")
	value.IsASCII()
	value.NotASCII()

	value.Match("").chain.assertFailed(t)
	value.NotMatch("")
	assert.NotNil(t, value.MatchAll(""))
	assert.Equal(t, 0, len(value.MatchAll("")))

	value.AsBoolean().chain.assertFailed(t)
	value.AsNumber().chain.assertFailed(t)
	value.AsDateTime().chain.assertFailed(t)
}

func TestString_Constructors(t *testing.T) {
	t.Run("reporter", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewString(reporter, "Hello")
		value.IsEqual("Hello")
		value.chain.assertNotFailed(t)
	})

	t.Run("config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewStringC(Config{
			Reporter: reporter,
		}, "Hello")
		value.IsEqual("Hello")
		value.chain.assertNotFailed(t)
	})

	t.Run("chain", func(t *testing.T) {
		chain := newMockChain(t)
		value := newString(chain, "Hello")
		assert.NotSame(t, value.chain, chain)
		assert.Equal(t, value.chain.context.Path, chain.context.Path)
	})
}

func TestString_Decode(t *testing.T) {
	t.Run("target is empty interface", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewString(reporter, "foo")

		var target interface{}
		value.Decode(&target)

		value.chain.assertNotFailed(t)
		assert.Equal(t, "foo", target)
	})

	t.Run("target is string", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewString(reporter, "foo")

		var target string
		value.Decode(&target)

		value.chain.assertNotFailed(t)
		assert.Equal(t, "foo", target)
	})

	t.Run("target is unmarshable", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewString(reporter, "foo")

		value.Decode(123)

		value.chain.assertFailed(t)
	})

	t.Run("target is nil", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewString(reporter, "foo")

		value.Decode(nil)

		value.chain.assertFailed(t)
	})
}

func TestString_Alias(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "123")
	assert.Equal(t, []string{"String()"}, value.chain.context.Path)
	assert.Equal(t, []string{"String()"}, value.chain.context.AliasedPath)

	value.Alias("foo")
	assert.Equal(t, []string{"String()"}, value.chain.context.Path)
	assert.Equal(t, []string{"foo"}, value.chain.context.AliasedPath)

	childValue := value.AsNumber()
	assert.Equal(t, []string{"String()", "AsNumber()"}, childValue.chain.context.Path)
	assert.Equal(t, []string{"foo", "AsNumber()"}, childValue.chain.context.AliasedPath)
}

func TestString_Getters(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "foo")

	assert.Equal(t, "foo", value.Raw())
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	assert.Equal(t, "foo", value.Path("$").Raw())
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Schema(`{"type": "string"}`)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Schema(`{"type": "object"}`)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	num := value.Length()
	value.chain.assertNotFailed(t)
	num.chain.assertNotFailed(t)
	assert.Equal(t, 3.0, num.Raw())
}

func TestString_IsEmpty(t *testing.T) {
	cases := []struct {
		name      string
		str       string
		wantEmpty bool
	}{
		{
			name:      "empty string",
			str:       "",
			wantEmpty: true,
		},
		{
			name:      "populated string",
			str:       "a",
			wantEmpty: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			if tc.wantEmpty {
				NewString(reporter, tc.str).IsEmpty().
					chain.assertNotFailed(t)

				NewString(reporter, tc.str).NotEmpty().
					chain.assertFailed(t)
			} else {
				NewString(reporter, tc.str).IsEmpty().
					chain.assertFailed(t)

				NewString(reporter, tc.str).NotEmpty().
					chain.assertNotFailed(t)
			}
		})
	}
}

func TestString_IsEqual(t *testing.T) {
	cases := []struct {
		name    string
		str     string
		value   string
		isEqual bool
	}{
		{
			name:    "equivalent string",
			str:     "foo",
			value:   "foo",
			isEqual: true,
		},
		{
			name:    "non-equivalent string",
			str:     "foo",
			value:   "FOO",
			isEqual: false,
		},
	}

	for _, tc := range cases {
		reporter := newMockReporter(t)

		if tc.isEqual {
			NewString(reporter, tc.str).
				IsEqual(tc.value).
				chain.assertNotFailed(t)
			NewString(reporter, tc.str).
				NotEqual(tc.value).
				chain.assertFailed(t)
		} else {
			NewString(reporter, tc.str).
				NotEqual(tc.value).
				chain.assertNotFailed(t)
			NewString(reporter, tc.str).
				IsEqual(tc.value).
				chain.assertFailed(t)
		}
	}
}

func TestString_IsEqualFold(t *testing.T) {
	cases := []struct {
		name        string
		str         string
		value       string
		isEqualFold bool
	}{
		{
			name:        "equivalent string",
			str:         "foo",
			value:       "foo",
			isEqualFold: true,
		},
		{
			name:        "equivalent string with case-insensitive match",
			str:         "foo",
			value:       "FOO",
			isEqualFold: true,
		},
		{
			name:        "non-equivalent string",
			str:         "foo",
			value:       "foo2",
			isEqualFold: false,
		},
	}

	for _, tc := range cases {
		reporter := newMockReporter(t)

		if tc.isEqualFold {
			NewString(reporter, tc.str).
				IsEqualFold(tc.value).
				chain.assertNotFailed(t)
			NewString(reporter, tc.str).
				NotEqualFold(tc.value).
				chain.assertFailed(t)
		} else {
			NewString(reporter, tc.str).
				NotEqualFold(tc.value).
				chain.assertNotFailed(t)
			NewString(reporter, tc.str).
				IsEqualFold(tc.value).
				chain.assertFailed(t)
		}
	}
}

func TestString_InList(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		cases := []struct {
			name     string
			str      string
			value    []string
			isInList bool
		}{
			{
				name:     "in list",
				str:      "foo",
				value:    []string{"foo", "bar"},
				isInList: true,
			},
			{
				name:     "not in list",
				str:      "foo",
				value:    []string{"FOO", "BAR"},
				isInList: false,
			},
			{
				name:     "not in list",
				str:      "foo",
				value:    []string{"FOO", "bar"},
				isInList: false,
			},
			{
				name:     "in list",
				str:      "foo",
				value:    []string{"foo", "BAR"},
				isInList: true,
			},
		}

		for _, tc := range cases {
			reporter := newMockReporter(t)

			if tc.isInList {
				NewString(reporter, tc.str).
					InList(tc.value...).
					chain.assertNotFailed(t)
				NewString(reporter, tc.str).
					NotInList(tc.value...).
					chain.assertFailed(t)
			} else {
				NewString(reporter, tc.str).
					NotInList(tc.value...).
					chain.assertNotFailed(t)
				NewString(reporter, tc.str).
					InList(tc.value...).
					chain.assertFailed(t)
			}
		}
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewString(reporter, "foo").
			InList(). // empty list
			chain.assertFailed(t)

		NewString(reporter, "foo").
			NotInList(). // empty list
			chain.assertFailed(t)
	})
}

func TestString_InListFold(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		cases := []struct {
			name         string
			str          string
			value        []string
			isInListFold bool
		}{
			{
				name:         "in list",
				str:          "foo",
				value:        []string{"foo", "bar"},
				isInListFold: true,
			},
			{
				name:         "in list with case-insensitive match",
				str:          "foo",
				value:        []string{"FOO", "BAR"},
				isInListFold: true,
			},
			{
				name:         "not in list",
				str:          "foo",
				value:        []string{"BAR", "BAZ"},
				isInListFold: false,
			},
		}

		for _, tc := range cases {
			reporter := newMockReporter(t)

			if tc.isInListFold {
				NewString(reporter, tc.str).
					InListFold(tc.value...).
					chain.assertNotFailed(t)
				NewString(reporter, tc.str).
					NotInListFold(tc.value...).
					chain.assertFailed(t)
			} else {
				NewString(reporter, tc.str).
					NotInListFold(tc.value...).
					chain.assertNotFailed(t)
				NewString(reporter, tc.str).
					InListFold(tc.value...).
					chain.assertFailed(t)
			}
		}
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewString(reporter, "foo").
			InListFold(). // empty list
			chain.assertFailed(t)

		NewString(reporter, "foo").
			NotInListFold(). // empty list
			chain.assertFailed(t)
	})
}

func TestString_Contains(t *testing.T) {
	cases := []struct {
		name       string
		str        string
		value      string
		isContains bool
	}{
		{
			name:       "contains",
			str:        "11-foo-22",
			value:      "foo",
			isContains: true,
		},
		{
			name:       "not contains",
			str:        "11-foo-22",
			value:      "FOO",
			isContains: false,
		},
	}

	for _, tc := range cases {
		reporter := newMockReporter(t)

		if tc.isContains {
			NewString(reporter, tc.str).
				Contains(tc.value).
				chain.assertNotFailed(t)
			NewString(reporter, tc.str).
				NotContains(tc.value).
				chain.assertFailed(t)
		} else {
			NewString(reporter, tc.str).
				NotContains(tc.value).
				chain.assertNotFailed(t)
			NewString(reporter, tc.str).
				Contains(tc.value).
				chain.assertFailed(t)
		}
	}
}

func TestString_ContainsFold(t *testing.T) {
	cases := []struct {
		name           string
		str            string
		value          string
		isContainsFold bool
	}{
		{
			name:           "contains",
			str:            "11-foo-22",
			value:          "foo",
			isContainsFold: true,
		},
		{
			name:           "contains with case-insensitive match",
			str:            "11-foo-22",
			value:          "FOO",
			isContainsFold: true,
		},
		{
			name:           "not contains",
			str:            "11-foo-22",
			value:          "foo3",
			isContainsFold: false,
		},
	}

	for _, tc := range cases {
		reporter := newMockReporter(t)

		if tc.isContainsFold {
			NewString(reporter, tc.str).
				ContainsFold(tc.value).
				chain.assertNotFailed(t)
			NewString(reporter, tc.str).
				NotContainsFold(tc.value).
				chain.assertFailed(t)
		} else {
			NewString(reporter, tc.str).
				NotContainsFold(tc.value).
				chain.assertNotFailed(t)
			NewString(reporter, tc.str).
				ContainsFold(tc.value).
				chain.assertFailed(t)
		}
	}
}

func TestString_HasPrefix(t *testing.T) {
	cases := []struct {
		name        string
		str         string
		value       string
		isHasPrefix bool
	}{
		{
			name:        "has prefix",
			str:         "Hello World",
			value:       "Hello",
			isHasPrefix: true,
		},
		{
			name:        "full sentence",
			str:         "Hello World",
			value:       "Hello World",
			isHasPrefix: true,
		},
		{
			name:        "empty string",
			str:         "Hello World",
			value:       "",
			isHasPrefix: true,
		},
		{
			name:        "not has prefix",
			str:         "Hello World",
			value:       "Hello!",
			isHasPrefix: false,
		},
		{
			name:        "not has prefix",
			str:         "Hello World",
			value:       "hello",
			isHasPrefix: false,
		},
		{
			name:        "not has prefix",
			str:         "Hello World",
			value:       "World",
			isHasPrefix: false,
		},
		{
			name:        "not has prefix",
			str:         "Hello World",
			value:       "Bye",
			isHasPrefix: false,
		},
	}

	for _, tc := range cases {
		reporter := newMockReporter(t)

		if tc.isHasPrefix {
			NewString(reporter, tc.str).
				HasPrefix(tc.value).
				chain.assertNotFailed(t)
			NewString(reporter, tc.str).
				NotHasPrefix(tc.value).
				chain.assertFailed(t)
		} else {
			NewString(reporter, tc.str).
				NotHasPrefix(tc.value).
				chain.assertNotFailed(t)
			NewString(reporter, tc.str).
				HasPrefix(tc.value).
				chain.assertFailed(t)
		}
	}
}

func TestString_HasSuffix(t *testing.T) {
	cases := []struct {
		name        string
		str         string
		value       string
		isHasSuffix bool
	}{
		{
			name:        "has suffix",
			str:         "Hello World",
			value:       "World",
			isHasSuffix: true,
		},
		{
			name:        "full sentence",
			str:         "Hello World",
			value:       "Hello World",
			isHasSuffix: true,
		},
		{
			name:        "empty string",
			str:         "Hello World",
			value:       "",
			isHasSuffix: true,
		},
		{
			name:        "not has suffix",
			str:         "Hello World",
			value:       "World!",
			isHasSuffix: false,
		},
		{
			name:        "not has suffix",
			str:         "Hello World",
			value:       "world",
			isHasSuffix: false,
		},
		{
			name:        "not has suffix",
			str:         "Hello World",
			value:       "Hello",
			isHasSuffix: false,
		},
		{
			name:        "not has suffix",
			str:         "Hello World",
			value:       "Bye",
			isHasSuffix: false,
		},
	}

	for _, tc := range cases {
		reporter := newMockReporter(t)

		if tc.isHasSuffix {
			NewString(reporter, tc.str).
				HasSuffix(tc.value).
				chain.assertNotFailed(t)
			NewString(reporter, tc.str).
				NotHasSuffix(tc.value).
				chain.assertFailed(t)
		} else {
			NewString(reporter, tc.str).
				NotHasSuffix(tc.value).
				chain.assertNotFailed(t)
			NewString(reporter, tc.str).
				HasSuffix(tc.value).
				chain.assertFailed(t)
		}
	}
}

func TestString_HasPrefixFold(t *testing.T) {
	cases := []struct {
		name            string
		str             string
		value           string
		isHasPrefixFold bool
	}{
		{
			name:            "has prefix with case-insensitive match",
			str:             "Hello World",
			value:           "hello",
			isHasPrefixFold: true,
		},
		{
			name:            "full sentence with case-insensitive match",
			str:             "Hello World",
			value:           "HeLlO wOrLd",
			isHasPrefixFold: true,
		},
		{
			name:            "empty string",
			str:             "Hello World",
			value:           "",
			isHasPrefixFold: true,
		},
		{
			name:            "not has prefix",
			str:             "Hello World",
			value:           "World",
			isHasPrefixFold: false,
		},
		{
			name:            "not has prefix",
			str:             "Hello World",
			value:           "Bye",
			isHasPrefixFold: false,
		},
		{
			name:            "not has prefix",
			str:             "Hello World",
			value:           "world",
			isHasPrefixFold: false,
		},
		{
			name:            "not has prefix",
			str:             "Hello World",
			value:           "world!",
			isHasPrefixFold: false,
		},
	}

	for _, tc := range cases {
		reporter := newMockReporter(t)

		if tc.isHasPrefixFold {
			NewString(reporter, tc.str).
				HasPrefixFold(tc.value).
				chain.assertNotFailed(t)
			NewString(reporter, tc.str).
				NotHasPrefixFold(tc.value).
				chain.assertFailed(t)
		} else {
			NewString(reporter, tc.str).
				NotHasPrefixFold(tc.value).
				chain.assertNotFailed(t)
			NewString(reporter, tc.str).
				HasPrefixFold(tc.value).
				chain.assertFailed(t)
		}
	}
}

func TestString_HasSuffixFold(t *testing.T) {
	cases := []struct {
		name            string
		str             string
		value           string
		isHasSuffixFold bool
	}{
		{
			name:            "has suffix",
			str:             "Hello World",
			value:           "world",
			isHasSuffixFold: true,
		},
		{
			name:            "full sentence with case-insensitive match",
			str:             "Hello World",
			value:           "hElLo WoRlD",
			isHasSuffixFold: true,
		},
		{
			name:            "empty string",
			str:             "Hello World",
			value:           "",
			isHasSuffixFold: true,
		},
		{
			name:            "not has suffix",
			str:             "Hello World",
			value:           "hello",
			isHasSuffixFold: false,
		},
		{
			name:            "not has suffix",
			str:             "Hello World",
			value:           "world!",
			isHasSuffixFold: false,
		},
		{
			name:            "not has suffix",
			str:             "Hello World",
			value:           "Bye",
			isHasSuffixFold: false,
		},
	}

	for _, tc := range cases {
		reporter := newMockReporter(t)

		if tc.isHasSuffixFold {
			NewString(reporter, tc.str).
				HasSuffixFold(tc.value).
				chain.assertNotFailed(t)
			NewString(reporter, tc.str).
				NotHasSuffixFold(tc.value).
				chain.assertFailed(t)
		} else {
			NewString(reporter, tc.str).
				NotHasSuffixFold(tc.value).
				chain.assertNotFailed(t)
			NewString(reporter, tc.str).
				HasSuffixFold(tc.value).
				chain.assertFailed(t)
		}
	}
}

func TestString_MatchOne(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "http://example.com/users/john")

	t.Run("named", func(t *testing.T) {
		m := value.Match(`http://(?P<host>.+)/users/(?P<user>.+)`)
		m.chain.assertNotFailed(t)

		assert.Equal(t,
			[]string{"http://example.com/users/john", "example.com", "john"},
			m.submatches)
	})

	t.Run("unnamed", func(t *testing.T) {
		m := value.Match(`http://(.+)/users/(.+)`)
		m.chain.assertNotFailed(t)

		assert.Equal(t,
			[]string{"http://example.com/users/john", "example.com", "john"},
			m.submatches)
	})
}

func TestString_MatchAll(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter,
		"http://example.com/users/john http://example.com/users/bob")

	m := value.MatchAll(`http://(\S+)/users/(\S+)`)

	assert.Equal(t, 2, len(m))

	m[0].chain.assertNotFailed(t)
	m[1].chain.assertNotFailed(t)

	assert.Equal(t,
		[]string{"http://example.com/users/john", "example.com", "john"},
		m[0].submatches)

	assert.Equal(t,
		[]string{"http://example.com/users/bob", "example.com", "bob"},
		m[1].submatches)
}

func TestString_MatchStatus(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "a")

	value.Match(`a`)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.MatchAll(`a`)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotMatch(`a`)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Match(`[^a]`)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.MatchAll(`[^a]`)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotMatch(`[^a]`)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	assert.Equal(t, []string{}, value.Match(`[^a]`).submatches)
	assert.Equal(t, []Match{}, value.MatchAll(`[^a]`))
}

func TestString_MatchInvalid(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "a")

	value.Match(`[`)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.MatchAll(`[`)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotMatch(`[`)
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestString_IsAscii(t *testing.T) {
	cases := []struct {
		str       string
		wantASCII bool
	}{
		{"Ascii", true},
		{string(rune(127)), true},
		{"Ascii is アスキー", false},
		{"アスキー", false},
		{string(rune(128)), false},
	}

	for _, value := range cases {
		t.Run(value.str, func(t *testing.T) {
			reporter := newMockReporter(t)

			if value.wantASCII {
				NewString(reporter, value.str).IsASCII().
					chain.assertNotFailed(t)

				NewString(reporter, value.str).NotASCII().
					chain.assertFailed(t)
			} else {
				NewString(reporter, value.str).IsASCII().
					chain.assertFailed(t)

				NewString(reporter, value.str).NotASCII().
					chain.assertNotFailed(t)
			}
		})
	}
}

func TestString_AsNumber(t *testing.T) {
	cases := []struct {
		name        string
		str         string
		base        []int
		fail        bool
		expectedNum float64
	}{
		{
			name:        "default_base_integer",
			str:         "1234567",
			fail:        false,
			expectedNum: float64(1234567),
		},
		{
			name:        "default_base_float",
			str:         "11.22",
			fail:        false,
			expectedNum: float64(11.22),
		},
		{
			name:        "default_base_bad",
			str:         "a1",
			fail:        true,
			expectedNum: 0,
		},
		{
			name:        "base10_integer",
			str:         "100",
			base:        []int{10},
			fail:        false,
			expectedNum: float64(100),
		},
		{
			name:        "base10_float",
			str:         "11.22",
			base:        []int{10},
			fail:        false,
			expectedNum: float64(11.22),
		},
		{
			name:        "base16_integer",
			str:         "100",
			base:        []int{16},
			fail:        false,
			expectedNum: float64(0x100),
		},
		{
			name:        "base16_float",
			str:         "11.22",
			base:        []int{16},
			fail:        true,
			expectedNum: 0,
		},
		{
			name:        "base16_large_integer",
			str:         "4000000000000000",
			base:        []int{16},
			fail:        false,
			expectedNum: float64(0x4000000000000000),
		},
		{
			name: "default_float_precision_max",
			str:  "4611686018427387905",
			fail: true,
		},
		{
			name: "base10_float_precision_max",
			str:  "4611686018427387905",
			base: []int{10},
			fail: true,
		},
		{
			name: "base16_float_precision_max",
			str:  "8000000000000001",
			base: []int{16},
			fail: true,
		},
		{
			name: "base16_float_precision_min",
			str:  "-4000000000000001",
			base: []int{16},
			fail: true,
		},
		{
			name: "multiple_base",
			str:  "100",
			base: []int{10, 16},
			fail: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			str := NewString(reporter, tc.str)
			num := str.AsNumber(tc.base...)

			if tc.fail {
				str.chain.assertFailed(t)
				num.chain.assertFailed(t)
				assert.Equal(t, float64(0), num.Raw())
			} else {
				str.chain.assertNotFailed(t)
				num.chain.assertNotFailed(t)
				assert.Equal(t, tc.expectedNum, num.Raw())
			}
		})
	}
}

func TestString_AsBoolean(t *testing.T) {
	trueValues := []string{"true", "True"}
	falseValues := []string{"false", "False"}
	badValues := []string{"TRUE", "FALSE", "t", "f", "1", "0", "bad"}

	for _, str := range trueValues {
		t.Run(str, func(t *testing.T) {
			reporter := newMockReporter(t)
			value := NewString(reporter, str)

			b := value.AsBoolean()
			b.chain.assertNotFailed(t)

			assert.True(t, b.Raw())
		})
	}

	for _, str := range falseValues {
		t.Run(str, func(t *testing.T) {
			reporter := newMockReporter(t)
			value := NewString(reporter, str)

			b := value.AsBoolean()
			b.chain.assertNotFailed(t)

			assert.False(t, b.Raw())
		})
	}

	for _, str := range badValues {
		t.Run(str, func(t *testing.T) {
			reporter := newMockReporter(t)
			value := NewString(reporter, str)

			b := value.AsBoolean()
			b.chain.assertFailed(t)
		})
	}
}

func TestString_AsDateTime(t *testing.T) {
	t.Run("default_formats_RFC1123+GMT", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewString(reporter, "Tue, 15 Nov 1994 08:12:31 GMT")

		dt := value.AsDateTime()
		value.chain.assertNotFailed(t)
		dt.chain.assertNotFailed(t)

		assert.True(t, time.Date(1994, 11, 15, 8, 12, 31, 0, time.UTC).Equal(dt.Raw()))
	})

	t.Run("RFC822", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewString(reporter, "15 Nov 94 08:12 GMT")

		dt := value.AsDateTime(time.RFC822)
		value.chain.assertNotFailed(t)
		dt.chain.assertNotFailed(t)

		assert.True(t, time.Date(1994, 11, 15, 8, 12, 0, 0, time.UTC).Equal(dt.Raw()))
	})

	t.Run("bad_input", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewString(reporter, "bad")

		dt := value.AsDateTime()
		value.chain.assertFailed(t)
		dt.chain.assertFailed(t)

		assert.True(t, time.Unix(0, 0).Equal(dt.Raw()))
	})

	formats := []string{
		http.TimeFormat,
		time.RFC850,
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC822,
		time.RFC822Z,
		time.RFC3339,
		time.RFC3339Nano,
	}

	for n, f := range formats {
		t.Run("default_formats_"+f, func(t *testing.T) {
			str := time.Now().Format(f)

			reporter := newMockReporter(t)
			value := NewString(reporter, str)

			dt := value.AsDateTime()
			dt.chain.assertNotFailed(t)
		})

		t.Run("all_formats_"+f, func(t *testing.T) {
			str := time.Now().Format(f)

			reporter := newMockReporter(t)
			value := NewString(reporter, str)

			dt := value.AsDateTime(formats...)
			dt.chain.assertNotFailed(t)
		})

		t.Run("same_format_"+f, func(t *testing.T) {
			str := time.Now().Format(f)

			reporter := newMockReporter(t)
			value := NewString(reporter, str)

			dt := value.AsDateTime(f)
			dt.chain.assertNotFailed(t)
		})

		if n != 0 {
			t.Run("different_format_"+f, func(t *testing.T) {
				str := time.Now().Format(f)

				reporter := newMockReporter(t)
				value := NewString(reporter, str)

				dt := value.AsDateTime(formats[0])
				dt.chain.assertFailed(t)
			})
		}
	}
}
