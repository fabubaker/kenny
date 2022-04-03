package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/fabubaker/kenny/server/store"
)

type Handler struct {
	Store          *store.Store
	Interval       time.Duration
	MinChanges     int
	CurrentPuts    int
	LastSeenPuts   int
	ReplicatorAddr string
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

		h.CurrentPuts++
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

func (h *Handler) Checkpoint() {
	log.Printf("Checkpointing to %s", h.ReplicatorAddr)

	base, err := url.Parse(h.ReplicatorAddr + "/checkpoint")
	if err != nil {
		log.Fatal(err)
	}

	_, err = http.Post(base.String(), "application/json", nil)
	if err != nil {
		log.Println(err)
	}

	time.AfterFunc(h.Interval*time.Millisecond, h.Checkpoint)
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
