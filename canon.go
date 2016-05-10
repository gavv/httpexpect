package httpexpect

import (
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
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
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		checker.Fail(err.Error())
	}
	return string(b)
}

func diffMaps(checker Checker, expected, actual map[string]interface{}) string {
	be, err := json.Marshal(expected)
	if err != nil {
		checker.Fail(err.Error())
		return ""
	}

	ba, err := json.Marshal(actual)
	if err != nil {
		checker.Fail(err.Error())
		return ""
	}

	differ := gojsondiff.New()

	d, err := differ.Compare(be, ba)
	if err != nil {
		checker.Fail(err.Error())
		return ""
	}

	formatter := formatter.NewAsciiFormatter(expected)
	formatter.ShowArrayIndex = true

	str, err := formatter.Format(d)
	if err != nil {
		checker.Fail(err.Error())
		return ""
	}

	return "--- expected\n+++ actual\n" + str
}
