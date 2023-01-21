package examples

import (
	"encoding/json"
	"net/http"
)

var fruits []string = []string{"apple", "orange", "melon"}

func CustomFormatterHandler() http.Handler {

	mux := http.NewServeMux()

	mux.HandleFunc("/fruits", func(w http.ResponseWriter, r *http.Request) {
		handleFruits(fruits, w, r)
	})

	return mux
}

func handleFruits(fruits []string, w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		b, err := json.Marshal(fruits)
		if err != nil {
			panic(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
