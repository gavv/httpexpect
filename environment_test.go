package httpexpect

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEnvironmentGeneric(t *testing.T) {
	env := newEnvironment(newMockChain(t))

	assert.False(t, env.Has("good_key"))
	env.chain.assertOK(t)

	env.Put("good_key", 123)
	env.chain.assertOK(t)

	assert.True(t, env.Has("good_key"))
	assert.NotNil(t, env.Get("good_key"))
	assert.Equal(t, 123, env.Get("good_key").(int))
	env.chain.assertOK(t)

	assert.False(t, env.Has("bad_key"))
	env.chain.assertOK(t)

	assert.Nil(t, env.Get("bad_key"))
	env.chain.assertFailed(t)
}

func TestEnvironmentNotFound(t *testing.T) {
	env := newEnvironment(newMockChain(t))

	assert.Nil(t, env.Get("bad_key"))
	env.chain.assertFailed(t)
	env.chain.reset()

	assert.Equal(t, false, env.GetBool("bad_key"))
	env.chain.assertFailed(t)
	env.chain.reset()

	assert.Equal(t, 0, env.GetInt("bad_key"))
	env.chain.assertFailed(t)
	env.chain.reset()

	assert.Equal(t, 0.0, env.GetFloat("bad_key"))
	env.chain.assertFailed(t)
	env.chain.reset()

	assert.Equal(t, "", env.GetString("bad_key"))
	env.chain.assertFailed(t)
	env.chain.reset()

	assert.Nil(t, env.GetBytes("bad_key"))
	env.chain.assertFailed(t)
	env.chain.reset()

	assert.Equal(t, time.Duration(0), env.GetDuration("bad_key"))
	env.chain.assertFailed(t)
	env.chain.reset()

	assert.Equal(t, time.Unix(0, 0), env.GetTime("bad_key"))
	env.chain.assertFailed(t)
	env.chain.reset()
}

func TestEnvironmentBool(t *testing.T) {
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
				env.chain.assertOK(t)

				val := env.GetBool("key")
				assert.Equal(t, tt.get, val)

				if tt.ok {
					env.chain.assertOK(t)
				} else {
					env.chain.assertFailed(t)
				}
			})
	}
}

func TestEnvironmentInt(t *testing.T) {
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
				env.chain.assertOK(t)

				val := env.GetInt("key")
				assert.Equal(t, tt.get, val)

				if tt.ok {
					env.chain.assertOK(t)
				} else {
					env.chain.assertFailed(t)
				}
			})
	}
}

func TestEnvironmentFloat(t *testing.T) {
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
				env.chain.assertOK(t)

				val := env.GetFloat("key")
				assert.Equal(t, tt.get, val)

				if tt.ok {
					env.chain.assertOK(t)
				} else {
					env.chain.assertFailed(t)
				}
			})
	}
}

func TestEnvironmentString(t *testing.T) {
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
				env.chain.assertOK(t)

				val := env.GetString("key")
				assert.Equal(t, tt.get, val)

				if tt.ok {
					env.chain.assertOK(t)
				} else {
					env.chain.assertFailed(t)
				}
			})
	}
}

func TestEnvironmentBytes(t *testing.T) {
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
				env.chain.assertOK(t)

				val := env.GetBytes("key")
				assert.Equal(t, tt.get, val)

				if tt.ok {
					env.chain.assertOK(t)
				} else {
					env.chain.assertFailed(t)
				}
			})
	}
}

func TestEnvironmentDuration(t *testing.T) {
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
				env.chain.assertOK(t)

				val := env.GetDuration("key")
				assert.Equal(t, tt.get, val)

				if tt.ok {
					env.chain.assertOK(t)
				} else {
					env.chain.assertFailed(t)
				}
			})
	}
}

func TestEnvironmentTime(t *testing.T) {
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
				env.chain.assertOK(t)

				val := env.GetTime("key")
				assert.Equal(t, tt.get, val)

				if tt.ok {
					env.chain.assertOK(t)
				} else {
					env.chain.assertFailed(t)
				}
			})
	}
}
