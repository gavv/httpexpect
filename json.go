package httpexpect

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
		chain.fail(&AssertionFailure{
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
		chain.fail(&AssertionFailure{
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

	valueLoader := gojsonschema.NewGoLoader(value)

	var schemaLoader gojsonschema.JSONLoader

	if str, ok := canonString(schema); ok {
		if ok, _ := regexp.MatchString(`^\w+://`, str); ok {
			schemaLoader = gojsonschema.NewReferenceLoader(str)
		} else {
			schemaLoader = gojsonschema.NewStringLoader(str)
		}
	} else {
		schemaLoader = gojsonschema.NewGoLoader(schema)
	}

	result, err := gojsonschema.Validate(schemaLoader, valueLoader)
	if err != nil {
		chain.fail(&AssertionFailure{
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
		if s, ok := schema.(string); ok {
			var buf bytes.Buffer
			if err := json.Indent(&buf, []byte(s), "", "  "); err == nil {
				schema = buf.String()
			}
		} else if b, ok := schema.([]byte); ok {
			var buf bytes.Buffer
			if err := json.Indent(&buf, b, "", "  "); err == nil {
				schema = buf.String()
			}
		}
		chain.fail(&AssertionFailure{
			Type:     AssertMatchSchema,
			Actual:   &AssertionValue{value},
			Expected: &AssertionValue{schema},
			Errors:   errors,
		})
		return
	}
}
