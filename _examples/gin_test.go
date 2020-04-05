package examples

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/gin-gonic/gin"
)

func TestGinHandler(t *testing.T) {
	// Create new gin instance
	engine := gin.New()
	// Add /example route via handler function to the gin instance
	handler := GinHandler(engine)
	// Create httpexpect instance
	e := httpexpect.WithConfig(httpexpect.Config{
		Client: &http.Client{
			Transport: httpexpect.NewBinder(handler),
			Jar:       httpexpect.NewJar(),
		},
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
	})

	// Assert response
	e.GET("/example").
		Expect().
		Status(http.StatusOK).JSON().Object().ValueEqual("message", "pong")
}
