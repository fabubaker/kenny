package main

import (
	"log"
	"net/http"

	"github.com/fabubaker/kenny/server/handler"
	"github.com/fabubaker/kenny/server/store"
)

func main() {
	store := &store.Store{
		Table: make(map[string]string),
	}

	handler := &handler.Handler{
		Store: store,
	}

	log.Println("Starting server...")
	http.ListenAndServe(":8080", handler)
}
