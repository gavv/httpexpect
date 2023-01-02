package examples

import (
	"fmt"
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
	}

	
	e.PUT("/fruits/apple").WithJSON(apple).
		Expect().
		Status(http.StatusNoContent).NoContent()

	fruitsResult := e.GET("/fruits").
		Expect().
		Status(http.StatusOK).JSON().Array()
	
	fruitsResult.ContainsOnly("orange", "apple")
	fruitsResult.ContainsAny("orange","melon")
	fruitsResult.Every(func(index int, value *httpexpect.Value) {
		value.String().NotEmpty()
	})

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

	obj.Keys().ContainsOnly("colors", "weight")

	obj.Value("colors").Array().Elements("green", "red")
	obj.Value("colors").Array().Element(0).String().Equal("green")
	obj.Value("colors").Array().Element(1).String().Equal("red")
	obj.Value("colors").Array().Element(1).String().IsASCII()
	obj.Value("colors").Array().Element(1).String().HasPrefix("re") 
	obj.Value("colors").Array().Element(1).String().HasSuffix("ed")

	obj.Value("weight").Number().Equal(200)

	e.GET("/fruits/melon").
		Expect().
		Status(http.StatusNotFound)

	var fruits fruitMapList
	apple["type"] = "fruit"
	orange["type"] = "fruit"
	fruits = append(fruits, apple)
	fruits = append(fruits, orange)

	fruitsData := e.GET("/fruits/with-data").WithJSON(fruits).
		Expect().
		Status(http.StatusOK).JSON().Array()

	for i, element := range fruitsData.Iter() {
		element.Object().ContainsKey("weight")
		element.Object().ContainsValue("fruit")
		element.Object().ContainsSubset(map[string]interface{}{
			"type": "fruit",
		})
		for j, _element := range element.Object().Iter() {
			fmt.Printf("index: %s, element: %v\n", j, _element.Raw())
		}
		fmt.Printf("index: %d, element: %v\n", i, element.Raw())	
	}
	
}
