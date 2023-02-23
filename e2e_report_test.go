package httpexpect

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
)

type recordingReporter struct {
	reported string
}

func (r *recordingReporter) Errorf(msg string, args ...interface{}) {
	r.reported += fmt.Sprintf(msg, args...)
}

func TestE2EReport_Names(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	rep := &recordingReporter{}

	e := WithConfig(Config{
		TestName: "TestExample",
		BaseURL:  server.URL,
		Reporter: rep,
	})

	e.GET("/test").
		WithName("RequestExample").
		Expect().
		JSON() // will fail

	t.Logf("%s", rep.reported)

	assert.Contains(t, rep.reported, "TestExample")
	assert.Contains(t, rep.reported, "RequestExample")
}

func TestE2EReport_Aliases(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"foo":123}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	rep := &recordingReporter{}

	e := WithConfig(Config{
		TestName: "TestExample",
		BaseURL:  server.URL,
		Reporter: rep,
	})

	foo := e.GET("/test").
		WithName("RequestExample").
		Expect().
		JSON().Alias("foo")

	foo.Object().ContainsKey("bar") // will fail

	t.Logf("%s", rep.reported)

	assert.Contains(t, rep.reported, "foo.Object().ContainsKey()")
}

const (
	intSize = 32 << (^uint(0) >> 63) // 32 or 64
	maxInt  = 1<<(intSize-1) - 1
)

func TestE2EReport_LineWidth(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"foo":123}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	splitWhiteSpaceOrDot := func(r rune) bool {
		return unicode.IsSpace(r) || r == '.'
	}

	tests := []struct {
		name             string
		formatter        *DefaultFormatter
		wantMaxLineWidth int
	}{
		{
			name: "zero value",
			formatter: &DefaultFormatter{
				LineWidth: 0,
			},
			wantMaxLineWidth: defaultLineWidth,
		},
		{
			name: "small positive value",
			formatter: &DefaultFormatter{
				LineWidth: 15,
			},
			wantMaxLineWidth: 15,
		},
		{
			name: "very large positive value",
			formatter: &DefaultFormatter{
				LineWidth: 1000,
			},
			wantMaxLineWidth: 1000,
		},
		{
			name: "negative value (no limit)",
			formatter: &DefaultFormatter{
				LineWidth: -1,
			},
			wantMaxLineWidth: maxInt,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep := &recordingReporter{}

			e := WithConfig(Config{
				TestName: "TestExample",
				BaseURL:  server.URL,
				AssertionHandler: &DefaultAssertionHandler{
					Formatter: tt.formatter,
					Reporter:  rep,
				},
			})

			e.GET("/test").
				Expect().
				JSON().
				Object().
				IsEqual(map[string]interface{}{"foo": 1234})

			var skip bool
			for _, s := range strings.Split(strings.Trim(rep.reported, "\n"), "\n") {
				// only assert first errors block and assertPath block in defaultFailureTemplate
				if strings.HasPrefix(s, "assertion:") {
					skip = false
				}
				if len(s) == 0 {
					skip = true
				}
				if skip {
					continue
				}

				s = strings.Trim(s, ".")

				ss := strings.FieldsFunc(s, splitWhiteSpaceOrDot)

				if len(ss) <= 1 {
					continue
				}

				lenLastWord := len(ss[len(ss)-1])

				// additional indent in template
				var lenIndent int
				if strings.HasPrefix(s, defaultIndent) {
					lenIndent = len(defaultIndent)
				}

				lenBeforeWrapped := len(s) - lenLastWord - lenIndent - 1

				t.Logf("%s", s)

				assert.LessOrEqual(t, lenBeforeWrapped, tt.wantMaxLineWidth)
			}
		})
	}
}
