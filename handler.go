package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Handler struct {
	store *Store
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var value string
	key := req.RequestURI

	switch req.Method {
	case "GET":
		value = h.store.Get(key)
		fmt.Fprintf(w, "%s", value)
	case "PUT":
		body, err := ioutil.ReadAll(req.Body)

		if err != nil {
			log.Fatal(err)
		}

		value = string(body)
		h.store.Put(key, value)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	log.Printf("%s %s -> %s", req.Method, key, value)
}
