package handler

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/fabubaker/kenny/server/store"
)

type Handler struct {
	Store *store.Store
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var value string
	key := req.RequestURI

	switch req.Method {
	case "GET":
		value = h.Store.Get(key)
		fmt.Fprintf(w, "%s", value)
	case "PUT":
		body, err := ioutil.ReadAll(req.Body)

		if err != nil {
			log.Fatal(err)
		}

		value = string(body)
		h.Store.Put(key, value)
	case "DELETE":
		value = h.Store.Delete(key)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	log.Printf("%s %s -> %s", req.Method, key, value)
}
