package httpexpect

import (
	"encoding/json"
	"reflect"
)

func canonNumber(checker Checker, number interface{}) (f float64, ok bool) {
	ok = true
	defer func() {
		if err := recover(); err != nil {
			checker.Fail("%v", err)
			ok = false
		}
	}()
	f = reflect.ValueOf(number).Convert(reflect.TypeOf(float64(0))).Float()
	return
}

func canonArray(checker Checker, in []interface{}) ([]interface{}, bool) {
	out, ok := canonValue(checker, in)
	if out == nil || !ok {
		return nil, ok
	} else {
		return out.([]interface{}), ok
	}
}

func canonMap(checker Checker, in map[string]interface{}) (map[string]interface{}, bool) {
	out, ok := canonValue(checker, in)
	if out == nil || !ok {
		return nil, ok
	} else {
		return out.(map[string]interface{}), ok
	}
}

func canonValue(checker Checker, in interface{}) (interface{}, bool) {
	b, err := json.Marshal(in)
	if err != nil {
		checker.Fail(err.Error())
		return nil, false
	}

	var out interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		checker.Fail(err.Error())
		return nil, false
	}

	return out, true
}
