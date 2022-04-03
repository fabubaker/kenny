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
	minChanges := flag.Int("keys", 100, "Minimum number of changes to trigger a checkpoint")
	replicatorAddr := flag.String("replicator", "http://localhost:9090", "Address of the replicator")

	handler := &handler.Handler{
		Store:          store,
		Interval:       time.Duration(*interval),
		MinChanges:     *minChanges,
		CurrentPuts:    0,
		LastSeenPuts:   0,
		ReplicatorAddr: *replicatorAddr,
	}

	flag.Parse()

	address := ":" + *portPtr

	log.Printf("Checking if replicator is up at %s", handler.ReplicatorAddr)
	for alive := false; !alive; {
		err := handler.ReplicatorCheck()
		if err == nil {
			alive = true
		} else {
			log.Printf("Replicator not alive, retrying in %d seconds", REPLICATOR_CHECK_RETRY_INTERVAL)
			time.Sleep(REPLICATOR_CHECK_RETRY_INTERVAL * time.Second)
		}
	}

	time.AfterFunc(handler.Interval*time.Millisecond, handler.Checkpoint)
	log.Printf("Starting server @ %s", address)
	log.Fatal(http.ListenAndServe(address, handler))
}
