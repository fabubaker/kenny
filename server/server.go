package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/fabubaker/kenny/server/handler"
	"github.com/fabubaker/kenny/server/store"
)

func main() {
	store := &store.Store{
		Table: make(map[string]map[string]string),
	}

	handler := &handler.Handler{
		Store: store,
	}

	portPtr := flag.String("port", "8080", "Port to listen on")
	replicatorAddrPtr := flag.String("replicator-addr", "localhost:9090", "Address of the replicator")

	flag.Parse()

	address := ":" + *portPtr
	replicatorAddr := "http://" + *replicatorAddrPtr

	// Setup ticker to capture periodic checkpoints
	ticker := time.NewTicker(200 * time.Millisecond)
	go func() {
		for {
			select {
			case _ = <-ticker.C:
				handler.Checkpoint(replicatorAddr)
			}
		}
	}()

	log.Printf("Starting server @ %s", address)
	log.Fatal(http.ListenAndServe(address, handler))
}
