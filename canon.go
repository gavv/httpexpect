package httpexpect

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"reflect"
	"strconv"
)

func canonNumber(opChain *chain, in interface{}) (out *big.Float, ok bool) {
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
		out, ok = nil, false
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

func canonNumberConvert(in interface{}) (out *big.Float, ok bool) {
	value := reflect.ValueOf(in)
	switch in := in.(type) {
	case big.Int:
		val := in
		return big.NewFloat(0).SetInt(&val), true
	case big.Float:
		return &in, true
	case json.Number:
		data := in.String()
		num, ok := big.NewFloat(0).SetString(data)
		return num, ok
	default:
		return canonConvertNumberNative(value, in)
	}
}

func canonConvertNumberNative(
	value reflect.Value,
	in interface{},
) (out *big.Float, ok bool) {
	t := reflect.TypeOf(in).Kind()
	switch t {
	case reflect.Float64, reflect.Float32:
		float := value.Float()
		return big.NewFloat(float), true
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		int := value.Int()
		return big.NewFloat(0).SetInt64(int), true
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		int := value.Uint()
		return big.NewFloat(0).SetUint64(int), true
	case reflect.Invalid,
		reflect.Bool,
		reflect.Uintptr,
		reflect.Complex64,
		reflect.Complex128,
		reflect.Array,
		reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Slice,
		reflect.String,
		reflect.Struct,
		reflect.Ptr,
		reflect.UnsafePointer:
		return big.NewFloat(0), false
	default:
		return big.NewFloat(0), false
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

	var out interface{}

	jsonDecode(opChain, b, &out)
	out = convertJSONNumberToFloatOrBigFloat(out)

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
	switch t := target.(type) {
	case *interface{}:
		*t = convertJSONNumberToFloatOrBigFloat(*t)
	case *[]interface{}:
		for i, val := range *t {
			(*t)[i] = convertJSONNumberToFloatOrBigFloat(val)
		}
	case *map[string]interface{}:
		for key, val := range *t {
			(*t)[key] = convertJSONNumberToFloatOrBigFloat(val)
		}
	default:
		v := reflect.ValueOf(t)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() == reflect.Struct {
			for i := 0; i < v.NumField(); i++ {
				field := v.Field(i)
				if field.CanInterface() {
					val := field.Interface()
					newVal := convertJSONNumberToFloatOrBigFloat(val)
					field.Set(reflect.ValueOf(newVal))
				}
			}
		}
	}
}

func jsonDecode(opChain *chain, b []byte, target interface{}) {
	reader := bytes.NewReader(b)
	dec := json.NewDecoder(reader)
	dec.UseNumber()

	for {
		if err := dec.Decode(target); err == io.EOF || target == nil {
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
	case reflect.Float64, reflect.Interface:
		f, _ := value.Float64()
		canonDecode(opChain, f, target)
	case reflect.Float32:
		f, _ := value.Float32()
		canonDecode(opChain, f, target)
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		i, _ := value.Int64()
		canonDecode(opChain, i, target)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		i, _ := value.Int64()
		canonDecode(opChain, i, target)
	case reflect.Invalid,
		reflect.Bool,
		reflect.Uintptr,
		reflect.Complex64,
		reflect.Complex128,
		reflect.Array,
		reflect.Chan,
		reflect.Func,
		reflect.Map,
		reflect.Slice,
		reflect.String,
		reflect.Struct,
		reflect.Ptr,
		reflect.UnsafePointer:
		canonDecode(opChain, value, target)
	default:
		canonDecode(opChain, value, target)
	}
}

func convertJSONNumberToFloatOrBigFloat(data interface{}) interface{} {
	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Map:
		for _, key := range v.MapKeys() {
			val := v.MapIndex(key)
			if val.IsNil() {
				continue
			}
			newVal := convertJSONNumberToFloatOrBigFloat(val.Interface())
			v.SetMapIndex(key, reflect.ValueOf(newVal))
		}
		return v.Interface()
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if v.Index(i).IsNil() {
				continue
			}
			newVal := convertJSONNumberToFloatOrBigFloat(v.Index(i).Interface())
			v.Index(i).Set(reflect.ValueOf(newVal))
		}
		return v.Interface()
	case reflect.Ptr:
		if v.IsNil() {
			return nil
		}
		newVal := convertJSONNumberToFloatOrBigFloat(v.Elem().Interface())
		newV := reflect.New(v.Type().Elem())
		newV.Elem().Set(reflect.ValueOf(newVal))
		return newV.Interface()
	case reflect.Interface:
		newVal := convertJSONNumberToFloatOrBigFloat(v.Elem().Interface())
		return reflect.ValueOf(newVal).Interface()
	case reflect.String:
		if jsonNum, ok := v.Interface().(json.Number); ok {
			if hasPrecisionLoss(jsonNum) {
				newVal := big.NewFloat(0)
				newVal, _ = newVal.SetString(v.String())
				return newVal
			}
			newVal, _ := strconv.ParseFloat(v.String(), 64)
			return newVal
		}
		return data
	case reflect.Struct:
		newVal := reflect.New(v.Type()).Elem()
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if !field.CanInterface() {
				continue
			}
			newField := convertJSONNumberToFloatOrBigFloat(field.Interface())
			newVal.Field(i).Set(reflect.ValueOf(newField))
		}
		return newVal.Interface()
	case reflect.Invalid,
		reflect.Bool,
		reflect.Uintptr,
		reflect.Complex64,
		reflect.Complex128,
		reflect.Chan,
		reflect.Func,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Float32,
		reflect.Float64,
		reflect.UnsafePointer:
		return data
	default:
		return data
	}
}
