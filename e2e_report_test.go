package httpexpect

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type recordingReporter struct {
	reported string
}

func (r *recordingReporter) Errorf(msg string, args ...interface{}) {
	r.reported += fmt.Sprintf(msg, args...)
}

func TestE2EReportNames(t *testing.T) {
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
