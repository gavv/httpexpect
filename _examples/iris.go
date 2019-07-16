package examples

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"github.com/kataras/iris/middleware/basicauth"
	"github.com/kataras/iris/sessions"
)

// IrisHandler tests iris handler
func IrisHandler() http.Handler {
	app := iris.New()

	sess := sessions.New(sessions.Config{
		Cookie: "irissessionid",
	})

	app.Get("/things", func(ctx context.Context) {
		ctx.JSON([]interface{}{
			context.Map{
				"name":        "foo",
				"description": "foo thing",
			},
			context.Map{
				"name":        "bar",
				"description": "bar thing",
			},
		})
	})

	app.Post("/redirect", func(ctx context.Context) {
		ctx.Redirect("/things", iris.StatusFound)
	})

	app.Post("/params/{x}/{y}", func(ctx context.Context) {
		ctx.JSON(context.Map{
			"x":  ctx.Params().Get("x"),
			"y":  ctx.Params().Get("y"),
			"q":  ctx.URLParam("q"),
			"p1": ctx.FormValue("p1"),
			"p2": ctx.FormValue("p2"),
		})
	})

	auth := basicauth.Default(map[string]string{
		"ford": "betelgeuse7",
	})

	app.Get("/auth", auth, func(ctx context.Context) {
		ctx.Writef("authenticated!")
	})

	app.Post("/session/set", func(ctx context.Context) {
		session := sess.Start(ctx)

		v := context.Map{}

		if err := ctx.ReadJSON(&v); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			return
		}

		session.Set("name", v["name"])
	})

	app.Get("/session/get", func(ctx context.Context) {
		session := sess.Start(ctx)

		ctx.JSON(context.Map{
			"name": session.GetString("name"),
		})
	})

	app.Get("/stream", func(ctx context.Context) {
		ctx.StreamWriter(func(w io.Writer) bool {
			for i := 0; i < 10; i++ {
				fmt.Fprintf(w, "%d", i)
			}
			// return true to continue, return false to stop and flush
			return false
		})
		// if we had to write here then the StreamWriter callback should
		// return true
	})

	app.Post("/stream", func(ctx context.Context) {
		body, err := ioutil.ReadAll(ctx.Request().Body)
		if err != nil {
			app.Logger().Error(err)
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.StopExecution()
			return
		}
		ctx.Write(body)
	})

	sub := app.Subdomain("subdomain")

	sub.Post("/set", func(ctx context.Context) {
		session := sess.Start(ctx)
		session.Set("message", "hello from subdomain")
	})

	sub.Get("/get", func(ctx context.Context) {
		session := sess.Start(ctx)
		ctx.WriteString(session.GetString("message"))
	})

	if err := app.Build(); err != nil {
		app.Logger().Error(err)
	}

	return app
}
