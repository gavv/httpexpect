package httpexpect

import (
	"errors"
	"fmt"
	"github.com/ohler55/ojg/jp"
	"reflect"
	"regexp"

	"github.com/xeipuuv/gojsonschema"
	"github.com/yalp/jsonpath"
)

func jsonPathOjg(opChain *chain, value interface{}, path string) *Value {
	if opChain.failed() {
		return newValue(opChain, nil)
	}

	expr, err := jp.ParseString(path)
	if err != nil {
		opChain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{path},
			Errors: []error{
				errors.New("expected: valid json path"),
				err,
			},
		})
		return newValue(opChain, nil)
	}
	result := expr.Get(value)
	// in order to keep the results somewhat consistent with yalp's results,
	// we return a single value where no wildcards or descends are involved.
	// TODO: it might be more consistent if it also included filters
	if len(result) == 1 && !hasWildcardsOrDescend(expr) {
		return newValue(opChain, result[0])
	}
	if result == nil {
		return newValue(opChain, []interface{}{})
	}
	return newValue(opChain, result)
}

func hasWildcardsOrDescend(expr jp.Expr) bool {
	for _, frag := range expr {
		switch frag.(type) {
		case jp.Wildcard, jp.Descent:
			return true
		}
	}
	return false
}

func jsonPath(opChain *chain, value interface{}, path string) *Value {
	if opChain.failed() {
		return newValue(opChain, nil)
	}

	filterFn, err := jsonpath.Prepare(path)
	if err != nil {
		opChain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{path},
			Errors: []error{
				errors.New("expected: valid json path"),
				err,
			},
		})
		return newValue(opChain, nil)
	}

	result, err := filterFn(value)
	if err != nil {
		opChain.fail(AssertionFailure{
			Type:     AssertMatchPath,
			Actual:   &AssertionValue{value},
			Expected: &AssertionValue{path},
			Errors: []error{
				errors.New("expected: value matches given json path"),
				err,
			},
		})
		return newValue(opChain, nil)
	}

	return newValue(opChain, result)
}

func jsonSchema(opChain *chain, value, schema interface{}) {
	if opChain.failed() {
		return
	}

	getString := func(in interface{}) (out string, ok bool) {
		ok = true
		defer func() {
			if err := recover(); err != nil {
				ok = false
			}
		}()
		out = reflect.ValueOf(in).Convert(reflect.TypeOf("")).String()
		return
	}

	var schemaLoader gojsonschema.JSONLoader
	var schemaData interface{}

	if str, ok := getString(schema); ok {
		if ok, _ := regexp.MatchString(`^\w+://`, str); ok {
			schemaLoader = gojsonschema.NewReferenceLoader(str)
			schemaData = str
		} else {
			schemaLoader = gojsonschema.NewStringLoader(str)
			schemaData, _ = schemaLoader.LoadJSON()
		}
	} else {
		schemaLoader = gojsonschema.NewGoLoader(schema)
		schemaData = schema
	}

	valueLoader := gojsonschema.NewGoLoader(value)

	result, err := gojsonschema.Validate(schemaLoader, valueLoader)
	if err != nil {
		opChain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{schema},
			Errors: []error{
				errors.New("expected: valid json schema"),
				err,
			},
		})
		return
	}

	if !result.Valid() {
		errors := []error{
			errors.New("expected: value matches given json schema"),
		}
		for _, err := range result.Errors() {
			errors = append(errors, fmt.Errorf("%s", err))
		}
		opChain.fail(AssertionFailure{
			Type:     AssertMatchSchema,
			Actual:   &AssertionValue{value},
			Expected: &AssertionValue{schemaData},
			Errors:   errors,
		})
	}
}
