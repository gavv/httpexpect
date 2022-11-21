package httpexpect

import (
	"errors"
	"time"
)

// Environment provides a container for arbitrary data shared between tests.
//
// Example:
//
//	env := NewEnvironment(t)
//	env.Put("key", "value")
//	value := env.GetString("key")
type Environment struct {
	chain *chain
	data  map[string]interface{}
}

// NewEnvironment returns a new Environment given a reporter.
//
// Reporter should not be nil.
//
// Example:
//
//	env := NewEnvironment(t)
func NewEnvironment(reporter Reporter) *Environment {
	return newEnvironment(newChainWithDefaults("Environment()", reporter))
}

func newEnvironment(parent *chain) *Environment {
	return &Environment{
		chain: parent.clone(),
		data:  make(map[string]interface{}),
	}
}

// Put saves the value with key in the environment.
//
// Example:
//
//	env := NewEnvironment(t)
//	env.Put("key1", "str")
//	env.Put("key2", 123)
func (e *Environment) Put(key string, value interface{}) {
	e.chain.enter("Put(%q)", key)
	defer e.chain.leave()

	e.data[key] = value
}

// Has returns true if value exists in the environment.
//
// Example:
//
//	if env.Has("key1") {
//	   ...
//	}
func (e *Environment) Has(key string) bool {
	e.chain.enter("Has(%q)", key)
	defer e.chain.leave()

	_, ok := e.data[key]
	return ok
}

// Get returns value stored in the environment.
//
// If value does not exist, reports failure and returns nil.
//
// Example:
//
//	value1 := env.Get("key1").(string)
//	value2 := env.Get("key1").(int)
func (e *Environment) Get(key string) interface{} {
	e.chain.enter("Get(%q)", key)
	defer e.chain.leave()

	value, _ := e.getValue(key)

	return value
}

// GetBool returns value stored in the environment, casted to bool.
//
// If value does not exist, or is not bool, reports failure and returns false.
//
// Example:
//
//	value := env.GetBool("key")
func (e *Environment) GetBool(key string) bool {
	e.chain.enter("GetBool(%q)", key)
	defer e.chain.leave()

	value, ok := e.getValue(key)
	if !ok {
		return false
	}

	casted, ok := value.(bool)
	if !ok {
		e.chain.fail(AssertionFailure{
			Type:   AssertType,
			Actual: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: bool value"),
			},
		})
		return false
	}

	return casted
}

// GetInt returns value stored in the environment, casted to int64.
//
// If value does not exist, or is not signed or unsigned integer that can be
// represented as int without overflow, reports failure and returns zero.
//
// Example:
//
//	value := env.GetInt("key")
func (e *Environment) GetInt(key string) int {
	e.chain.enter("GetInt(%q)", key)
	defer e.chain.leave()

	value, ok := e.getValue(key)
	if !ok {
		return 0
	}

	var casted int

	const (
		intSize = 32 << (^uint(0) >> 63) // 32 or 64
		maxInt  = 1<<(intSize-1) - 1
		minInt  = -1 << (intSize - 1)
	)

	switch num := value.(type) {
	case int8:
		casted = int(num)
		ok = (int64(num) >= minInt) && (int64(num) <= maxInt)
	case int16:
		casted = int(num)
		ok = (int64(num) >= minInt) && (int64(num) <= maxInt)
	case int32:
		casted = int(num)
		ok = (int64(num) >= minInt) && (int64(num) <= maxInt)
	case int64:
		casted = int(num)
		ok = (int64(num) >= minInt) && (int64(num) <= maxInt)
	case int:
		casted = num
		ok = (int64(num) >= minInt) && (int64(num) <= maxInt)

	case uint8:
		casted = int(num)
		ok = (uint64(num) <= maxInt)
	case uint16:
		casted = int(num)
		ok = (uint64(num) <= maxInt)
	case uint32:
		casted = int(num)
		ok = (uint64(num) <= maxInt)
	case uint64:
		casted = int(num)
		ok = (uint64(num) <= maxInt)
	case uint:
		casted = int(num)
		ok = (uint64(num) <= maxInt)

	default:
		e.chain.fail(AssertionFailure{
			Type:   AssertType,
			Actual: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: signed or unsigned integer"),
			},
		})
		return 0
	}

	if !ok {
		e.chain.fail(AssertionFailure{
			Type:     AssertInRange,
			Actual:   &AssertionValue{value},
			Expected: &AssertionValue{AssertionRange{minInt, maxInt}},
			Errors: []error{
				errors.New(
					"expected: value can be represented as int without overflow"),
			},
		})
		return 0
	}

	return casted
}

// GetFloat returns value stored in the environment, casted to float64.
//
// If value does not exist, or is not floating point value, reports failure
// and returns zero value.
//
// Example:
//
//	value := env.GetFloat("key")
func (e *Environment) GetFloat(key string) float64 {
	e.chain.enter("GetFloat(%q)", key)
	defer e.chain.leave()

	value, ok := e.getValue(key)
	if !ok {
		return 0
	}

	var casted float64

	switch num := value.(type) {
	case float32:
		casted = float64(num)

	case float64:
		casted = num

	default:
		e.chain.fail(AssertionFailure{
			Type:   AssertType,
			Actual: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: float32 or float64"),
			},
		})
		return 0
	}

	return casted
}

// GetString returns value stored in the environment, casted to string.
//
// If value does not exist, or is not string, reports failure and returns
// empty string.
//
// Example:
//
//	value := env.GetString("key")
func (e *Environment) GetString(key string) string {
	e.chain.enter("GetString(%q)", key)
	defer e.chain.leave()

	value, ok := e.getValue(key)
	if !ok {
		return ""
	}

	casted, ok := value.(string)
	if !ok {
		e.chain.fail(AssertionFailure{
			Type:   AssertType,
			Actual: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: string value"),
			},
		})
		return ""
	}

	return casted
}

// GetBytes returns value stored in the environment, casted to []byte.
//
// If value does not exist, or is not []byte slice, reports failure and returns nil.
//
// Example:
//
//	value := env.GetBytes("key")
func (e *Environment) GetBytes(key string) []byte {
	e.chain.enter("GetBytes(%q)", key)
	defer e.chain.leave()

	value, ok := e.getValue(key)
	if !ok {
		return nil
	}

	casted, ok := value.([]byte)
	if !ok {
		e.chain.fail(AssertionFailure{
			Type:   AssertType,
			Actual: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: []byte slice"),
			},
		})
		return nil
	}

	return casted
}

// GetDuration returns value stored in the environment, casted to time.Duration.
//
// If value does not exist, is not time.Duration, reports failure and returns
// zero duration.
//
// Example:
//
//	value := env.GetDuration("key")
func (e *Environment) GetDuration(key string) time.Duration {
	e.chain.enter("GetDuration(%q)", key)
	defer e.chain.leave()

	value, ok := e.getValue(key)
	if !ok {
		return time.Duration(0)
	}

	casted, ok := value.(time.Duration)
	if !ok {
		e.chain.fail(AssertionFailure{
			Type:   AssertType,
			Actual: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: time.Duration value"),
			},
		})
		return time.Duration(0)
	}

	return casted
}

// GetTime returns value stored in the environment, casted to time.Time.
//
// If value does not exist, is not time.Time, reports failure and returns
// zero time.
//
// Example:
//
//	value := env.GetTime("key")
func (e *Environment) GetTime(key string) time.Time {
	e.chain.enter("GetTime(%q)", key)
	defer e.chain.leave()

	value, ok := e.getValue(key)
	if !ok {
		return time.Unix(0, 0)
	}

	casted, ok := value.(time.Time)
	if !ok {
		e.chain.fail(AssertionFailure{
			Type:   AssertType,
			Actual: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: time.Time value"),
			},
		})
		return time.Unix(0, 0)
	}

	return casted
}

func (e *Environment) getValue(key string) (interface{}, bool) {
	v, ok := e.data[key]

	if !ok {
		e.chain.fail(AssertionFailure{
			Type:     AssertContainsKey,
			Actual:   &AssertionValue{e.data},
			Expected: &AssertionValue{key},
			Errors: []error{
				errors.New("expected: environment contains key"),
			},
		})
		return nil, false
	}

	return v, true
}
