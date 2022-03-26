package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/fabubaker/kenny/server/store"
)

type Handler struct {
	Store *store.Store
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	key := req.RequestURI

	switch req.Method {
	case "GET":
		var fields []string

		err := json.NewDecoder(req.Body).Decode(&fields)

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
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
