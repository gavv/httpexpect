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
		Status(http.StatusOK).JSON().Array().Empty()

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
		Status(http.StatusOK).JSON().Object().Equal(orange).NotEqual(apple)

	e.GET("/fruits/orange").
		Expect().
		Status(http.StatusOK).
		JSON().Object().ContainsKey("weight").ValueEqual("weight", 100)

	obj := e.GET("/fruits/apple").
		Expect().
		Status(http.StatusOK).JSON().Object()

	obj.Keys().ContainsOnly("colors", "weight", "image")

	obj.Value("colors").Array().Elements("green", "red")
	obj.Value("colors").Array().Element(0).String().Equal("green")
	obj.Value("colors").Array().Element(1).String().Equal("red")
	obj.Value("colors").Array().Element(1).String().IsASCII()
	obj.Value("colors").Array().Element(1).String().HasPrefix("re")
	obj.Value("colors").Array().Element(1).String().HasSuffix("ed")

	obj.Value("weight").Number().Equal(200)

	for _, element := range obj.Value("image").Array().Iter() {
		element.Object().ContainsKey("id")
		element.Object().ContainsValue("fruit")
		element.Object().ContainsSubset(map[string]interface{}{
			"type": "fruit",
		})
	}

	e.GET("/fruits/melon").
		Expect().
		Status(http.StatusNotFound)
}
