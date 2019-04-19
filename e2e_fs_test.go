package httpexpect

import (
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/valyala/fasthttp"
)

func TestE2EFsFastBinder(t *testing.T) {
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

	e := WithConfig(Config{
		Client: &http.Client{
			Transport: NewFastBinder(handler),
			Jar:       NewJar(),
		},
		Reporter: NewAssertReporter(t),
		Printers: []Printer{
			NewDebugPrinter(t, true),
		},
	})

	e.GET("/hello").
		Expect().
		Status(http.StatusOK).
		Text().Equal("hello, world!")
}
