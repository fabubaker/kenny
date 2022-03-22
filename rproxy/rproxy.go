package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type handler struct {
	activeIP    string
	activeProxy *httputil.ReverseProxy

	standbyIP    string
	standbyProxy *httputil.ReverseProxy
}

func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Promote requests are handled as a special case
	if req.Method == "POST" && req.RequestURI == "/promote" {
		log.Printf("Received request to promote %s as active!", h.standbyIP)

		h.activeIP = h.standbyIP
		h.activeProxy = h.standbyProxy

		w.WriteHeader(http.StatusOK)
		return
	}

	// Everything else is forwarded to the active server
	log.Printf("Forwarding %s %s to: %s", req.Method, req.RequestURI, h.activeIP)
	h.activeProxy.ServeHTTP(w, req)
}

func NewProxy(targetHost string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}

	return httputil.NewSingleHostReverseProxy(url), nil
}

// Usage: ./rproxy <IP of active server> <IP of passive server>
func main() {
	activeIP := os.Args[1]
	standbyIP := os.Args[2]

	activeProxy, err := NewProxy(activeIP)
	if err != nil {
		panic(err)
	}
	standbyProxy, err := NewProxy(standbyIP)
	if err != nil {
		panic(err)
	}

	handler := &handler{
		activeIP:    activeIP,
		activeProxy: activeProxy,

		standbyIP:    standbyIP,
		standbyProxy: standbyProxy,
	}

	portPtr := flag.String("port", "8080", "Port to listen on")

	flag.Parse()

	address := ":" + *portPtr

	log.Printf("Starting rproxy @ %s", address)
	log.Printf("Active replica: %s", activeIP)
	log.Printf("Standby replica: %s", standbyIP)

	log.Fatal(http.ListenAndServe(address, handler))
}
