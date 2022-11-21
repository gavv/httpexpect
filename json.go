package httpexpect

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"

	"github.com/xeipuuv/gojsonschema"
	"github.com/yalp/jsonpath"
)

func jsonPath(chain *chain, value interface{}, path string) *Value {
	if chain.failed() {
		return newValue(chain, nil)
	}

	filterFn, err := jsonpath.Prepare(path)
	if err != nil {
		chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{path},
			Errors: []error{
				errors.New("expected: valid json path"),
				err,
			},
		})
		return newValue(chain, nil)
	}

	result, err := filterFn(value)
	if err != nil {
		chain.fail(AssertionFailure{
			Type:     AssertMatchPath,
			Actual:   &AssertionValue{value},
			Expected: &AssertionValue{path},
			Errors: []error{
				errors.New("expected: value matches given json path"),
				err,
			},
		})
		return newValue(chain, nil)
	}

	return newValue(chain, result)
}

func jsonSchema(chain *chain, value, schema interface{}) {
	if chain.failed() {
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
		chain.fail(AssertionFailure{
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
		chain.fail(AssertionFailure{
			Type:     AssertMatchSchema,
			Actual:   &AssertionValue{value},
			Expected: &AssertionValue{schemaData},
			Errors:   errors,
		})
	}
}
