package httpexpect

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEnvironment_Constructors(t *testing.T) {
	t.Run("reporter", func(t *testing.T) {
		reporter := newMockReporter(t)
		env := NewEnvironment(reporter)
		env.chain.assertNotFailed(t)
	})

	t.Run("config", func(t *testing.T) {
		reporter := newMockReporter(t)
		env := NewEnvironmentC(Config{
			Reporter: reporter,
		})
		env.chain.assertNotFailed(t)
	})

	t.Run("chain", func(t *testing.T) {
		chain := newMockChain(t)
		value := newEnvironment(chain)
		assert.NotSame(t, value.chain, chain)
		assert.Equal(t, value.chain.context.Path, chain.context.Path)
	})
}

func TestEnvironment_Reentrant(t *testing.T) {
	reporter := newMockReporter(t)

	env := NewEnvironment(reporter)

	reportCalled := false
	reporter.reportCb = func() {
		env.Put("good_key", 123)
		reportCalled = true
	}

	env.Get("bad_key")
	env.chain.assertFailed(t)

	assert.True(t, reportCalled)
}

func TestEnvironment_Basic(t *testing.T) {
	env := newEnvironment(newMockChain(t))

	assert.False(t, env.Has("good_key"))
	env.chain.assertNotFailed(t)

	env.Put("good_key", 123)
	env.chain.assertNotFailed(t)

	assert.True(t, env.Has("good_key"))
	assert.NotNil(t, env.Get("good_key"))
	assert.Equal(t, 123, env.Get("good_key").(int))
	env.chain.assertNotFailed(t)

	assert.False(t, env.Has("bad_key"))
	env.chain.assertNotFailed(t)

	assert.Nil(t, env.Get("bad_key"))
	env.chain.assertFailed(t)
}

func TestEnvironment_Delete(t *testing.T) {
	env := newEnvironment(newMockChain(t))

	env.Put("good_key", 123)
	env.chain.assertNotFailed(t)

	assert.True(t, env.Has("good_key"))
	assert.NotNil(t, env.Get("good_key"))
	assert.Equal(t, 123, env.Get("good_key").(int))
	env.chain.assertNotFailed(t)

	env.Delete("good_key")
	env.chain.assertNotFailed(t)

	assert.False(t, env.Has("good_key"))
	assert.Nil(t, env.Get("good_key"))
	env.chain.assertFailed(t)
	env.chain.clearFailed()
}

func TestEnvironment_NotFound(t *testing.T) {
	env := newEnvironment(newMockChain(t))

	assert.Nil(t, env.Get("bad_key"))
	env.chain.assertFailed(t)
	env.chain.clearFailed()

	assert.Equal(t, false, env.GetBool("bad_key"))
	env.chain.assertFailed(t)
	env.chain.clearFailed()

	assert.Equal(t, 0, env.GetInt("bad_key"))
	env.chain.assertFailed(t)
	env.chain.clearFailed()

	assert.Equal(t, 0.0, env.GetFloat("bad_key"))
	env.chain.assertFailed(t)
	env.chain.clearFailed()

	assert.Equal(t, "", env.GetString("bad_key"))
	env.chain.assertFailed(t)
	env.chain.clearFailed()

	assert.Nil(t, env.GetBytes("bad_key"))
	env.chain.assertFailed(t)
	env.chain.clearFailed()

	assert.Equal(t, time.Duration(0), env.GetDuration("bad_key"))
	env.chain.assertFailed(t)
	env.chain.clearFailed()

	assert.Equal(t, time.Unix(0, 0), env.GetTime("bad_key"))
	env.chain.assertFailed(t)
	env.chain.clearFailed()
}

func TestEnvironment_Bool(t *testing.T) {
	tests := []struct {
		put interface{}
		get bool
		ok  bool
	}{
		{
			put: true,
			get: true,
			ok:  true,
		},
		{
			put: 1,
			get: false,
			ok:  false,
		},
		{
			put: 1.0,
			get: false,
			ok:  false,
		},
		{
			put: "true",
			get: false,
			ok:  false,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%T-%v", tt.put, tt.put),
			func(t *testing.T) {
				env := newEnvironment(newMockChain(t))

				env.Put("key", tt.put)
				env.chain.assertNotFailed(t)

				val := env.GetBool("key")
				assert.Equal(t, tt.get, val)

				if tt.ok {
					env.chain.assertNotFailed(t)
				} else {
					env.chain.assertFailed(t)
				}
			})
	}
}

func TestEnvironment_Int(t *testing.T) {
	tests := []struct {
		put interface{}
		get int
		ok  bool
	}{
		{
			put: int(123),
			get: 123,
			ok:  true,
		},
		{
			put: int8(123),
			get: 123,
			ok:  true,
		},
		{
			put: int16(123),
			get: 123,
			ok:  true,
		},
		{
			put: int32(123),
			get: 123,
			ok:  true,
		},
		{
			put: int64(123),
			get: 123,
			ok:  true,
		},
		{
			put: uint(123),
			get: 123,
			ok:  true,
		},
		{
			put: uint8(123),
			get: 123,
			ok:  true,
		},
		{
			put: uint16(123),
			get: 123,
			ok:  true,
		},
		{
			put: uint32(123),
			get: 123,
			ok:  true,
		},
		{
			put: uint64(math.MaxUint64),
			get: 0,
			ok:  false,
		},
		{
			put: 123.0,
			get: 0,
			ok:  false,
		},
		{
			put: false,
			get: 0,
			ok:  false,
		},
		{
			put: time.Second,
			get: 0,
			ok:  false,
		},
		{
			put: "123",
			get: 0,
			ok:  false,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%T-%v", tt.put, tt.put),
			func(t *testing.T) {
				env := newEnvironment(newMockChain(t))

				env.Put("key", tt.put)
				env.chain.assertNotFailed(t)

				val := env.GetInt("key")
				assert.Equal(t, tt.get, val)

				if tt.ok {
					env.chain.assertNotFailed(t)
				} else {
					env.chain.assertFailed(t)
				}
			})
	}
}

func TestEnvironment_Float(t *testing.T) {
	tests := []struct {
		put interface{}
		get float64
		ok  bool
	}{
		{
			put: float32(123),
			get: 123.0,
			ok:  true,
		},
		{
			put: float64(123),
			get: 123.0,
			ok:  true,
		},
		{
			put: int(123),
			get: 0,
			ok:  false,
		},
		{
			put: false,
			get: 0,
			ok:  false,
		},
		{
			put: "123.0",
			get: 0,
			ok:  false,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%T-%v", tt.put, tt.put),
			func(t *testing.T) {
				env := newEnvironment(newMockChain(t))

				env.Put("key", tt.put)
				env.chain.assertNotFailed(t)

				val := env.GetFloat("key")
				assert.Equal(t, tt.get, val)

				if tt.ok {
					env.chain.assertNotFailed(t)
				} else {
					env.chain.assertFailed(t)
				}
			})
	}
}

func TestEnvironment_String(t *testing.T) {
	tests := []struct {
		put interface{}
		get string
		ok  bool
	}{
		{
			put: "test",
			get: "test",
			ok:  true,
		},
		{
			put: []byte("test"),
			get: "",
			ok:  false,
		},
		{
			put: 123,
			get: "",
			ok:  false,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%T-%v", tt.put, tt.put),
			func(t *testing.T) {
				env := newEnvironment(newMockChain(t))

				env.Put("key", tt.put)
				env.chain.assertNotFailed(t)

				val := env.GetString("key")
				assert.Equal(t, tt.get, val)

				if tt.ok {
					env.chain.assertNotFailed(t)
				} else {
					env.chain.assertFailed(t)
				}
			})
	}
}

func TestEnvironment_Bytes(t *testing.T) {
	tests := []struct {
		put interface{}
		get []byte
		ok  bool
	}{
		{
			put: []byte("test"),
			get: []byte("test"),
			ok:  true,
		},
		{
			put: "test",
			get: nil,
			ok:  false,
		},
		{
			put: 123,
			get: nil,
			ok:  false,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%T-%v", tt.put, tt.put),
			func(t *testing.T) {
				env := newEnvironment(newMockChain(t))

				env.Put("key", tt.put)
				env.chain.assertNotFailed(t)

				val := env.GetBytes("key")
				assert.Equal(t, tt.get, val)

				if tt.ok {
					env.chain.assertNotFailed(t)
				} else {
					env.chain.assertFailed(t)
				}
			})
	}
}

func TestEnvironment_Duration(t *testing.T) {
	tests := []struct {
		put interface{}
		get time.Duration
		ok  bool
	}{
		{
			put: time.Second,
			get: time.Second,
			ok:  true,
		},
		{
			put: int64(999999999),
			get: time.Duration(0),
			ok:  false,
		},
		{
			put: "1s",
			get: time.Duration(0),
			ok:  false,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%T-%v", tt.put, tt.put),
			func(t *testing.T) {
				env := newEnvironment(newMockChain(t))

				env.Put("key", tt.put)
				env.chain.assertNotFailed(t)

				val := env.GetDuration("key")
				assert.Equal(t, tt.get, val)

				if tt.ok {
					env.chain.assertNotFailed(t)
				} else {
					env.chain.assertFailed(t)
				}
			})
	}
}

func TestEnvironment_Time(t *testing.T) {
	tests := []struct {
		put interface{}
		get time.Time
		ok  bool
	}{
		{
			put: time.Unix(9999999, 0),
			get: time.Unix(9999999, 0),
			ok:  true,
		},
		{
			put: 9999999,
			get: time.Unix(0, 0),
			ok:  false,
		},
		{
			put: time.Unix(9999999, 0).String(),
			get: time.Unix(0, 0),
			ok:  false,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%T-%v", tt.put, tt.put),
			func(t *testing.T) {
				env := newEnvironment(newMockChain(t))

				env.Put("key", tt.put)
				env.chain.assertNotFailed(t)

				val := env.GetTime("key")
				assert.Equal(t, tt.get, val)

				if tt.ok {
					env.chain.assertNotFailed(t)
				} else {
					env.chain.assertFailed(t)
				}
			})
	}
}
