package httpexpect

import (
	"encoding/json"
	"fmt"
	"github.com/gavv/gojsondiff"
	"github.com/gavv/gojsondiff/formatter"
	"io"
	"reflect"
)

type readCloserAdapter struct {
	io.Reader
}

func (b readCloserAdapter) Close() error {
	return nil
}

func canonNumber(chain *chain, number interface{}) (f float64, ok bool) {
	ok = true
	defer func() {
		if err := recover(); err != nil {
			chain.fail("%v", err)
			ok = false
		}
	}()
	f = reflect.ValueOf(number).Convert(reflect.TypeOf(float64(0))).Float()
	return
}

func canonArray(chain *chain, in interface{}) ([]interface{}, bool) {
	var out []interface{}
	data, ok := canonValue(chain, in)
	if ok {
		out, ok = data.([]interface{})
		if !ok {
			chain.fail("expected array, got %v", out)
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
			chain.fail("expected map, got %v", out)
		}
	}
	return out, ok
}

func canonValue(chain *chain, in interface{}) (interface{}, bool) {
	b, err := json.Marshal(in)
	if err != nil {
		chain.fail(err.Error())
		return nil, false
	}

	var out interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		chain.fail(err.Error())
		return nil, false
	}

	return out, true
}

func dumpValue(value interface{}) string {
	b, err := json.MarshalIndent(value, " ", "  ")
	if err != nil {
		return " " + fmt.Sprintf("%#v", value)
	}
	return " " + string(b)
}

func diffValues(expected, actual interface{}) string {
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
		return " (unavailable)"
	}

	return "--- expected\n+++ actual\n" + str
}
