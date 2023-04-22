package httpexpect

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestString_FailedChain(t *testing.T) {
	chain := newMockChain(t, flagFailed)

	value := newString(chain, "")
	value.chain.assert(t, failure)

	value.Path("$").chain.assert(t, failure)
	value.Schema("")
	value.Alias("foo")

	var target interface{}
	value.Decode(target)

	value.Length().chain.assert(t, failure)

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

	value.Match("").chain.assert(t, failure)
	value.NotMatch("")
	assert.NotNil(t, value.MatchAll(""))
	assert.Equal(t, 0, len(value.MatchAll("")))

	value.AsBoolean().chain.assert(t, failure)
	value.AsNumber().chain.assert(t, failure)
	value.AsDateTime().chain.assert(t, failure)
}

func TestString_Constructors(t *testing.T) {
	t.Run("reporter", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewString(reporter, "Hello")
		value.IsEqual("Hello")
		value.chain.assert(t, success)
	})

	t.Run("config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewStringC(Config{
			Reporter: reporter,
		}, "Hello")
		value.IsEqual("Hello")
		value.chain.assert(t, success)
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

		value.chain.assert(t, success)
		assert.Equal(t, "foo", target)
	})

	t.Run("target is string", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewString(reporter, "foo")

		var target string
		value.Decode(&target)

		value.chain.assert(t, success)
		assert.Equal(t, "foo", target)
	})

	t.Run("target is unmarshable", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewString(reporter, "foo")

		value.Decode(123)

		value.chain.assert(t, failure)
	})

	t.Run("target is nil", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewString(reporter, "foo")

		value.Decode(nil)

		value.chain.assert(t, failure)
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

func TestString_Path(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "foo")

	assert.Equal(t, "foo", value.Path("$").Raw())
	value.chain.assert(t, success)
}

func TestString_Schema(t *testing.T) {
	reporter := newMockReporter(t)

	NewString(reporter, "foo").Schema(`{"type": "string"}`).
		chain.assert(t, success)

	NewString(reporter, "foo").Schema(`{"type": "object"}`).
		chain.assert(t, failure)
}

func TestString_Getters(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "foo")

	assert.Equal(t, "foo", value.Raw())
	value.chain.assert(t, success)
	value.chain.clear()

	num := value.Length()
	value.chain.assert(t, success)
	num.chain.assert(t, success)
	assert.Equal(t, 3.0, num.Raw())
}

func TestString_IsEmpty(t *testing.T) {
	cases := []struct {
		name      string
		str       string
		wantEmpty chainResult
	}{
		{
			name:      "empty string",
			str:       "",
			wantEmpty: success,
		},
		{
			name:      "non-empty string",
			str:       "foo",
			wantEmpty: failure,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewString(reporter, tc.str).IsEmpty().
				chain.assert(t, tc.wantEmpty)

			NewString(reporter, tc.str).NotEmpty().
				chain.assert(t, !tc.wantEmpty)
		})
	}
}

func TestString_IsEqual(t *testing.T) {
	cases := []struct {
		name          string
		str           string
		value         string
		wantEqual     chainResult
		wantEqualFold chainResult
	}{
		{
			name:          "equivalent string",
			str:           "foo",
			value:         "foo",
			wantEqual:     success,
			wantEqualFold: success,
		},
		{
			name:          "non-equivalent string",
			str:           "foo",
			value:         "bar",
			wantEqual:     failure,
			wantEqualFold: failure,
		},
		{
			name:          "different case",
			str:           "foo",
			value:         "FOO",
			wantEqual:     failure,
			wantEqualFold: success,
		},
	}

	for _, tc := range cases {
		reporter := newMockReporter(t)

		NewString(reporter, tc.str).IsEqual(tc.value).
			chain.assert(t, tc.wantEqual)
		NewString(reporter, tc.str).NotEqual(tc.value).
			chain.assert(t, !tc.wantEqual)

		NewString(reporter, tc.str).IsEqualFold(tc.value).
			chain.assert(t, tc.wantEqualFold)
		NewString(reporter, tc.str).NotEqualFold(tc.value).
			chain.assert(t, !tc.wantEqualFold)
	}
}

func TestString_InList(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		cases := []struct {
			name           string
			str            string
			value          []string
			wantInList     chainResult
			wantInListFold chainResult
		}{
			{
				name:           "in list",
				str:            "foo",
				value:          []string{"foo", "bar"},
				wantInList:     success,
				wantInListFold: success,
			},
			{
				name:           "not in list",
				str:            "baz",
				value:          []string{"foo", "bar"},
				wantInList:     failure,
				wantInListFold: failure,
			},
			{
				name:           "different case",
				str:            "FOO",
				value:          []string{"foo", "bar"},
				wantInList:     failure,
				wantInListFold: success,
			},
		}

		for _, tc := range cases {
			reporter := newMockReporter(t)

			NewString(reporter, tc.str).InList(tc.value...).
				chain.assert(t, tc.wantInList)
			NewString(reporter, tc.str).NotInList(tc.value...).
				chain.assert(t, !tc.wantInList)

			NewString(reporter, tc.str).InListFold(tc.value...).
				chain.assert(t, tc.wantInListFold)
			NewString(reporter, tc.str).NotInListFold(tc.value...).
				chain.assert(t, !tc.wantInListFold)
		}
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewString(reporter, "foo").InList().
			chain.assert(t, failure)
		NewString(reporter, "foo").NotInList().
			chain.assert(t, failure)

		NewString(reporter, "foo").InListFold().
			chain.assert(t, failure)
		NewString(reporter, "foo").NotInListFold().
			chain.assert(t, failure)
	})
}

func TestString_Contains(t *testing.T) {
	cases := []struct {
		name             string
		str              string
		value            string
		wantContains     chainResult
		wantContainsFold chainResult
	}{
		{
			name:             "contains",
			str:              "11-foo-22",
			value:            "foo",
			wantContains:     success,
			wantContainsFold: success,
		},
		{
			name:             "not contains",
			str:              "11-foo-22",
			value:            "bar",
			wantContains:     failure,
			wantContainsFold: failure,
		},
		{
			name:             "different case",
			str:              "11-foo-22",
			value:            "FOO",
			wantContains:     failure,
			wantContainsFold: success,
		},
	}

	for _, tc := range cases {
		reporter := newMockReporter(t)

		NewString(reporter, tc.str).Contains(tc.value).
			chain.assert(t, tc.wantContains)
		NewString(reporter, tc.str).NotContains(tc.value).
			chain.assert(t, !tc.wantContains)

		NewString(reporter, tc.str).ContainsFold(tc.value).
			chain.assert(t, tc.wantContainsFold)
		NewString(reporter, tc.str).NotContainsFold(tc.value).
			chain.assert(t, !tc.wantContainsFold)
	}
}

func TestString_HasPrefix(t *testing.T) {
	cases := []struct {
		name              string
		str               string
		value             string
		wantHasPrefix     chainResult
		wantHasPrefixFold chainResult
	}{
		{
			name:              "has prefix",
			str:               "Hello World",
			value:             "Hello",
			wantHasPrefix:     success,
			wantHasPrefixFold: success,
		},
		{
			name:              "full match",
			str:               "Hello World",
			value:             "Hello World",
			wantHasPrefix:     success,
			wantHasPrefixFold: success,
		},
		{
			name:              "empty string",
			str:               "Hello World",
			value:             "",
			wantHasPrefix:     success,
			wantHasPrefixFold: success,
		},
		{
			name:              "extra char",
			str:               "Hello World",
			value:             "Hello!",
			wantHasPrefix:     failure,
			wantHasPrefixFold: failure,
		},
		{
			name:              "different case",
			str:               "Hello World",
			value:             "hell",
			wantHasPrefix:     failure,
			wantHasPrefixFold: success,
		},
		{
			name:              "different case extra char",
			str:               "Hello World",
			value:             "hella",
			wantHasPrefix:     failure,
			wantHasPrefixFold: failure,
		},
		{
			name:              "different case full match",
			str:               "Hello World",
			value:             "hELLO wORLD",
			wantHasPrefix:     failure,
			wantHasPrefixFold: success,
		},
	}

	for _, tc := range cases {
		reporter := newMockReporter(t)

		NewString(reporter, tc.str).HasPrefix(tc.value).
			chain.assert(t, tc.wantHasPrefix)
		NewString(reporter, tc.str).NotHasPrefix(tc.value).
			chain.assert(t, !tc.wantHasPrefix)

		NewString(reporter, tc.str).HasPrefixFold(tc.value).
			chain.assert(t, tc.wantHasPrefixFold)
		NewString(reporter, tc.str).NotHasPrefixFold(tc.value).
			chain.assert(t, !tc.wantHasPrefixFold)
	}
}

func TestString_HasSuffix(t *testing.T) {
	cases := []struct {
		name              string
		str               string
		value             string
		wantHasSuffix     chainResult
		wantHasSuffixFold chainResult
	}{
		{
			name:              "has suffix",
			str:               "Hello World",
			value:             "World",
			wantHasSuffix:     success,
			wantHasSuffixFold: success,
		},
		{
			name:              "full match",
			str:               "Hello World",
			value:             "Hello World",
			wantHasSuffix:     success,
			wantHasSuffixFold: success,
		},
		{
			name:              "empty string",
			str:               "Hello World",
			value:             "",
			wantHasSuffix:     success,
			wantHasSuffixFold: success,
		},
		{
			name:              "extra char",
			str:               "Hello World",
			value:             "!World",
			wantHasSuffix:     failure,
			wantHasSuffixFold: failure,
		},
		{
			name:              "different case",
			str:               "Hello World",
			value:             "WORLD",
			wantHasSuffix:     failure,
			wantHasSuffixFold: success,
		},
		{
			name:              "different case extra char",
			str:               "Hello World",
			value:             "!WORLD",
			wantHasSuffix:     failure,
			wantHasSuffixFold: failure,
		},
		{
			name:              "different case full match",
			str:               "Hello World",
			value:             "hELLO wORLD",
			wantHasSuffix:     failure,
			wantHasSuffixFold: success,
		},
	}

	for _, tc := range cases {
		reporter := newMockReporter(t)

		NewString(reporter, tc.str).HasSuffix(tc.value).
			chain.assert(t, tc.wantHasSuffix)
		NewString(reporter, tc.str).NotHasSuffix(tc.value).
			chain.assert(t, !tc.wantHasSuffix)

		NewString(reporter, tc.str).HasSuffixFold(tc.value).
			chain.assert(t, tc.wantHasSuffixFold)
		NewString(reporter, tc.str).NotHasSuffixFold(tc.value).
			chain.assert(t, !tc.wantHasSuffixFold)
	}
}

func TestString_Match(t *testing.T) {
	t.Run("named", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewString(reporter, "http://example.com/users/john")

		m := value.Match(`http://(?P<host>.+)/users/(?P<user>.+)`)
		m.chain.assert(t, success)

		assert.Equal(t,
			[]string{"http://example.com/users/john", "example.com", "john"},
			m.submatches)
	})

	t.Run("unnamed", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewString(reporter, "http://example.com/users/john")

		m := value.Match(`http://(.+)/users/(.+)`)
		m.chain.assert(t, success)

		assert.Equal(t,
			[]string{"http://example.com/users/john", "example.com", "john"},
			m.submatches)
	})

	t.Run("all", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewString(reporter,
			"http://example.com/users/john http://example.com/users/bob")

		m := value.MatchAll(`http://(\S+)/users/(\S+)`)

		assert.Equal(t, 2, len(m))

		m[0].chain.assert(t, success)
		m[1].chain.assert(t, success)

		assert.Equal(t,
			[]string{"http://example.com/users/john", "example.com", "john"},
			m[0].submatches)

		assert.Equal(t,
			[]string{"http://example.com/users/bob", "example.com", "bob"},
			m[1].submatches)
	})

	t.Run("status", func(t *testing.T) {
		cases := []struct {
			str          string
			re           string
			wantMatch    chainResult
			wantNotMatch chainResult
		}{
			{
				str:          `a`,
				re:           `a`,
				wantMatch:    success,
				wantNotMatch: failure,
			},
			{
				str:          `a`,
				re:           `[^a]`,
				wantMatch:    failure,
				wantNotMatch: success,
			},
			{
				str:          `a`,
				re:           `[`,
				wantMatch:    failure,
				wantNotMatch: failure,
			},
		}

		for _, tc := range cases {
			t.Run(tc.str, func(t *testing.T) {
				reporter := newMockReporter(t)

				var value *String

				value = NewString(reporter, tc.str)
				value.Match(tc.re).chain.assert(t, tc.wantMatch)
				value.chain.assert(t, tc.wantMatch)

				value = NewString(reporter, tc.str)
				value.MatchAll(tc.re)
				value.chain.assert(t, tc.wantMatch)

				value = NewString(reporter, tc.str)
				value.NotMatch(tc.re)
				value.chain.assert(t, tc.wantNotMatch)
			})
		}
	})
}

func TestString_IsAscii(t *testing.T) {
	cases := []struct {
		str         string
		wantIsASCII chainResult
	}{
		{"Ascii", success},
		{string(rune(127)), success},
		{"Ascii is アスキー", failure},
		{"アスキー", failure},
		{string(rune(128)), failure},
	}

	for _, tc := range cases {
		t.Run(tc.str, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewString(reporter, tc.str).IsASCII().
				chain.assert(t, tc.wantIsASCII)

			NewString(reporter, tc.str).NotASCII().
				chain.assert(t, !tc.wantIsASCII)
		})
	}
}

func TestString_AsNumber(t *testing.T) {
	cases := []struct {
		name        string
		str         string
		base        []int
		result      chainResult
		expectedNum float64
	}{
		{
			name:        "default base integer",
			str:         "1234567",
			result:      success,
			expectedNum: float64(1234567),
		},
		{
			name:        "default base float",
			str:         "11.22",
			result:      success,
			expectedNum: float64(11.22),
		},
		{
			name:   "default base bad",
			str:    "a1",
			result: failure,
		},
		{
			name:        "base10 integer",
			str:         "100",
			base:        []int{10},
			result:      success,
			expectedNum: float64(100),
		},
		{
			name:        "base10 float",
			str:         "11.22",
			base:        []int{10},
			result:      success,
			expectedNum: float64(11.22),
		},
		{
			name:        "base16 integer",
			str:         "100",
			base:        []int{16},
			result:      success,
			expectedNum: float64(0x100),
		},
		{
			name:   "base16 float",
			str:    "11.22",
			base:   []int{16},
			result: failure,
		},
		{
			name:        "base16 large integer",
			str:         "4000000000000000",
			base:        []int{16},
			result:      success,
			expectedNum: float64(0x4000000000000000),
		},
		{
			name:   "default float precision max",
			str:    "4611686018427387905",
			result: failure,
		},
		{
			name:   "base10 float precision max",
			str:    "4611686018427387905",
			base:   []int{10},
			result: failure,
		},
		{
			name:   "base16 float precision max",
			str:    "8000000000000001",
			base:   []int{16},
			result: failure,
		},
		{
			name:   "base16 float precision min",
			str:    "-4000000000000001",
			base:   []int{16},
			result: failure,
		},
		{
			name:   "multiple bases",
			str:    "100",
			base:   []int{10, 16},
			result: failure,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			str := NewString(reporter, tc.str)
			num := str.AsNumber(tc.base...)

			str.chain.assert(t, tc.result)
			num.chain.assert(t, tc.result)

			if tc.result {
				assert.Equal(t, tc.expectedNum, num.Raw())
			} else {
				assert.Equal(t, float64(0), num.Raw())
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
			b.chain.assert(t, success)

			assert.True(t, b.Raw())
		})
	}

	for _, str := range falseValues {
		t.Run(str, func(t *testing.T) {
			reporter := newMockReporter(t)
			value := NewString(reporter, str)

			b := value.AsBoolean()
			b.chain.assert(t, success)

			assert.False(t, b.Raw())
		})
	}

	for _, str := range badValues {
		t.Run(str, func(t *testing.T) {
			reporter := newMockReporter(t)
			value := NewString(reporter, str)

			b := value.AsBoolean()
			b.chain.assert(t, failure)
		})
	}
}

func TestString_AsDateTime(t *testing.T) {
	t.Run("default formats - RFC1123+GMT", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewString(reporter, "Tue, 15 Nov 1994 08:12:31 GMT")

		dt := value.AsDateTime()
		value.chain.assert(t, success)
		dt.chain.assert(t, success)

		assert.True(t, time.Date(1994, 11, 15, 8, 12, 31, 0, time.UTC).Equal(dt.Raw()))
	})

	t.Run("RFC822", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewString(reporter, "15 Nov 94 08:12 GMT")

		dt := value.AsDateTime(time.RFC822)
		value.chain.assert(t, success)
		dt.chain.assert(t, success)

		assert.True(t, time.Date(1994, 11, 15, 8, 12, 0, 0, time.UTC).Equal(dt.Raw()))
	})

	t.Run("bad input", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewString(reporter, "bad")

		dt := value.AsDateTime()
		value.chain.assert(t, failure)
		dt.chain.assert(t, failure)

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
		t.Run("default formats - "+f, func(t *testing.T) {
			str := time.Now().Format(f)

			reporter := newMockReporter(t)
			value := NewString(reporter, str)

			dt := value.AsDateTime()
			dt.chain.assert(t, success)
		})

		t.Run("all formats - "+f, func(t *testing.T) {
			str := time.Now().Format(f)

			reporter := newMockReporter(t)
			value := NewString(reporter, str)

			dt := value.AsDateTime(formats...)
			dt.chain.assert(t, success)
		})

		t.Run("same format - "+f, func(t *testing.T) {
			str := time.Now().Format(f)

			reporter := newMockReporter(t)
			value := NewString(reporter, str)

			dt := value.AsDateTime(f)
			dt.chain.assert(t, success)
		})

		if n != 0 {
			t.Run("different format - "+f, func(t *testing.T) {
				str := time.Now().Format(f)

				reporter := newMockReporter(t)
				value := NewString(reporter, str)

				dt := value.AsDateTime(formats[0])
				dt.chain.assert(t, failure)
			})
		}
	}
}
