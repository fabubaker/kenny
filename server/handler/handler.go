package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/fabubaker/kenny/server/store"
)

type Handler struct {
	Store *store.Store
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

func (h *Handler) Checkpoint(replicatorAddr string) {
	log.Printf("Checkpointing to %s", replicatorAddr)

	base, err := url.Parse(replicatorAddr + "/checkpoint")
	if err != nil {
		log.Fatal(err)
	}

	params := url.Values{}
	params.Add("pid", strconv.Itoa(os.Getpid()))
	base.RawQuery = params.Encode()

	_, err = http.Post(base.String(), "application/json", nil)
	if err != nil {
		log.Println(err)
	}
}
