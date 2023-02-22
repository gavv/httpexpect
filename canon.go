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

	if in != in {
		out, ok = *big.NewFloat(0), false
		return
	}

	out, ok = canonNumberConvert(in)
	if !ok {
		opChain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{in},
			Errors: []error{
				errors.New("expected: valid number"),
			},
		})
		ok = false
	}
	return
}

func canonNumberConvert(in interface{}) (out big.Float, ok bool) {
	value := reflect.ValueOf(in)
	switch in := in.(type) {
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
		return canonConvertNumberNative(value, in)
	}
}

func canonConvertNumberNative(
	value reflect.Value,
	in interface{},
) (out big.Float, ok bool) {
	t := reflect.TypeOf(in).Kind()
	switch t {
	case reflect.Float64:
		float := value.Float()
		return *big.NewFloat(float), true
	case reflect.Float32:
		float := value.Float()
		return *big.NewFloat(float), true
	case reflect.Int8:
		int := value.Int()
		return *big.NewFloat(0).SetInt64(int), true
	case reflect.Int16:
		int := value.Int()
		return *big.NewFloat(0).SetInt64(int), true
	case reflect.Int32:
		int := value.Int()
		return *big.NewFloat(0).SetInt64(int), true
	case reflect.Int64:
		int := value.Int()
		return *big.NewFloat(0).SetInt64(int), true
	case reflect.Int:
		int := value.Int()
		return *big.NewFloat(0).SetInt64(int), true
	case reflect.Uint8:
		int := value.Uint()
		return *big.NewFloat(0).SetUint64(int), true
	case reflect.Uint16:
		int := value.Uint()
		return *big.NewFloat(0).SetUint64(int), true
	case reflect.Uint32:
		int := value.Uint()
		return *big.NewFloat(0).SetUint64(int), true
	case reflect.Uint64:
		int := value.Uint()
		return *big.NewFloat(0).SetUint64(int), true
	case reflect.Uint:
		int := value.Uint()
		return *big.NewFloat(0).SetUint64(int), true
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

	jsonDecode(opChain, b, target)
}

func jsonDecode(opChain *chain, b []byte, target interface{}) {
	reader := bytes.NewReader(b)
	dec := json.NewDecoder(reader)
	dec.UseNumber()

	for {
		if err := dec.Decode(target); err == io.EOF {
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

func canonNumberDecode(opChain *chain, value big.Float, target interface{}) {
	if target == nil {
		opChain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected nil target argument"),
			},
		})
		return
	}
	t := reflect.Indirect(reflect.ValueOf(target)).Kind()
	switch t {
	case reflect.Float64:
		f, _ := value.Float64()
		canonDecode(opChain, f, target)
	case reflect.Float32:
		f, _ := value.Float32()
		canonDecode(opChain, f, target)
	case reflect.Int8:
		i, _ := value.Int64()
		canonDecode(opChain, i, target)
	case reflect.Int16:
		i, _ := value.Int64()
		canonDecode(opChain, i, target)
	case reflect.Int32:
		i, _ := value.Int64()
		canonDecode(opChain, i, target)
	case reflect.Int64:
		i, _ := value.Int64()
		canonDecode(opChain, i, target)
	case reflect.Int:
		i, _ := value.Int64()
		canonDecode(opChain, i, target)
	case reflect.Uint8:
		i, _ := value.Int64()
		canonDecode(opChain, i, target)
	case reflect.Uint16:
		i, _ := value.Int64()
		canonDecode(opChain, i, target)
	case reflect.Uint32:
		i, _ := value.Int64()
		canonDecode(opChain, i, target)
	case reflect.Uint64:
		i, _ := value.Int64()
		canonDecode(opChain, i, target)
	case reflect.Uint:
		i, _ := value.Int64()
		canonDecode(opChain, i, target)
	case reflect.Interface:
		f, _ := value.Float64()
		canonDecode(opChain, f, target)
	default:
		canonDecode(opChain, value, target)
	}
}
