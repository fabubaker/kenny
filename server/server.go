package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/fabubaker/kenny/server/handler"
	"github.com/fabubaker/kenny/server/replicator"
	"github.com/fabubaker/kenny/server/store"
)

func main() {
	portPtr := flag.String("port", "8080", "Port to listen on")
	flag.Parse()

	store := &store.Store{
		Table: make(map[string]string),
	}

	replicator, err := replicator.MakeReplicator("/tmp/kenny.sock", "/tmp/kenny/checkpoint")
	if err != nil {
		log.Fatal(err)
	}

	handler := &handler.Handler{
		Store:      store,
		Replicator: replicator,
	}

	address := ":" + *portPtr

	handler.Replicator.Replicate()

	log.Printf("Starting server @ %s", address)
	log.Fatal(http.ListenAndServe(address, handler))
}
