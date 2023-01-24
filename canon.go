package httpexpect

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
)

func canonNumber(chain *chain, in interface{}) (out big.Float, ok bool) {
	ok = true
	defer func() {
		if err := recover(); err != nil {
			chain.fail(AssertionFailure{
				Type:   AssertValid,
				Actual: &AssertionValue{in},
				Errors: []error{
					errors.New("expected: valid number"),
					fmt.Errorf("%s", err),
				},
			})
			ok = false
		}
	}()
	out, ok = canonNumberConvert(in)
	return
}

func canonNumberConvert(in interface{}) (out big.Float, ok bool) {
	value := reflect.ValueOf(in)
	switch in.(type) {
	case float64:
		float := value.Float()
		return *big.NewFloat(float), true
	case float32:
		float := value.Float()
		return *big.NewFloat(float), true
	case int8:
		int := value.Int()
		return *big.NewFloat(0).SetInt64(int), true
	case int16:
		int := value.Int()
		return *big.NewFloat(0).SetInt64(int), true
	case int32:
		int := value.Int()
		return *big.NewFloat(0).SetInt64(int), true
	case int64:
		int := value.Int()
		return *big.NewFloat(0).SetInt64(int), true
	case uint8:
		int := value.Uint()
		return *big.NewFloat(0).SetUint64(int), true
	case uint16:
		int := value.Uint()
		return *big.NewFloat(0).SetUint64(int), true
	case uint32:
		int := value.Uint()
		return *big.NewFloat(0).SetUint64(int), true
	case uint64:
		int := value.Uint()
		return *big.NewFloat(0).SetUint64(int), true
	case big.Int:
		val, ok := in.(big.Int)
		if ok {
			return *big.NewFloat(0).SetInt(&val), true
		}
		return *big.NewFloat(0), false
	case big.Float:
		return in.(big.Float), true
	case json.Number:
		data := in.(json.Number).String()
		num, ok := big.NewFloat(0).SetString(data)
		return *num, ok
	default:
		return *big.NewFloat(0), false
	}
}

func canonArray(chain *chain, in interface{}) ([]interface{}, bool) {
	var out []interface{}
	data, ok := canonValue(chain, in)
	if ok {
		out, ok = data.([]interface{})
		if !ok {
			chain.fail(AssertionFailure{
				Type:   AssertValid,
				Actual: &AssertionValue{in},
				Errors: []error{
					errors.New("expected: valid array"),
				},
			})
		}
	}
	return out, ok
}

func canonMap(chain *chain, in interface{}) (map[string]interface{}, bool) {
	var out map[string]interface{}
	data, ok := canonValue(chain, in)
	if ok {
		out, ok = data.(map[string]interface{})
		if !ok {
			chain.fail(AssertionFailure{
				Type:   AssertValid,
				Actual: &AssertionValue{in},
				Errors: []error{
					errors.New("expected: valid map"),
				},
			})
		}
	}
	return out, ok
}

func canonValue(chain *chain, in interface{}) (interface{}, bool) {
	b, err := json.Marshal(in)
	if err != nil {
		chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{in},
			Errors: []error{
				errors.New("expected: marshalable value"),
				err,
			},
		})
		return nil, false
	}

	var out interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{in},
			Errors: []error{
				errors.New("expected: unmarshalable value"),
				err,
			},
		})
		return nil, false
	}

	return out, true
}
