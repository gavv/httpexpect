package httpexpect

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStringFailed(t *testing.T) {
	chain := newMockChain(t)
	chain.fail(AssertionFailure{})

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
	value.Match("")
	value.NotMatch("")
	value.MatchAll("")
}

func TestStringGetters(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "foo")

	assert.Equal(t, "foo", value.Raw())
	value.chain.assertOK(t)
	value.chain.reset()

	assert.Equal(t, "foo", value.Path("$").Raw())
	value.chain.assertOK(t)
	value.chain.reset()

	value.Schema(`{"type": "string"}`)
	value.chain.assertOK(t)
	value.chain.reset()

	value.Schema(`{"type": "object"}`)
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestStringLength(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "1234567")

	num := value.Length()
	value.chain.assertOK(t)
	num.chain.assertOK(t)
	assert.Equal(t, 7.0, num.Raw())
}

func TestStringEmpty(t *testing.T) {
	reporter := newMockReporter(t)

	value1 := NewString(reporter, "")

	value1.Empty()
	value1.chain.assertOK(t)
	value1.chain.reset()

	value1.NotEmpty()
	value1.chain.assertFailed(t)
	value1.chain.reset()

	value2 := NewString(reporter, "a")

	value2.Empty()
	value2.chain.assertFailed(t)
	value2.chain.reset()

	value2.NotEmpty()
	value2.chain.assertOK(t)
	value2.chain.reset()
}

func TestStringEqual(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "foo")

	assert.Equal(t, "foo", value.Raw())

	value.Equal("foo")
	value.chain.assertOK(t)
	value.chain.reset()

	value.Equal("FOO")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqual("FOO")
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotEqual("foo")
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestStringEqualFold(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "foo")

	value.EqualFold("foo")
	value.chain.assertOK(t)
	value.chain.reset()

	value.EqualFold("FOO")
	value.chain.assertOK(t)
	value.chain.reset()

	value.EqualFold("foo2")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqualFold("foo")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqualFold("FOO")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqualFold("foo2")
	value.chain.assertOK(t)
	value.chain.reset()
}

func TestStringContains(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "11-foo-22")

	value.Contains("foo")
	value.chain.assertOK(t)
	value.chain.reset()

	value.Contains("FOO")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotContains("FOO")
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotContains("foo")
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestStringContainsFold(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "11-foo-22")

	value.ContainsFold("foo")
	value.chain.assertOK(t)
	value.chain.reset()

	value.ContainsFold("FOO")
	value.chain.assertOK(t)
	value.chain.reset()

	value.ContainsFold("foo3")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotContainsFold("foo")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotContainsFold("FOO")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotContainsFold("foo3")
	value.chain.assertOK(t)
	value.chain.reset()
}

func TestStringMatchOne(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "http://example.com/users/john")

	m1 := value.Match(`http://(?P<host>.+)/users/(?P<user>.+)`)
	m1.chain.assertOK(t)

	assert.Equal(t,
		[]string{"http://example.com/users/john", "example.com", "john"},
		m1.submatches)

	m2 := value.Match(`http://(.+)/users/(.+)`)
	m2.chain.assertOK(t)

	assert.Equal(t,
		[]string{"http://example.com/users/john", "example.com", "john"},
		m2.submatches)
}

func TestStringMatchAll(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter,
		"http://example.com/users/john http://example.com/users/bob")

	m := value.MatchAll(`http://(\S+)/users/(\S+)`)

	assert.Equal(t, 2, len(m))

	m[0].chain.assertOK(t)
	m[1].chain.assertOK(t)

	assert.Equal(t,
		[]string{"http://example.com/users/john", "example.com", "john"},
		m[0].submatches)

	assert.Equal(t,
		[]string{"http://example.com/users/bob", "example.com", "bob"},
		m[1].submatches)
}

func TestStringMatchStatus(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "a")

	value.Match(`a`)
	value.chain.assertOK(t)
	value.chain.reset()

	value.MatchAll(`a`)
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotMatch(`a`)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Match(`[^a]`)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.MatchAll(`[^a]`)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotMatch(`[^a]`)
	value.chain.assertOK(t)
	value.chain.reset()

	assert.Equal(t, []string{}, value.Match(`[^a]`).submatches)
	assert.Equal(t, []Match{}, value.MatchAll(`[^a]`))
}

func TestStringMatchInvalid(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewString(reporter, "a")

	value.Match(`[`)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.MatchAll(`[`)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotMatch(`[`)
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestStringAsNumber(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("default_base", func(t *testing.T) {
		value1 := NewString(reporter, "1234567")
		num1 := value1.AsNumber()
		value1.chain.assertOK(t)
		num1.chain.assertOK(t)
		assert.Equal(t, float64(1234567), num1.Raw())

		value2 := NewString(reporter, "11.22")
		num2 := value2.AsNumber()
		value2.chain.assertOK(t)
		num2.chain.assertOK(t)
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
		value1.chain.assertOK(t)
		num1.chain.assertOK(t)
		assert.Equal(t, float64(100), num1.Raw())

		value2 := NewString(reporter, "11.22")
		num2 := value2.AsNumber(10)
		value2.chain.assertOK(t)
		num2.chain.assertOK(t)
		assert.Equal(t, float64(11.22), num2.Raw())
	})

	t.Run("base16", func(t *testing.T) {
		value1 := NewString(reporter, "100")
		num1 := value1.AsNumber(16)
		value1.chain.assertOK(t)
		num1.chain.assertOK(t)
		assert.Equal(t, float64(256), num1.Raw())

		value2 := NewString(reporter, "11.22")
		num2 := value2.AsNumber(16)
		value2.chain.assertFailed(t)
		num2.chain.assertFailed(t)
		assert.Equal(t, float64(0), num2.Raw())

		value3 := NewString(reporter, "8000000000000000")
		num3 := value3.AsNumber(16)
		value3.chain.assertOK(t)
		num3.chain.assertOK(t)
		assert.Equal(t, float64(9223372036854775808), num3.Raw())
	})

	t.Run("multiple_base", func(t *testing.T) {
		value1 := NewString(reporter, "100")
		num1 := value1.AsNumber(10, 16)
		value1.chain.assertFailed(t)
		num1.chain.assertFailed(t)
		assert.Equal(t, float64(0), num1.Raw())
	})
}

func TestStringAsBoolean(t *testing.T) {
	reporter := newMockReporter(t)

	trueValues := []string{"true", "True"}
	falseValues := []string{"false", "False"}
	badValues := []string{"TRUE", "FALSE", "t", "f", "1", "0", "bad"}

	for _, str := range trueValues {
		value := NewString(reporter, str)

		b := value.AsBoolean()
		b.chain.assertOK(t)

		assert.True(t, b.Raw())
	}

	for _, str := range falseValues {
		value := NewString(reporter, str)

		b := value.AsBoolean()
		b.chain.assertOK(t)

		assert.False(t, b.Raw())
	}

	for _, str := range badValues {
		value := NewString(reporter, str)

		b := value.AsBoolean()
		b.chain.assertFailed(t)
	}
}

func TestStringAsDateTime(t *testing.T) {
	reporter := newMockReporter(t)

	value1 := NewString(reporter, "Tue, 15 Nov 1994 08:12:31 GMT")
	dt1 := value1.AsDateTime()
	value1.chain.assertOK(t)
	dt1.chain.assertOK(t)
	assert.True(t, time.Date(1994, 11, 15, 8, 12, 31, 0, time.UTC).Equal(dt1.Raw()))

	value2 := NewString(reporter, "15 Nov 94 08:12 GMT")
	dt2 := value2.AsDateTime(time.RFC822)
	value2.chain.assertOK(t)
	dt2.chain.assertOK(t)
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
		dt1.chain.assertOK(t)

		value2 := NewString(reporter, str)
		dt2 := value2.AsDateTime(formats...)
		dt2.chain.assertOK(t)

		value3 := NewString(reporter, str)
		dt3 := value3.AsDateTime(f)
		dt3.chain.assertOK(t)

		if n != 0 {
			value4 := NewString(reporter, str)
			dt4 := value4.AsDateTime(formats[0])
			dt4.chain.assertFailed(t)
		}
	}
}
