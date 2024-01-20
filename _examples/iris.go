package examples

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/basicauth"
	"github.com/kataras/iris/v12/sessions"
)

// IrisHandler tests iris handler
func IrisHandler() http.Handler {
	app := iris.New()

	sess := sessions.New(sessions.Config{
		Cookie: "irissessionid",
	})

	app.Get("/things", func(ctx iris.Context) {
		ctx.JSON([]interface{}{
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

	app.Post("/redirect", func(ctx iris.Context) {
		ctx.Redirect("/things", iris.StatusFound)
	})

	app.Post("/params/{x}/{y}", func(ctx iris.Context) {
		ctx.JSON(iris.Map{
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

	app.Get("/auth", auth, func(ctx iris.Context) {
		ctx.Writef("authenticated!")
	})

	app.Post("/session/set", func(ctx iris.Context) {
		session := sess.Start(ctx)

		v := iris.Map{}

		if err := ctx.ReadJSON(&v); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			return
		}

		session.Set("name", v["name"])
	})

	app.Get("/session/get", func(ctx iris.Context) {
		session := sess.Start(ctx)

		ctx.JSON(iris.Map{
			"name": session.GetString("name"),
		})
	})

	app.Get("/stream", func(ctx iris.Context) {
		i := 0
		ctx.StreamWriter(func(w io.Writer) error {
			if i == 10 {
				// causes StreamWriter to stop writing
				return io.EOF
			} else if _, err := fmt.Fprintf(w, "%d", i); err != nil {
				return err
			}

			i++
			// return nil to continue writing
			return nil
		})
		// if we had to write here then the StreamWriter callback should
		// return nil instead of EOF
	})

	app.Post("/stream", func(ctx iris.Context) {
		body, err := io.ReadAll(ctx.Request().Body)
		if err != nil {
			app.Logger().Error(err)
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.StopExecution()
			return
		}
		ctx.Write(body)
	})

	sub := app.Subdomain("subdomain")

	sub.Post("/set", func(ctx iris.Context) {
		session := sess.Start(ctx)
		session.Set("message", "hello from subdomain")
	})

	sub.Get("/get", func(ctx iris.Context) {
		session := sess.Start(ctx)
		ctx.WriteString(session.GetString("message"))
	})

	if err := app.Build(); err != nil {
		app.Logger().Error(err)
	}

	return app
}
