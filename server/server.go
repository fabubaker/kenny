package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/fabubaker/kenny/server/handler"
	"github.com/fabubaker/kenny/server/store"
)

func main() {
	store := &store.Store{
		Table: make(map[string]map[string]string),
	}

	handler := &handler.Handler{
		Store:        store,
		CurrentPuts:  0,
		LastSeenPuts: 0,
	}

	portPtr := flag.String("port", "8080", "Port to listen on")

	flag.Parse()

	address := ":" + *portPtr

	log.Printf("Starting server @ %s", address)
	log.Fatal(http.ListenAndServe(address, handler))
}
