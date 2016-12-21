package examples

import (
	"net/http"
	"os"
	"testing"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"

	"github.com/gavv/httpexpect"
)

// init is used by GAE to start serving the app
// added here for illustration purposes
// func init() {
// 	http.Handle("/", Router())
// }

func Router() http.Handler {
	m := http.NewServeMux()

	m.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_ = appengine.NewContext(r)
		w.Write([]byte("pong"))
	})

	return m
}

// GaeInstance is our global dev_appserver instance
var GaeInstance aetest.Instance

// TestMain is called first to create the GaeInstance
func TestMain(m *testing.M) {

	// INFO: Remove the return to actually run the tests.
	// Requires installed Google Appengine SDK.
	// https://cloud.google.com/appengine/downloads
	return

	var err error
	GaeInstance, err = aetest.NewInstance(nil)
	if err != nil {
		panic(err)
	}

	c := m.Run() // call all actual tests
	GaeInstance.Close()
	os.Exit(c)
}

// newHttpExpect returns a new Expect instance for testing
func newHttpExpect(t *testing.T) *httpexpect.Expect {
	return httpexpect.WithConfig(httpexpect.Config{
		Client: &http.Client{
			Transport: httpexpect.NewBinder(Router()),
			Jar:       httpexpect.NewJar(),
		},
		Reporter:       httpexpect.NewAssertReporter(t),
		RequestFactory: GaeInstance,
	})
}

// TestPing is an actual tests, using the global GaeInstance
func TestPing(t *testing.T) {
	e := newHttpExpect(t)
	e.GET("/ping").Expect().Status(200)
}
