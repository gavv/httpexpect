package e2e

import (
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/valyala/fasthttp"
)

func TestE2EFs_FastBinder(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "httpexpect")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempdir)

	if err := ioutil.WriteFile(
		path.Join(tempdir, "hello"), []byte("hello, world!"), 0666); err != nil {
		t.Fatal(err)
	}

	fs := &fasthttp.FS{
		Root: tempdir,
	}

	handler := fs.NewRequestHandler()

	e := httpexpect.WithConfig(httpexpect.Config{
		Client: &http.Client{
			Transport: httpexpect.NewFastBinder(handler),
			Jar:       httpexpect.NewCookieJar(),
		},
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
	})

	e.GET("/hello").
		Expect().
		Status(http.StatusOK).
		Text().IsEqual("hello, world!")
}
