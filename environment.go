package httpexpect

import (
	"errors"
	"github.com/spf13/cast"
)

var (
	ErrEnvKeyNotFound = errors.New("key not found")
)

// Env is a context for holding data.
type Env struct {
	data map[string]interface{}
}

// Put puts the value with key in this environment.
func (e *Env) Put(key string, value interface{}) {
	e.data[key] = value
}

// Has returns a true if exists a value, otherwise false.
func (e *Env) Has(key string) bool {
	_, ok := e.data[key]
	return ok
}

// Get returns a raw value with true if exists with a given key, otherwise returns nil with false.
func (e *Env) Get(key string) (interface{}, bool) {
	v, ok := e.data[key]
	return v, ok
}

// GetString returns a string value if exists with a given key and possible to cast string type,
// otherwise returns an error.
func (e *Env) GetString(key string) (string, error) {
	v, ok := e.data[key]
	if !ok {
		return "", ErrEnvKeyNotFound
	}
	return cast.ToStringE(v)
}

// GetInt64 returns an int64 value if exists with a given key and possible to cast int64 type,
// otherwise returns an error.
func (e *Env) GetInt64(key string) (int64, error) {
	v, ok := e.data[key]
	if !ok {
		return 0, ErrEnvKeyNotFound
	}
	return cast.ToInt64E(v)
}

// GetBool returns a bool value if exists with a given key and possible to cast bool type,
// otherwise returns an error.
func (e *Env) GetBool(key string) (bool, error) {
	v, ok := e.data[key]
	if !ok {
		return false, ErrEnvKeyNotFound
	}
	return cast.ToBoolE(v)
}

// GetStringSlice returns string values if exists with a given key and possible to cast []string type,
// otherwise returns an error.
func (e *Env) GetStringSlice(key string) ([]string, error) {
	v, ok := e.data[key]
	if !ok {
		return nil, ErrEnvKeyNotFound
	}
	return cast.ToStringSliceE(v)
}
