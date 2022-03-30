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

const CHECKPOINT_INTERVAL = 2

type handler struct {
	activeAddr           string
	activeReplicatorAddr string
	activeProxy          *httputil.ReverseProxy

	standbyAddr           string
	standbyReplicatorAddr string
	standbyProxy          *httputil.ReverseProxy
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
		log.Printf("Heartbeat timed out, restoring and promoting %s to active!", h.standbyAddr)

		resp, err = http.Post(h.standbyReplicatorAddr+"/restore", "application/json", nil)
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Fatal("Unable to restore standby replica!")
		}

		// Swap the active and standby
		h.activeAddr, h.standbyAddr = h.standbyAddr, h.activeAddr
		h.activeReplicatorAddr, h.standbyReplicatorAddr = h.standbyReplicatorAddr, h.activeReplicatorAddr
		h.activeProxy, h.standbyProxy = h.standbyProxy, h.activeProxy
	}

	time.AfterFunc(CHECKPOINT_INTERVAL*time.Second, h.CheckHeartbeat)
}

// Usage: ./rproxy <IP of active server> <IP of active replicator> <IP of standby server> <IP of standby replicator>
func main() {
	activeAddr := os.Args[1]
	activeReplicatorAddr := os.Args[2]
	standbyAddr := os.Args[3]
	standbyReplicatorAddr := os.Args[4]

	activeProxy, err := NewProxy(activeAddr)
	if err != nil {
		panic(err)
	}
	standbyProxy, err := NewProxy(standbyAddr)
	if err != nil {
		panic(err)
	}

	handler := &handler{
		activeAddr:           activeAddr,
		activeReplicatorAddr: activeReplicatorAddr,
		activeProxy:          activeProxy,

		standbyAddr:           standbyAddr,
		standbyReplicatorAddr: standbyReplicatorAddr,
		standbyProxy:          standbyProxy,
	}

	portPtr := flag.String("port", "10000", "Port to listen on")

	flag.Parse()

	address := ":" + *portPtr

	log.Printf("Starting rproxy @ %s", address)
	log.Printf("Active replica: %s", activeAddr)
	log.Printf("Standby replica: %s", standbyAddr)
	time.AfterFunc(CHECKPOINT_INTERVAL*time.Second, handler.CheckHeartbeat)
	log.Fatal(http.ListenAndServe(address, handler))
}
