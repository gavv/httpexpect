package httpexpect

import (
	"github.com/gavv/gojsondiff"
	"github.com/gavv/gojsondiff/formatter"
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

func canonArray(checker Checker, in interface{}) ([]interface{}, bool) {
	var out []interface{}
	data, ok := canonValue(checker, in)
	if ok {
		out, ok = data.([]interface{})
		if !ok {
			checker.Fail("expected array, got %v", out)
		}
	}
	return out, ok
}

func canonMap(checker Checker, in interface{}) (map[string]interface{}, bool) {
	var out map[string]interface{}
	data, ok := canonValue(checker, in)
	if ok {
		out, ok = data.(map[string]interface{})
		if !ok {
			checker.Fail("expected map, got %v", out)
		}
	}
	return out, ok
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

func dumpValue(checker Checker, value interface{}) string {
	b, err := json.MarshalIndent(value, " ", "  ")
	if err != nil {
		checker.Fail(err.Error())
	}
	return " " + string(b)
}

func diffValues(checker Checker, expected, actual interface{}) string {
	differ := gojsondiff.New()

	var diff gojsondiff.Diff

	if ve, ok := expected.(map[string]interface{}); ok {
		if va, ok := actual.(map[string]interface{}); ok {
			diff = differ.CompareObjects(ve, va)
		} else {
			return " (unavailable)"
		}
	} else if ve, ok := expected.([]interface{}); ok {
		if va, ok := actual.([]interface{}); ok {
			diff = differ.CompareArrays(ve, va)
		} else {
			return " (unavailable)"
		}
	} else {
		return " (unavailable)"
	}

	formatter := formatter.NewAsciiFormatter(expected)
	formatter.ShowArrayIndex = true

	str, err := formatter.Format(diff)
	if err != nil {
		checker.Fail(err.Error())
		return ""
	}

	return "--- expected\n+++ actual\n" + str
}
