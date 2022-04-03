package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/fabubaker/kenny/server/handler"
	"github.com/fabubaker/kenny/server/store"
)

const REPLICATOR_CHECK_RETRY_INTERVAL = 5

func main() {
	store := &store.Store{
		Table: make(map[string]map[string]string),
	}

	portPtr := flag.String("port", "8080", "Port to listen on")
	interval := flag.Int("interval", 500, "Checkpointing interval in milliseconds")
	minChanges := flag.Int("changes", 100, "Minimum number of changes to trigger a checkpoint")
	replicatorAddr := flag.String("replicator", "http://localhost:9090", "Address of the replicator")

	flag.Parse()

	handler := &handler.Handler{
		Store:                   store,
		Interval:                time.Duration(*interval) * time.Millisecond,
		MinChanges:              *minChanges,
		CurrentChanges:          0,
		LastCheckpointedChanges: 0,
		ReplicatorAddr:          *replicatorAddr,
	}

	address := ":" + *portPtr

	// Create an initial dump, then subsequent checkpoints are iteratively captured.
	iterative := false
	handler.Checkpoint(iterative)

	log.Printf("Starting server @ %s", address)
	log.Fatal(http.ListenAndServe(address, handler))
}
