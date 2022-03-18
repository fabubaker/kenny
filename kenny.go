package main

import (
	"log"
	"net/http"
)

func main() {
	store := &Store{
		table: make(map[string]string),
	}

	handler := &Handler{
		store: store,
	}

	log.Println("Starting server...")
	http.ListenAndServe(":8080", handler)
}
