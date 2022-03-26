package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

type handler struct {
	activeAddr  string
	activeProxy *httputil.ReverseProxy

	standbyAddr  string
	standbyProxy *httputil.ReverseProxy
}

func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("Forwarding %s %s to: %s", req.Method, req.RequestURI, h.activeAddr)
	h.activeProxy.ServeHTTP(w, req)
}

func NewProxy(targetHost string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}

	return httputil.NewSingleHostReverseProxy(url), nil
}

func (h *handler) CheckHeartbeat() {
	log.Println("Checking heartbeat...")
	resp, err := http.Post(h.activeAddr+"/heartbeat", "application/json", nil)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Heartbeat timed out, promoting %s to active!", h.standbyAddr)
	}

	time.AfterFunc(5*time.Second, h.CheckHeartbeat)
}

// Usage: ./rproxy <IP of active server> <IP of passive server>
func main() {
	activeAddr := os.Args[1]
	standbyAddr := os.Args[2]

	activeProxy, err := NewProxy(activeAddr)
	if err != nil {
		panic(err)
	}
	standbyProxy, err := NewProxy(standbyAddr)
	if err != nil {
		panic(err)
	}

	handler := &handler{
		activeAddr:  activeAddr,
		activeProxy: activeProxy,

		standbyAddr:  standbyAddr,
		standbyProxy: standbyProxy,
	}

	portPtr := flag.String("port", "8080", "Port to listen on")

	flag.Parse()

	address := ":" + *portPtr

	log.Printf("Starting rproxy @ %s", address)
	log.Printf("Active replica: %s", activeAddr)
	log.Printf("Standby replica: %s", standbyAddr)
	time.AfterFunc(5*time.Second, handler.CheckHeartbeat)
	log.Fatal(http.ListenAndServe(address, handler))
}
