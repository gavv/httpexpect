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
	t.Run("Decode into empty interface", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewString(reporter, "foo")

		var target interface{}
		value.Decode(&target)

		value.chain.assertNotFailed(t)
		assert.Equal(t, "foo", target)
	})

	t.Run("Decode into string", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewString(reporter, "foo")

		var target string
		value.Decode(&target)

		value.chain.assertNotFailed(t)
		assert.Equal(t, "foo", target)
	})

	t.Run("Target is unmarshable", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewString(reporter, "foo")

		value.Decode(123)

		value.chain.assertFailed(t)
	})

	t.Run("Target is nil", func(t *testing.T) {
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
}

func TestString_Length(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "1234567")

	num := value.Length()
	value.chain.assertNotFailed(t)
	num.chain.assertNotFailed(t)
	assert.Equal(t, 7.0, num.Raw())
}

func TestString_Empty(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "")

	value.IsEmpty()
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotEmpty()
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value = NewString(reporter, "a")

	value.IsEmpty()
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEmpty()
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}

func TestString_Equal(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "foo")

	assert.Equal(t, "foo", value.Raw())

	value.IsEqual("foo")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.IsEqual("FOO")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual("FOO")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotEqual("foo")
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestString_EqualFold(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "foo")

	value.IsEqualFold("foo")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.IsEqualFold("FOO")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.IsEqualFold("foo2")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqualFold("foo")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqualFold("FOO")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqualFold("foo2")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}

func TestString_InList(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "foo")

	assert.Equal(t, "foo", value.Raw())

	value.InList()
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInList()
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InList("foo", "bar")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.InList("FOO", "BAR")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInList("FOO", "bar")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotInList("foo", "BAR")
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestString_InListFold(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "Foo")

	assert.Equal(t, "Foo", value.Raw())

	value.InListFold()
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInListFold()
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InListFold("foo", "bar")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.InListFold("FOO", "BAR")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.InListFold("BAR", "BAZ")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInListFold("foo", "bar")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInListFold("FOO", "BAR")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInListFold("BAR", "BAZ")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}

func TestString_Contains(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "11-foo-22")

	value.Contains("foo")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Contains("FOO")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotContains("FOO")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotContains("foo")
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestString_ContainsFold(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "11-foo-22")

	value.ContainsFold("foo")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.ContainsFold("FOO")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.ContainsFold("foo3")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotContainsFold("foo")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotContainsFold("FOO")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotContainsFold("foo3")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}

func TestString_MatchOne(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "http://example.com/users/john")

	m1 := value.Match(`http://(?P<host>.+)/users/(?P<user>.+)`)
	m1.chain.assertNotFailed(t)

	assert.Equal(t,
		[]string{"http://example.com/users/john", "example.com", "john"},
		m1.submatches)

	m2 := value.Match(`http://(.+)/users/(.+)`)
	m2.chain.assertNotFailed(t)

	assert.Equal(t,
		[]string{"http://example.com/users/john", "example.com", "john"},
		m2.submatches)
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
	reporter := newMockReporter(t)

	value := NewString(reporter, "Ascii")
	value.IsASCII()
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value = NewString(reporter, "Ascii is アスキー")
	value.IsASCII()
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value = NewString(reporter, "アスキー")
	value.IsASCII()
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value = NewString(reporter, string(rune(127)))
	value.IsASCII()
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value = NewString(reporter, string(rune(128)))
	value.IsASCII()
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestString_NotAscii(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "Ascii")
	value.NotASCII()
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value = NewString(reporter, "Ascii is アスキー")
	value.NotASCII()
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value = NewString(reporter, "アスキー")
	value.NotASCII()
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value = NewString(reporter, string(rune(127)))
	value.NotASCII()
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value = NewString(reporter, string(rune(128)))
	value.NotASCII()
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}

func TestString_AsNumber(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("default_base_integer", func(t *testing.T) {
		value := NewString(reporter, "1234567")
		num := value.AsNumber()
		value.chain.assertNotFailed(t)
		num.chain.assertNotFailed(t)
		assert.Equal(t, float64(1234567), num.Raw())
	})

	t.Run("default_base_float", func(t *testing.T) {
		value := NewString(reporter, "11.22")
		num := value.AsNumber()
		value.chain.assertNotFailed(t)
		num.chain.assertNotFailed(t)
		assert.Equal(t, float64(11.22), num.Raw())
	})

	t.Run("default_base_bad", func(t *testing.T) {
		value := NewString(reporter, "a1")
		num := value.AsNumber()
		value.chain.assertFailed(t)
		num.chain.assertFailed(t)
		assert.Equal(t, float64(0), num.Raw())
	})

	t.Run("base10_integer", func(t *testing.T) {
		value := NewString(reporter, "100")
		num := value.AsNumber(10)
		value.chain.assertNotFailed(t)
		num.chain.assertNotFailed(t)
		assert.Equal(t, float64(100), num.Raw())
	})

	t.Run("base10_float", func(t *testing.T) {
		value := NewString(reporter, "11.22")
		num := value.AsNumber(10)
		value.chain.assertNotFailed(t)
		num.chain.assertNotFailed(t)
		assert.Equal(t, float64(11.22), num.Raw())
	})

	t.Run("base16_integer", func(t *testing.T) {
		value := NewString(reporter, "100")
		num := value.AsNumber(16)
		value.chain.assertNotFailed(t)
		num.chain.assertNotFailed(t)
		assert.Equal(t, float64(0x100), num.Raw())
	})

	t.Run("base16_float", func(t *testing.T) {
		value := NewString(reporter, "11.22")
		num := value.AsNumber(16)
		value.chain.assertFailed(t)
		num.chain.assertFailed(t)
		assert.Equal(t, float64(0), num.Raw())
	})

	t.Run("base16_large_integer", func(t *testing.T) {
		value := NewString(reporter, "4000000000000000")
		num := value.AsNumber(16)
		value.chain.assertNotFailed(t)
		num.chain.assertNotFailed(t)
		assert.Equal(t, float64(0x4000000000000000), num.Raw())
	})

	t.Run("default_base_float_precision", func(t *testing.T) {
		value := NewString(reporter, "4611686018427387905")
		num := value.AsNumber()
		value.chain.assertFailed(t)
		num.chain.assertFailed(t)
		assert.Equal(t, float64(0), num.Raw())
	})

	t.Run("base10_float_precision", func(t *testing.T) {
		value := NewString(reporter, "4611686018427387905")
		num := value.AsNumber(10)
		value.chain.assertFailed(t)
		num.chain.assertFailed(t)
		assert.Equal(t, float64(0), num.Raw())
	})

	t.Run("base16_float_precision_max", func(t *testing.T) {
		value := NewString(reporter, "8000000000000001")
		num := value.AsNumber(16)
		value.chain.assertFailed(t)
		num.chain.assertFailed(t)
		assert.Equal(t, float64(0), num.Raw())
	})

	t.Run("base16_float_precision_min", func(t *testing.T) {
		value := NewString(reporter, "-4000000000000001")
		num := value.AsNumber(16)
		value.chain.assertFailed(t)
		num.chain.assertFailed(t)
		assert.Equal(t, float64(0), num.Raw())
	})

	t.Run("multiple_base", func(t *testing.T) {
		value := NewString(reporter, "100")
		num := value.AsNumber(10, 16)
		value.chain.assertFailed(t)
		num.chain.assertFailed(t)
		assert.Equal(t, float64(0), num.Raw())
	})
}

func TestString_AsBoolean(t *testing.T) {
	reporter := newMockReporter(t)

	trueValues := []string{"true", "True"}
	falseValues := []string{"false", "False"}
	badValues := []string{"TRUE", "FALSE", "t", "f", "1", "0", "bad"}

	for _, str := range trueValues {
		t.Run(str, func(t *testing.T) {
			value := NewString(reporter, str)

			b := value.AsBoolean()
			b.chain.assertNotFailed(t)

			assert.True(t, b.Raw())
		})
	}

	for _, str := range falseValues {
		t.Run(str, func(t *testing.T) {
			value := NewString(reporter, str)

			b := value.AsBoolean()
			b.chain.assertNotFailed(t)

			assert.False(t, b.Raw())
		})
	}

	for _, str := range badValues {
		t.Run(str, func(t *testing.T) {
			value := NewString(reporter, str)

			b := value.AsBoolean()
			b.chain.assertFailed(t)
		})
	}
}

func TestString_AsDateTime(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("default_formats_RFC1123+GMT", func(t *testing.T) {
		value := NewString(reporter, "Tue, 15 Nov 1994 08:12:31 GMT")
		dt := value.AsDateTime()
		value.chain.assertNotFailed(t)
		dt.chain.assertNotFailed(t)
		assert.True(t, time.Date(1994, 11, 15, 8, 12, 31, 0, time.UTC).Equal(dt.Raw()))
	})

	t.Run("RFC822", func(t *testing.T) {
		value := NewString(reporter, "15 Nov 94 08:12 GMT")
		dt := value.AsDateTime(time.RFC822)
		value.chain.assertNotFailed(t)
		dt.chain.assertNotFailed(t)
		assert.True(t, time.Date(1994, 11, 15, 8, 12, 0, 0, time.UTC).Equal(dt.Raw()))
	})

	t.Run("bad_input", func(t *testing.T) {
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

			value := NewString(reporter, str)
			dt := value.AsDateTime()
			dt.chain.assertNotFailed(t)
		})

		t.Run("all_formats_"+f, func(t *testing.T) {
			str := time.Now().Format(f)

			value := NewString(reporter, str)
			dt := value.AsDateTime(formats...)
			dt.chain.assertNotFailed(t)
		})

		t.Run("same_format_"+f, func(t *testing.T) {
			str := time.Now().Format(f)

			value := NewString(reporter, str)
			dt := value.AsDateTime(f)
			dt.chain.assertNotFailed(t)
		})

		if n != 0 {
			t.Run("different_format_"+f, func(t *testing.T) {
				str := time.Now().Format(f)

				value := NewString(reporter, str)
				dt := value.AsDateTime(formats[0])
				dt.chain.assertFailed(t)
			})
		}
	}
}

func TestString_HasPrefix(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "Hello World")

	value.HasPrefix("Hello")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.HasPrefix("Hello World")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.HasPrefix("")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.HasPrefix("Hello!")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.HasPrefix("hello")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.HasPrefix("World")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotHasPrefix("Bye")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotHasPrefix("Hello")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotHasPrefix("hello")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}

func TestString_HasSuffix(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "Hello World")

	value.HasSuffix("World")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.HasSuffix("Hello World")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.HasSuffix("")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.HasPrefix("World!")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.HasSuffix("world")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.HasSuffix("Hello")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotHasSuffix("Bye")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotHasSuffix("World")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotHasSuffix("world")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}

func TestString_HasPrefixFold(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "Hello World")

	value.HasPrefixFold("hello")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.HasPrefixFold("HeLlO wOrLd")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.HasPrefixFold("")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.HasPrefixFold("World")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotHasPrefixFold("Bye")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotHasPrefixFold("world")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotHasPrefixFold("world!")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotHasPrefixFold("hello")
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestString_HasSuffixFold(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "Hello World")

	value.HasSuffixFold("world")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.HasSuffixFold("hElLo WoRlD")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.HasSuffixFold("")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.HasSuffixFold("hello")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.HasSuffixFold("world!")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotHasSuffixFold("Bye")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotHasSuffixFold("world")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotHasSuffixFold("world!")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}
