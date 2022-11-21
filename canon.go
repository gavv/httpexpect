package httpexpect

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

func canonNumber(chain *chain, in interface{}) (out float64, ok bool) {
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
	out = reflect.ValueOf(in).Convert(reflect.TypeOf(float64(0))).Float()
	return
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
