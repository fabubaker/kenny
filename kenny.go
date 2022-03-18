package main

import (
	"fmt"
	"net/http"
)

func get(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello\n")
}

func put(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func main() {
	store := &Store{
		table: make(map[string]string),
	}

	handler := &Handler{
		store: store,
	}

	http.ListenAndServe(":8080", handler)
}
