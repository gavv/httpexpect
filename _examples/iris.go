package examples

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"gopkg.in/kataras/iris.v6"
	"gopkg.in/kataras/iris.v6/adaptors/httprouter"
	"gopkg.in/kataras/iris.v6/adaptors/sessions"
	"gopkg.in/kataras/iris.v6/middleware/basicauth"
)

// IrisHandler tests iris v6's handler
func IrisHandler() http.Handler {
	app := iris.New(iris.Configuration{VHost: "example.com"})
	app.Adapt(iris.DevLogger())
	app.Adapt(httprouter.New())
	app.Adapt(sessions.New(sessions.Config{Cookie: "irissessionid"}))

	app.Get("/things", func(c *iris.Context) {
		c.JSON(iris.StatusOK, []interface{}{
			iris.Map{
				"name":        "foo",
				"description": "foo thing",
			},
			iris.Map{
				"name":        "bar",
				"description": "bar thing",
			},
		})
	})

	app.Post("/redirect", func(c *iris.Context) {
		c.Redirect("/things", iris.StatusFound)
	})

	app.Post("/params/:x/:y", func(c *iris.Context) {
		c.JSON(iris.StatusOK, iris.Map{
			"x":  c.Param("x"),
			"y":  c.Param("y"),
			"q":  c.URLParam("q"),
			"p1": c.FormValue("p1"),
			"p2": c.FormValue("p2"),
		})
	})

	auth := basicauth.Default(map[string]string{
		"ford": "betelgeuse7",
	})

	app.Get("/auth", auth, func(c *iris.Context) {
		c.Writef("authenticated!")
	})

	app.Post("/session/set", func(c *iris.Context) {
		sess := iris.Map{}

		if err := c.ReadJSON(&sess); err != nil {
			panic(err.Error())
		}

		c.Session().Set("name", sess["name"])
	})

	app.Get("/session/get", func(c *iris.Context) {
		name := c.Session().GetString("name")

		c.JSON(iris.StatusOK, iris.Map{
			"name": name,
		})
	})

	app.Get("/stream", func(c *iris.Context) {
		c.StreamWriter(func(w io.Writer) bool {
			for i := 0; i < 10; i++ {
				fmt.Fprintf(w, "%d", i)
			}
			// return true to continue, return false to stop and flush
			return false
		})
		// if we had to write here then the StreamWriter callback should
		// return true
	})

	app.Post("/stream", func(c *iris.Context) {
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			c.EmitError(iris.StatusBadRequest)
			return
		}
		c.Write(body)
	})

	sub := app.Party("subdomain.")

	sub.Post("/set", func(c *iris.Context) {
		c.Session().Set("message", "hello from subdomain")
	})

	sub.Get("/get", func(c *iris.Context) {
		c.Writef(c.Session().GetString("message"))
	})

	app.Boot()
	return app.Router
}
