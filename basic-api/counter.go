package main

import (
	"encoding/json"
	"log"
	"net/http"
)

var (
	counter = make(map[string]int)
)

func handler(w http.ResponseWriter, r *http.Request) {
	ids := r.URL.Query()["id"]

	if len(ids) == 1 {
		counter[ids[0]]++
	}

	data, err := json.Marshal(counter)

	if err != nil {
		w.WriteHeader(500)
		return
	}

	w.Write(data)
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
