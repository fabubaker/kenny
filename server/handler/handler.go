package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/fabubaker/kenny/server/store"
)

type Handler struct {
	Store                   *store.Store
	Interval                time.Duration
	MinChanges              int
	CurrentChanges          int
	LastCheckpointedChanges int
	ReplicatorAddr          string
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	key := req.URL.Path

	switch req.Method {
	case "GET":
		var fields []string

		queryMap, err := url.ParseQuery(req.URL.RawQuery)
		if err != nil {
			log.Fatal(err)
		}
		fields = queryMap["fields"]

		if err != nil {
			log.Fatal(err)
		}

		log.Printf("%s %s -> %s", req.Method, key, fields)
		values := h.Store.Get(key, fields)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(values)
	case "PUT":
		var values map[string]string

		err := json.NewDecoder(req.Body).Decode(&values)

		if err != nil {
			log.Fatal(err)
		}

		log.Printf("%s %s -> %s", req.Method, key, values)
		h.Store.Put(key, values)

		h.CurrentChanges++
		w.WriteHeader(http.StatusOK)
	case "DELETE":
		log.Printf("%s %s", req.Method, key)
		h.Store.Delete(key)

		w.WriteHeader(http.StatusOK)
	case "POST":
		if key == "/heartbeat" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) Checkpoint(iterative bool) {
	defer time.AfterFunc(h.Interval, func() { h.Checkpoint(true) })

	// Skip checkpointing if not enough changes have been made since last time.
	if iterative && h.CurrentChanges-h.LastCheckpointedChanges < h.MinChanges {
		return
	}

	log.Printf("Checkpointing to %s", h.ReplicatorAddr)

	pid := os.Getpid()

	base, err := url.Parse(h.ReplicatorAddr + "/checkpoint")
	if err != nil {
		log.Fatal(err)
	}

	params := url.Values{}
	params.Add("pid", fmt.Sprintf("%d", pid))
	params.Add("iterative", fmt.Sprintf("%t", iterative))
	base.RawQuery = params.Encode()

	_, err = http.Post(base.String(), "application/json", nil)
	if err != nil {
		log.Println(err)
	}

	h.LastCheckpointedChanges = h.CurrentChanges
}

func (h *Handler) ReplicatorCheck() error {
	base, err := url.Parse(h.ReplicatorAddr + "/heartbeat")
	if err != nil {
		log.Fatal(err)
	}

	_, err = http.Get(base.String())
	if err != nil {
		return err
	}

	return nil
}
