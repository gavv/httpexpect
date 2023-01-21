package httpexpect

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestString_Failed(t *testing.T) {
	chain := newMockChain(t)
	chain.fail(mockFailure())

	value := newString(chain, "")

	value.Path("$")
	value.Schema("")

	value.Length()
	value.AsBoolean()
	value.AsNumber()
	value.AsDateTime()
	value.Empty()
	value.NotEmpty()
	value.Equal("")
	value.NotEqual("")
	value.EqualFold("")
	value.NotEqualFold("")
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
	value.Match("")
	value.NotMatch("")
	value.MatchAll("")
	value.IsASCII()
	value.NotASCII()
}

func TestString_Constructors(t *testing.T) {
	t.Run("Constructor without config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewString(reporter, "Hello")
		value.Equal("Hello")
		value.chain.assertNotFailed(t)
	})

	t.Run("Constructor with config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewStringC(Config{
			Reporter: reporter,
		}, "Hello")
		value.Equal("Hello")
		value.chain.assertNotFailed(t)
	})

	t.Run("chain Constructor", func(t *testing.T) {
		chain := newMockChain(t)
		value := newString(chain, "Hello")
		assert.NotSame(t, value.chain, chain)
		assert.Equal(t, value.chain.context.Path, chain.context.Path)
	})
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

	value1 := NewString(reporter, "")

	value1.Empty()
	value1.chain.assertNotFailed(t)
	value1.chain.clearFailed()

	value1.NotEmpty()
	value1.chain.assertFailed(t)
	value1.chain.clearFailed()

	value2 := NewString(reporter, "a")

	value2.Empty()
	value2.chain.assertFailed(t)
	value2.chain.clearFailed()

	value2.NotEmpty()
	value2.chain.assertNotFailed(t)
	value2.chain.clearFailed()
}

func TestString_Equal(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "foo")

	assert.Equal(t, "foo", value.Raw())

	value.Equal("foo")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Equal("FOO")
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

	value.EqualFold("foo")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.EqualFold("FOO")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.EqualFold("foo2")
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

	value1 := NewString(reporter, "Ascii")
	value1.IsASCII()
	value1.chain.assertNotFailed(t)
	value1.chain.clearFailed()

	value2 := NewString(reporter, "Ascii is アスキー")
	value2.IsASCII()
	value2.chain.assertFailed(t)
	value2.chain.clearFailed()

	value3 := NewString(reporter, "アスキー")
	value3.IsASCII()
	value3.chain.assertFailed(t)
	value3.chain.clearFailed()

	value4 := NewString(reporter, string(rune(127)))
	value4.IsASCII()
	value4.chain.assertNotFailed(t)
	value4.chain.clearFailed()

	value5 := NewString(reporter, string(rune(128)))
	value5.IsASCII()
	value5.chain.assertFailed(t)
	value5.chain.clearFailed()
}

func TestString_IsNotAscii(t *testing.T) {
	reporter := newMockReporter(t)

	value1 := NewString(reporter, "Ascii")
	value1.NotASCII()
	value1.chain.assertFailed(t)
	value1.chain.clearFailed()

	value2 := NewString(reporter, "Ascii is アスキー")
	value2.NotASCII()
	value2.chain.assertNotFailed(t)
	value2.chain.clearFailed()

	value3 := NewString(reporter, "アスキー")
	value3.NotASCII()
	value3.chain.assertNotFailed(t)
	value3.chain.clearFailed()

	value4 := NewString(reporter, string(rune(127)))
	value4.NotASCII()
	value4.chain.assertFailed(t)
	value4.chain.clearFailed()

	value5 := NewString(reporter, string(rune(128)))
	value5.NotASCII()
	value5.chain.assertNotFailed(t)
	value5.chain.clearFailed()
}

func TestString_AsNumber(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("default_base", func(t *testing.T) {
		value1 := NewString(reporter, "1234567")
		num1 := value1.AsNumber()
		value1.chain.assertNotFailed(t)
		num1.chain.assertNotFailed(t)
		assert.Equal(t, float64(1234567), num1.Raw())

		value2 := NewString(reporter, "11.22")
		num2 := value2.AsNumber()
		value2.chain.assertNotFailed(t)
		num2.chain.assertNotFailed(t)
		assert.Equal(t, float64(11.22), num2.Raw())

		value3 := NewString(reporter, "a1")
		num3 := value3.AsNumber()
		value3.chain.assertFailed(t)
		num3.chain.assertFailed(t)
		assert.Equal(t, float64(0), num3.Raw())
	})

	t.Run("base10", func(t *testing.T) {
		value1 := NewString(reporter, "100")
		num1 := value1.AsNumber(10)
		value1.chain.assertNotFailed(t)
		num1.chain.assertNotFailed(t)
		assert.Equal(t, float64(100), num1.Raw())

		value2 := NewString(reporter, "11.22")
		num2 := value2.AsNumber(10)
		value2.chain.assertNotFailed(t)
		num2.chain.assertNotFailed(t)
		assert.Equal(t, float64(11.22), num2.Raw())
	})

	t.Run("base16", func(t *testing.T) {
		value1 := NewString(reporter, "100")
		num1 := value1.AsNumber(16)
		value1.chain.assertNotFailed(t)
		num1.chain.assertNotFailed(t)
		assert.Equal(t, float64(0x100), num1.Raw())

		value2 := NewString(reporter, "11.22")
		num2 := value2.AsNumber(16)
		value2.chain.assertFailed(t)
		num2.chain.assertFailed(t)
		assert.Equal(t, float64(0), num2.Raw())

		value3 := NewString(reporter, "4000000000000000")
		num3 := value3.AsNumber(16)
		value3.chain.assertNotFailed(t)
		num3.chain.assertNotFailed(t)
		assert.Equal(t, float64(0x4000000000000000), num3.Raw())
	})

	t.Run("float_precision", func(t *testing.T) {
		value1 := NewString(reporter, "4611686018427387905")
		num1 := value1.AsNumber()
		value1.chain.assertFailed(t)
		num1.chain.assertFailed(t)
		assert.Equal(t, float64(0), num1.Raw())

		value2 := NewString(reporter, "4611686018427387905")
		num2 := value2.AsNumber(10)
		value2.chain.assertFailed(t)
		num2.chain.assertFailed(t)
		assert.Equal(t, float64(0), num2.Raw())

		value3 := NewString(reporter, "8000000000000001")
		num3 := value3.AsNumber(16)
		value3.chain.assertFailed(t)
		num3.chain.assertFailed(t)
		assert.Equal(t, float64(0), num3.Raw())

		value4 := NewString(reporter, "-4000000000000001")
		num4 := value4.AsNumber(16)
		value4.chain.assertFailed(t)
		num4.chain.assertFailed(t)
		assert.Equal(t, float64(0), num4.Raw())
	})

	t.Run("multiple_base", func(t *testing.T) {
		value1 := NewString(reporter, "100")
		num1 := value1.AsNumber(10, 16)
		value1.chain.assertFailed(t)
		num1.chain.assertFailed(t)
		assert.Equal(t, float64(0), num1.Raw())
	})
}

func TestString_AsBoolean(t *testing.T) {
	reporter := newMockReporter(t)

	trueValues := []string{"true", "True"}
	falseValues := []string{"false", "False"}
	badValues := []string{"TRUE", "FALSE", "t", "f", "1", "0", "bad"}

	for _, str := range trueValues {
		value := NewString(reporter, str)

		b := value.AsBoolean()
		b.chain.assertNotFailed(t)

		assert.True(t, b.Raw())
	}

	for _, str := range falseValues {
		value := NewString(reporter, str)

		b := value.AsBoolean()
		b.chain.assertNotFailed(t)

		assert.False(t, b.Raw())
	}

	for _, str := range badValues {
		value := NewString(reporter, str)

		b := value.AsBoolean()
		b.chain.assertFailed(t)
	}
}

func TestString_AsDateTime(t *testing.T) {
	reporter := newMockReporter(t)

	value1 := NewString(reporter, "Tue, 15 Nov 1994 08:12:31 GMT")
	dt1 := value1.AsDateTime()
	value1.chain.assertNotFailed(t)
	dt1.chain.assertNotFailed(t)
	assert.True(t, time.Date(1994, 11, 15, 8, 12, 31, 0, time.UTC).Equal(dt1.Raw()))

	value2 := NewString(reporter, "15 Nov 94 08:12 GMT")
	dt2 := value2.AsDateTime(time.RFC822)
	value2.chain.assertNotFailed(t)
	dt2.chain.assertNotFailed(t)
	assert.True(t, time.Date(1994, 11, 15, 8, 12, 0, 0, time.UTC).Equal(dt2.Raw()))

	value3 := NewString(reporter, "bad")
	dt3 := value3.AsDateTime()
	value3.chain.assertFailed(t)
	dt3.chain.assertFailed(t)
	assert.True(t, time.Unix(0, 0).Equal(dt3.Raw()))

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
		str := time.Now().Format(f)

		value1 := NewString(reporter, str)
		dt1 := value1.AsDateTime()
		dt1.chain.assertNotFailed(t)

		value2 := NewString(reporter, str)
		dt2 := value2.AsDateTime(formats...)
		dt2.chain.assertNotFailed(t)

		value3 := NewString(reporter, str)
		dt3 := value3.AsDateTime(f)
		dt3.chain.assertNotFailed(t)

		if n != 0 {
			value4 := NewString(reporter, str)
			dt4 := value4.AsDateTime(formats[0])
			dt4.chain.assertFailed(t)
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
