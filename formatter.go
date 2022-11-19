package httpexpect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

// Formatter is used to format assertion messages into strings.
type Formatter interface {
	FormatSuccess(*AssertionContext) string
	FormatFailure(*AssertionContext, *AssertionFailure) string
}

// DefaultFormatter is the default Formatter implementation.
//
// DefaultFormatter gathers values from AssertionContext and AssertionFailure,
// converts them to strings, and creates FormatData struct. Then it passes
// FormatData to the template engine (text/template) to format message.
//
// You can control what is included and what is excluded from messages via
// several public fields.
//
// If desired, you can provide custom templates and function map. This may
// be easier than creating your own formatter from scratch.
type DefaultFormatter struct {
	// Exclude test name from failure report.
	DisableName bool

	// Exclude assertion path from failure report.
	DisablePath bool

	// Exclude diff from failure report.
	DisableDiff bool

	// If not empty, used to format success messages.
	// If empty, default template is used.
	SuccessTemplate string

	// If not empty, used to format failure messages.
	// If empty, default template is used.
	FailureTemplate string

	// When SuccessTemplate or FailureTemplate is set, this field
	// defines the function map passed to template engine.
	// May be nil.
	TemplateFuncs template.FuncMap
}

// FormatSuccess implements Formatter.FormatSuccess.
func (f *DefaultFormatter) FormatSuccess(ctx *AssertionContext) string {
	if f.SuccessTemplate != "" {
		return f.formatTemplate("SuccessTemplate",
			f.SuccessTemplate, f.TemplateFuncs, ctx, nil)
	} else {
		return f.formatTemplate("SuccessTemplate",
			defaultSuccessTemplate, defaultTemplateFuncs, ctx, nil)
	}
}

// FormatFailure implements Formatter.FormatFailure.
func (f *DefaultFormatter) FormatFailure(
	ctx *AssertionContext, failure *AssertionFailure,
) string {
	if f.FailureTemplate != "" {
		return f.formatTemplate("FailureTemplate",
			f.FailureTemplate, f.TemplateFuncs, ctx, failure)
	} else {
		return f.formatTemplate("FailureTemplate",
			defaultFailureTemplate, defaultTemplateFuncs, ctx, failure)
	}
}

// FormatData defines data passed to template engine.
type FormatData struct {
	TestName string

	AssertPath string
	AssertType string

	Errors []string

	HaveActual bool
	Actual     string

	HaveExpected bool
	IsUnexpected bool
	IsComparison bool
	ExpectedKind string
	Expected     []string

	HaveDelta bool
	Delta     string

	HaveDiff bool
	Diff     string
}

const (
	kindRange     = "range"
	kindSchema    = "schema"
	kindPath      = "path"
	kindRegexp    = "regexp"
	kindKey       = "key"
	kindElement   = "element"
	kindSubset    = "subset"
	kindValue     = "value"
	kindValueList = "values"
)

func (f *DefaultFormatter) formatTemplate(
	templateName string,
	templateString string,
	templateFuncs template.FuncMap,
	ctx *AssertionContext,
	failure *AssertionFailure,
) string {
	templateData := f.buildFormatData(ctx, failure)

	t, err := template.New(templateName).Funcs(templateFuncs).Parse(templateString)
	if err != nil {
		panic(err)
	}

	var b bytes.Buffer

	err = t.Execute(&b, templateData)
	if err != nil {
		panic(err)
	}

	return b.String()
}

func (f *DefaultFormatter) buildFormatData(
	ctx *AssertionContext, failure *AssertionFailure,
) *FormatData {
	data := FormatData{}

	f.fillDescription(&data, ctx)

	if failure != nil {
		data.AssertType = failure.Type.String()

		f.fillErrors(&data, ctx, failure)

		if failure.Actual != nil {
			f.fillActual(&data, ctx, failure)
		}

		if failure.Expected != nil {
			f.fillExpected(&data, ctx, failure)
			f.fillIsUnexpected(&data, ctx, failure)
			f.fillIsComparison(&data, ctx, failure)
		}

		if failure.Delta != 0 {
			f.fillDelta(&data, ctx, failure)
		}
	}

	return &data
}

func (f *DefaultFormatter) fillDescription(
	data *FormatData, ctx *AssertionContext,
) {
	if !f.DisableName {
		data.TestName = ctx.TestName
	}

	if !f.DisablePath {
		data.AssertPath = formatPath(ctx.Path)
	}
}

func (f *DefaultFormatter) fillErrors(
	data *FormatData, ctx *AssertionContext, failure *AssertionFailure,
) {
	for _, err := range failure.Errors {
		if err == nil {
			continue
		}
		data.Errors = append(data.Errors, err.Error())
	}
}

func (f *DefaultFormatter) fillActual(
	data *FormatData, ctx *AssertionContext, failure *AssertionFailure,
) {
	switch failure.Type { //nolint
	case AssertUsage, AssertOperation:
		data.HaveActual = false

	default:
		data.HaveActual = true
		data.Actual = formatValue(failure.Actual.Value)
	}
}

func (f *DefaultFormatter) fillExpected(
	data *FormatData, ctx *AssertionContext, failure *AssertionFailure,
) {
	switch failure.Type {
	case AssertUsage, AssertOperation,
		AssertValid, AssertNotValid,
		AssertNil, AssertNotNil,
		AssertEmpty, AssertNotEmpty:
		data.HaveExpected = false

	case AssertNotEqual:
		data.HaveExpected = false

	case AssertEqual, AssertLt, AssertLe, AssertGt, AssertGe:
		data.HaveExpected = true
		data.ExpectedKind = kindValue
		data.Expected = []string{
			formatValue(failure.Expected.Value),
		}

		if !f.DisableDiff && failure.Actual != nil && failure.Expected != nil {
			data.Diff, data.HaveDiff = formatDiff(
				failure.Expected.Value, failure.Actual.Value)
		}

	case AssertInRange, AssertNotInRange:
		data.HaveExpected = true
		data.ExpectedKind = kindRange
		data.Expected = []string{
			formatRange(failure.Expected.Value),
		}

	case AssertMatchSchema, AssertNotMatchSchema:
		data.HaveExpected = true
		data.ExpectedKind = kindSchema
		data.Expected = []string{
			formatString(failure.Expected.Value),
		}

	case AssertMatchPath, AssertNotMatchPath:
		data.HaveExpected = true
		data.ExpectedKind = kindPath
		data.Expected = []string{
			formatString(failure.Expected.Value),
		}

	case AssertMatchRegexp, AssertNotMatchRegexp:
		data.HaveExpected = true
		data.ExpectedKind = kindRegexp
		data.Expected = []string{
			formatString(failure.Expected.Value),
		}

	case AssertContainsKey, AssertNotContainsKey:
		data.HaveExpected = true
		data.ExpectedKind = kindKey
		data.Expected = []string{
			formatValue(failure.Expected.Value),
		}

	case AssertContainsElement, AssertNotContainsElement:
		data.HaveExpected = true
		data.ExpectedKind = kindElement
		data.Expected = []string{
			formatValue(failure.Expected.Value),
		}

	case AssertContainsSubset, AssertNotContainsSubset:
		data.HaveExpected = true
		data.ExpectedKind = kindSubset
		data.Expected = []string{
			formatValue(failure.Expected.Value),
		}

	case AssertBelongs, AssertNotBelongs:
		data.HaveExpected = true
		data.ExpectedKind = kindValueList
		data.Expected = formatList(failure.Expected.Value)
	}
}

func (f *DefaultFormatter) fillIsUnexpected(
	data *FormatData, ctx *AssertionContext, failure *AssertionFailure,
) {
	switch failure.Type {
	case AssertUsage, AssertOperation,
		AssertValid,
		AssertNil,
		AssertEmpty,
		AssertEqual,
		AssertLt, AssertLe, AssertGt, AssertGe,
		AssertInRange,
		AssertMatchSchema,
		AssertMatchPath,
		AssertMatchRegexp,
		AssertContainsKey,
		AssertContainsElement,
		AssertContainsSubset,
		AssertBelongs:
		break

	case AssertNotValid,
		AssertNotNil,
		AssertNotEmpty,
		AssertNotEqual,
		AssertNotInRange,
		AssertNotMatchSchema,
		AssertNotMatchPath,
		AssertNotMatchRegexp,
		AssertNotContainsKey,
		AssertNotContainsElement,
		AssertNotContainsSubset,
		AssertNotBelongs:
		data.IsUnexpected = true
	}
}

func (f *DefaultFormatter) fillIsComparison(
	data *FormatData, ctx *AssertionContext, failure *AssertionFailure,
) {
	switch failure.Type { //nolint
	case AssertLt, AssertLe, AssertGt, AssertGe:
		data.IsComparison = true
	}
}

func (f *DefaultFormatter) fillDelta(
	data *FormatData, ctx *AssertionContext, failure *AssertionFailure,
) {
	data.HaveDelta = true
	data.Delta = fmt.Sprintf("%f", failure.Delta)
}

func formatPath(path []string) string {
	return strings.Join(path, ".")
}

func formatValue(v interface{}) string {
	isNil := func(a interface{}) bool {
		defer func() {
			_ = recover()
		}()
		return a == nil || reflect.ValueOf(a).IsNil()
	}
	if !isNil(v) {
		if s, _ := v.(fmt.Stringer); s != nil {
			return s.String()
		}
		if b, err := json.MarshalIndent(v, "", "  "); err == nil {
			return string(b)
		}
	}
	return fmt.Sprintf("%#v", v)
}

func formatString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	} else {
		return formatValue(v)
	}
}

func formatRange(v interface{}) string {
	if r, ok := v.(AssertionRange); ok {
		return fmt.Sprintf("[%v; %v]", r[0], r[1])
	} else {
		return formatValue(v)
	}
}

func formatList(v interface{}) []string {
	if l, ok := v.(AssertionList); ok {
		s := make([]string, 0, len(l))
		for _, e := range l {
			s = append(s, formatValue(e))
		}
		return s
	} else {
		return []string{
			formatValue(v),
		}
	}
}

func formatDiff(expected, actual interface{}) (string, bool) {
	differ := gojsondiff.New()

	var diff gojsondiff.Diff

	if ve, ok := expected.(map[string]interface{}); ok {
		if va, ok := actual.(map[string]interface{}); ok {
			diff = differ.CompareObjects(ve, va)
		} else {
			return "", false
		}
	} else if ve, ok := expected.([]interface{}); ok {
		if va, ok := actual.([]interface{}); ok {
			diff = differ.CompareArrays(ve, va)
		} else {
			return "", false
		}
	} else {
		return "", false
	}

	if !diff.Modified() {
		return "", false
	}

	config := formatter.AsciiFormatterConfig{
		ShowArrayIndex: true,
	}
	f := formatter.NewAsciiFormatter(expected, config)

	str, err := f.Format(diff)
	if err != nil {
		return "", false
	}

	diffText := "--- expected\n+++ actual\n" + str

	return diffText, true
}

var defaultTemplateFuncs = template.FuncMap{
	"strip": func(in string) string {
		return strings.TrimSpace(in)
	},
	"indent": func(in string) string {
		out := ""
		for _, s := range strings.Split(in, "\n") {
			if out != "" {
				out += "\n"
			}
			out += "  " + s
		}
		return out
	},
}

var defaultSuccessTemplate = `[OK] {{ .AssertPath }}`

var defaultFailureTemplate = `
{{- range $n, $err := .Errors }}
{{ if eq $n 0 -}}
{{ $err | strip }}
{{- else -}}
{{ $err | indent }}
{{- end -}}
{{- end -}}
{{- if .TestName }}

test: {{ .TestName }}
{{- end -}}
{{- if .AssertPath }}

assertion:
{{ .AssertPath | indent }}
{{- end -}}
{{- if .HaveExpected }}

{{ if .IsUnexpected }}denied
{{- else if .IsComparison }}compared
{{- else }}expected
{{- end }} {{ .ExpectedKind }}:
{{- range $n, $exp := .Expected }}
{{ $exp | indent }}
{{- end -}}
{{- end -}}
{{- if .HaveActual }}

actual value:
{{ .Actual | indent }}
{{- end -}}
{{- if .HaveDelta }}

allowed delta:
{{ .Delta | indent }}
{{- end -}}
{{- if .HaveDiff }}

diff:
{{ .Diff | indent }}
{{- end -}}
`
