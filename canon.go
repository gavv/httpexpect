package httpexpect

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"reflect"
)

func canonNumber(opChain *chain, in interface{}) (out big.Float, ok bool) {
	ok = true
	defer func() {
		if err := recover(); err != nil {
			opChain.fail(AssertionFailure{
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
	switch in := in.(type) {
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
		val := in
		return *big.NewFloat(0).SetInt(&val), true
	case big.Float:
		return in, true
	case json.Number:
		data := in.String()
		num, ok := big.NewFloat(0).SetString(data)
		return *num, ok
	default:
		return *big.NewFloat(0), false
	}
}

func canonArray(opChain *chain, in interface{}) ([]interface{}, bool) {
	var out []interface{}
	data, ok := canonValue(opChain, in)
	if ok {
		out, ok = data.([]interface{})
		if !ok {
			opChain.fail(AssertionFailure{
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

func canonMap(opChain *chain, in interface{}) (map[string]interface{}, bool) {
	var out map[string]interface{}
	data, ok := canonValue(opChain, in)
	if ok {
		out, ok = data.(map[string]interface{})
		if !ok {
			opChain.fail(AssertionFailure{
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

func canonValue(opChain *chain, in interface{}) (interface{}, bool) {
	b, err := json.Marshal(in)
	if err != nil {
		opChain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{in},
			Errors: []error{
				errors.New("expected: marshalable value"),
				err,
			},
		})
		return nil, false
	}

	reader := bytes.NewReader(b)
	dec := json.NewDecoder(reader)
	dec.UseNumber()

	var out interface{}

	for {
		if err := dec.Decode(&out); err == io.EOF {
			break
		} else if err != nil {
			opChain.fail(AssertionFailure{
				Type:   AssertValid,
				Actual: &AssertionValue{in},
				Errors: []error{
					errors.New("expected: unmarshalable value"),
					err,
				},
			})
			return nil, false
		}
	}

	return out, true
}

func canonDecode(opChain *chain, value interface{}, target interface{}) {
	if target == nil {
		opChain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected nil target argument"),
			},
		})
		return
	}

	b, err := json.Marshal(value)
	if err != nil {
		opChain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: marshable value"),
			},
		})
		return
	}

	reader := bytes.NewReader(b)
	dec := json.NewDecoder(reader)
	dec.UseNumber()

	for {
		if err := dec.Decode(&target); err == io.EOF {
			break
		} else if err != nil {
			opChain.fail(AssertionFailure{
				Type:   AssertValid,
				Actual: &AssertionValue{target},
				Errors: []error{
					errors.New("expected: value can be decoded into target argument"),
				},
			})
			return
		}
	}
}
