// Package fruits is usage example for httpexpect.
package fruits

import (
	"encoding/json"
	"net/http"
	"path"
)

var (
	fruits = make(map[string]interface{})
)

// FruitServer creates http.Handler for fruits server.
//
// Implemented API:
//  GET /fruits           get fruit list
//  GET /fruits/{name}    get fruit
//  PUT /fruits/{name}    add or update fruit
func FruitServer() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/fruits", handleFruitList)
	mux.HandleFunc("/fruits/", handleFruit)

	return mux
}

func handleFruitList(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		ret := []string{}
		for k := range fruits {
			ret = append(ret, k)
		}

		b, err := json.Marshal(ret)
		if err != nil {
			panic(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(b)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleFruit(w http.ResponseWriter, r *http.Request) {
	_, name := path.Split(r.URL.Path)

	switch r.Method {
	case "GET":
		if data, ok := fruits[name]; ok {
			b, err := json.Marshal(data)
			if err != nil {
				panic(err)
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(b)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}

	case "PUT":
		var data map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			panic(err)
		}

		fruits[name] = data
		w.WriteHeader(http.StatusNoContent)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
