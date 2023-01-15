package httpexpect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"text/template"

	"github.com/mitchellh/go-wordwrap"
	"github.com/sanity-io/litter"
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
	// Exclude test name and request name from failure report.
	DisableNames bool

	// Exclude assertion path from failure report.
	DisablePaths bool

	// Exclude aliased assertion path from failure report.
	DisableAliases bool

	// Exclude diff from failure report.
	DisableDiffs bool

	// Wrap text to keep lines below given width.
	// Use zero for default width, and negative value to disable wrapping.
	LineWidth int

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
		return f.applyTemplate("SuccessTemplate",
			f.SuccessTemplate, f.TemplateFuncs, ctx, nil)
	} else {
		return f.applyTemplate("SuccessTemplate",
			defaultSuccessTemplate, defaultTemplateFuncs, ctx, nil)
	}
}

// FormatFailure implements Formatter.FormatFailure.
func (f *DefaultFormatter) FormatFailure(
	ctx *AssertionContext, failure *AssertionFailure,
) string {
	if f.FailureTemplate != "" {
		return f.applyTemplate("FailureTemplate",
			f.FailureTemplate, f.TemplateFuncs, ctx, failure)
	} else {
		return f.applyTemplate("FailureTemplate",
			defaultFailureTemplate, defaultTemplateFuncs, ctx, failure)
	}
}

// FormatData defines data passed to template engine when DefaultFormatter
// formats assertion. You can use these fields in your custom templates.
type FormatData struct {
	TestName    string
	RequestName string

	AssertPath     []string
	AssertType     string
	AssertSeverity string

	Errors []string

	HaveActual bool
	Actual     string

	HaveExpected bool
	IsNegation   bool
	IsComparison bool
	ExpectedKind string
	Expected     []string

	HaveReference bool
	Reference     string

	HaveDelta bool
	Delta     string

	HaveDiff bool
	Diff     string

	LineWidth int
}

const (
	kindRange      = "range"
	kindSchema     = "schema"
	kindPath       = "path"
	kindRegexp     = "regexp"
	kindFormat     = "format"
	kindFormatList = "formats"
	kindKey        = "key"
	kindElement    = "element"
	kindSubset     = "subset"
	kindValue      = "value"
	kindValueList  = "values"
)

func (f *DefaultFormatter) applyTemplate(
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
		data.AssertSeverity = failure.Severity.String()

		f.fillErrors(&data, ctx, failure)

		if failure.Actual != nil {
			f.fillActual(&data, ctx, failure)
		}

		if failure.Expected != nil {
			f.fillExpected(&data, ctx, failure)
			f.fillIsNegation(&data, ctx, failure)
			f.fillIsComparison(&data, ctx, failure)
		}

		if failure.Reference != nil {
			f.fillReference(&data, ctx, failure)
		}

		if failure.Delta != nil {
			f.fillDelta(&data, ctx, failure)
		}
	}

	return &data
}

func (f *DefaultFormatter) fillDescription(
	data *FormatData, ctx *AssertionContext,
) {
	if !f.DisableNames {
		data.TestName = ctx.TestName
		data.RequestName = ctx.RequestName
	}

	if !f.DisablePaths {
		if !f.DisableAliases {
			data.AssertPath = ctx.AliasedPath
		} else {
			data.AssertPath = ctx.Path
		}
	}

	if f.LineWidth != 0 {
		data.LineWidth = f.LineWidth
	} else {
		data.LineWidth = defaultLineWidth
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

	case AssertType, AssertNotType:
		data.HaveActual = true
		data.Actual = f.formatTyped(failure.Actual.Value)

	default:
		data.HaveActual = true
		data.Actual = f.formatValue(failure.Actual.Value)
	}
}

func (f *DefaultFormatter) fillExpected(
	data *FormatData, ctx *AssertionContext, failure *AssertionFailure,
) {
	switch failure.Type {
	case AssertUsage, AssertOperation,
		AssertType, AssertNotType,
		AssertValid, AssertNotValid,
		AssertNil, AssertNotNil,
		AssertEmpty, AssertNotEmpty,
		AssertNotEqual:
		data.HaveExpected = false

	case AssertEqual:
		data.HaveExpected = true
		data.ExpectedKind = kindValue
		data.Expected = []string{
			f.formatValue(failure.Expected.Value),
		}

		if !f.DisableDiffs && failure.Actual != nil && failure.Expected != nil {
			data.Diff, data.HaveDiff = f.formatDiff(
				failure.Expected.Value, failure.Actual.Value)
		}

	case AssertLt, AssertLe, AssertGt, AssertGe:
		data.HaveExpected = true
		data.ExpectedKind = kindValue
		data.Expected = []string{
			f.formatValue(failure.Expected.Value),
		}

	case AssertInRange, AssertNotInRange:
		data.HaveExpected = true
		data.ExpectedKind = kindRange
		data.Expected = f.formatRange(failure.Expected.Value)

	case AssertMatchSchema, AssertNotMatchSchema:
		data.HaveExpected = true
		data.ExpectedKind = kindSchema
		data.Expected = []string{
			f.formatBareString(failure.Expected.Value),
		}

	case AssertMatchPath, AssertNotMatchPath:
		data.HaveExpected = true
		data.ExpectedKind = kindPath
		data.Expected = []string{
			f.formatBareString(failure.Expected.Value),
		}

	case AssertMatchRegexp, AssertNotMatchRegexp:
		data.HaveExpected = true
		data.ExpectedKind = kindRegexp
		data.Expected = []string{
			f.formatBareString(failure.Expected.Value),
		}

	case AssertMatchFormat, AssertNotMatchFormat:
		data.HaveExpected = true
		if extractList(failure.Expected.Value) != nil {
			data.ExpectedKind = kindFormatList
		} else {
			data.ExpectedKind = kindFormat
		}
		data.Expected = f.formatList(failure.Expected.Value)

	case AssertContainsKey, AssertNotContainsKey:
		data.HaveExpected = true
		data.ExpectedKind = kindKey
		data.Expected = []string{
			f.formatValue(failure.Expected.Value),
		}

	case AssertContainsElement, AssertNotContainsElement:
		data.HaveExpected = true
		data.ExpectedKind = kindElement
		data.Expected = []string{
			f.formatValue(failure.Expected.Value),
		}

	case AssertContainsSubset, AssertNotContainsSubset:
		data.HaveExpected = true
		data.ExpectedKind = kindSubset
		data.Expected = []string{
			f.formatValue(failure.Expected.Value),
		}

	case AssertBelongs, AssertNotBelongs:
		data.HaveExpected = true
		data.ExpectedKind = kindValueList
		data.Expected = f.formatList(failure.Expected.Value)
	}
}

func (f *DefaultFormatter) fillIsNegation(
	data *FormatData, ctx *AssertionContext, failure *AssertionFailure,
) {
	switch failure.Type {
	case AssertUsage, AssertOperation,
		AssertType,
		AssertValid,
		AssertNil,
		AssertEmpty,
		AssertEqual,
		AssertLt, AssertLe, AssertGt, AssertGe,
		AssertInRange,
		AssertMatchSchema,
		AssertMatchPath,
		AssertMatchRegexp,
		AssertMatchFormat,
		AssertContainsKey,
		AssertContainsElement,
		AssertContainsSubset,
		AssertBelongs:
		break

	case AssertNotType,
		AssertNotValid,
		AssertNotNil,
		AssertNotEmpty,
		AssertNotEqual,
		AssertNotInRange,
		AssertNotMatchSchema,
		AssertNotMatchPath,
		AssertNotMatchRegexp,
		AssertNotMatchFormat,
		AssertNotContainsKey,
		AssertNotContainsElement,
		AssertNotContainsSubset,
		AssertNotBelongs:
		data.IsNegation = true
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

func (f *DefaultFormatter) fillReference(
	data *FormatData, ctx *AssertionContext, failure *AssertionFailure,
) {
	data.HaveReference = true
	data.Reference = f.formatValue(failure.Reference.Value)
}

func (f *DefaultFormatter) fillDelta(
	data *FormatData, ctx *AssertionContext, failure *AssertionFailure,
) {
	data.HaveDelta = true
	data.Delta = f.formatValue(failure.Delta.Value)
}

func (f *DefaultFormatter) formatTyped(value interface{}) string {
	if isNumber(value) {
		return fmt.Sprintf("%T(%v)", value, f.formatValue(value))
	}

	return fmt.Sprintf("%T(%#v)", value, value)
}

func (f *DefaultFormatter) formatValue(value interface{}) string {
	if isNumber(value) {
		return fmt.Sprintf("%v", value)
	}

	if !isNil(value) && !isHTTP(value) {
		if s, _ := value.(fmt.Stringer); s != nil {
			if ss := s.String(); strings.TrimSpace(ss) != "" {
				return ss
			}
		}
		if b, err := json.MarshalIndent(value, "", defaultIndent); err == nil {
			return string(b)
		}
	}

	sq := litter.Options{
		Separator: defaultIndent,
	}

	return sq.Sdump(value)
}

func (f *DefaultFormatter) formatBareString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	}

	return f.formatValue(value)
}

func (f *DefaultFormatter) formatRange(value interface{}) []string {
	if rng := exctractRange(value); rng != nil {
		if isNumber(rng.Min) && isNumber(rng.Max) {
			return []string{
				fmt.Sprintf("[%v; %v]", f.formatValue(rng.Min), f.formatValue(rng.Max)),
			}
		} else {
			return []string{
				fmt.Sprintf("%v", rng.Min),
				fmt.Sprintf("%v", rng.Max),
			}
		}
	} else {
		return []string{
			f.formatValue(value),
		}
	}
}

func (f *DefaultFormatter) formatList(value interface{}) []string {
	if lst := extractList(value); lst != nil {
		s := make([]string, 0, len(*lst))
		for _, e := range *lst {
			s = append(s, f.formatValue(e))
		}
		return s
	} else {
		return []string{
			f.formatValue(value),
		}
	}
}

func (f *DefaultFormatter) formatDiff(expected, actual interface{}) (string, bool) {
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
	fa := formatter.NewAsciiFormatter(expected, config)

	str, err := fa.Format(diff)
	if err != nil {
		return "", false
	}

	diffText := "--- expected\n+++ actual\n" + str

	return diffText, true
}

func exctractRange(value interface{}) *AssertionRange {
	switch rng := value.(type) {
	case AssertionRange:
		return &rng
	case *AssertionRange: // invalid, but we handle it
		return rng
	default:
		return nil
	}
}

func extractList(value interface{}) *AssertionList {
	switch lst := value.(type) {
	case AssertionList:
		return &lst
	case *AssertionList: // invalid, but we handle it
		return lst
	default:
		return nil
	}
}

func isNil(value interface{}) bool {
	defer func() {
		_ = recover()
	}()
	return value == nil || reflect.ValueOf(value).IsNil()
}

func isNumber(value interface{}) bool {
	defer func() {
		_ = recover()
	}()
	reflect.ValueOf(value).Convert(reflect.TypeOf(float64(0))).Float()
	return true
}

func isHTTP(value interface{}) bool {
	switch value.(type) {
	case *http.Client, http.Client,
		*http.Transport, http.Transport,
		*http.Request, http.Request,
		*http.Response, http.Response,
		*http.Header, http.Header,
		*http.Cookie, http.Cookie:
		return true
	default:
		return false
	}
}

const (
	defaultIndent    = "  "
	defaultLineWidth = 60
)

var defaultTemplateFuncs = template.FuncMap{
	"indent": func(s string) string {
		var sb strings.Builder

		for _, s := range strings.Split(s, "\n") {
			if sb.Len() != 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(defaultIndent)
			sb.WriteString(s)
		}

		return sb.String()
	},
	"wrap": func(s string, width int) string {
		s = strings.TrimSpace(s)
		if width < 0 {
			return s
		}

		return wordwrap.WrapString(s, uint(width))
	},
	"join": func(strs []string, width int) string {
		if width < 0 {
			return strings.Join(strs, ".")
		}

		var sb strings.Builder

		lineLen := 0
		lineNum := 0

		write := func(s string) {
			sb.WriteString(s)
			lineLen += len(s)
		}

		for n, s := range strs {
			if lineLen > width {
				write("\n")
				lineLen = 0
				lineNum++
			}
			if lineLen == 0 {
				for l := 0; l < lineNum; l++ {
					write(defaultIndent)
				}
			}
			write(s)
			if n != len(strs)-1 {
				write(".")
			}
		}

		return sb.String()
	},
}

var defaultSuccessTemplate = `[OK] {{ join .AssertPath .LineWidth }}`

var defaultFailureTemplate = `
{{- range $n, $err := .Errors }}
{{ if eq $n 0 -}}
{{ wrap $err $.LineWidth }}
{{- else -}}
{{ wrap $err $.LineWidth | indent }}
{{- end -}}
{{- end -}}
{{- if .TestName }}

test name: {{ .TestName }}
{{- end -}}
{{- if .RequestName }}

request name: {{ .RequestName }}
{{- end -}}
{{- if .AssertPath }}

assertion:
{{ join .AssertPath .LineWidth | indent }}
{{- end -}}
{{- if .HaveExpected }}

{{ if .IsNegation }}denied
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
{{- if .HaveReference }}

reference value:
{{ .Reference | indent }}
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
