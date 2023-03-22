package examples

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gavv/httpexpect/v2"
)

func TestFruits(t *testing.T) {
	handler := FruitsHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	e := httpexpect.Default(t, server.URL)

	e.GET("/fruits").
		Expect().
		Status(http.StatusOK).JSON().Array().IsEmpty()

	orange := map[string]interface{}{
		"weight": 100,
	}

	e.PUT("/fruits/orange").WithJSON(orange).
		Expect().
		Status(http.StatusNoContent).NoContent()

	apple := map[string]interface{}{
		"colors": []interface{}{"green", "red"},
		"weight": 200,
		"image": []map[string]string{
			{
				"id":   " 1",
				"url":  "http://example.com",
				"type": "fruit",
			},
			{
				"id":   "2",
				"url":  "http://example2.com",
				"type": "fruit",
			},
		},
	}

	e.PUT("/fruits/apple").WithJSON(apple).
		Expect().
		Status(http.StatusNoContent).NoContent()

	fruits := e.GET("/fruits").
		Expect().
		Status(http.StatusOK).JSON().Array()

	fruits.Every(func(index int, value *httpexpect.Value) {
		value.String().NotEmpty()
	})
	fruits.ContainsAny("orange", "melon")
	fruits.ContainsOnly("orange", "apple")

	e.GET("/fruits/orange").
		Expect().
		Status(http.StatusOK).JSON().Object().IsEqual(orange).NotEqual(apple)

	e.GET("/fruits/orange").
		Expect().
		Status(http.StatusOK).
		JSON().Object().ContainsKey("weight").HasValue("weight", 100)

	fruit := e.GET("/fruits/apple").
		Expect().
		Status(http.StatusOK).JSON().Object()

	fruit.Keys().ContainsOnly("colors", "weight", "image")

	colors := fruit.Value("colors").Array()
	colors.Alias("colors")

	colors.ConsistsOf("green", "red")

	colors.Length().IsEqual(2)
	colors.Value(0).String().IsEqual("green")
	colors.Value(1).String().IsEqual("red")

	colors.Value(0).String().IsASCII()
	colors.Value(0).String().HasPrefix("gr")
	colors.Value(0).String().HasSuffix("een")

	fruit.Value("weight").Number().InList(100, 200, 300)
	fruit.Value("weight").Number().NotInList(400, 500, 600)

	for _, element := range fruit.Value("image").Array().Iter() {
		element.Object().
			ContainsKey("type").
			ContainsValue("fruit").
			ContainsSubset(map[string]interface{}{
				"type": "fruit",
			})
	}

	e.GET("/fruits/melon").
		Expect().
		Status(http.StatusNotFound)
}
