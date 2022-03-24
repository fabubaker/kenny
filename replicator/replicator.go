package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/checkpoint-restore/go-criu/v5"
	"github.com/checkpoint-restore/go-criu/v5/rpc"
	"google.golang.org/protobuf/proto"
)

type Replicator struct {
	criu *criu.Criu
	opts *rpc.CriuOpts
}

func MakeReplicator(socketPath string, checkpointDir string) (*Replicator, error) {
	criu := criu.MakeCriu()

	dir, err := os.Open(checkpointDir)
	if err != nil {
		return nil, err
	}

	opts := &rpc.CriuOpts{
		LogLevel:     proto.Int32(4),
		LogFile:      proto.String("dump.log"),
		LeaveRunning: proto.Bool(true),
		ImagesDirFd:  proto.Int32(int32(dir.Fd())),
		ShellJob:     proto.Bool(true),
		TcpClose:     proto.Bool(true),
	}

	return &Replicator{
		criu: criu,
		opts: opts,
	}, nil
}

func (r *Replicator) Replicate(w http.ResponseWriter, httpreq *http.Request) {
	log.Printf("Received %v", httpreq)

	pidStr := httpreq.URL.Query().Get("pid")
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		log.Fatal(err)
	}

	pidAddr := proto.Int32(int32(pid))
	r.opts.Pid = pidAddr

	err = r.criu.Dump(r.opts, criu.NoNotify{})
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	portPtr := flag.String("port", "9090", "Port to listen on")
	flag.Parse()

	address := ":" + *portPtr

	replicator, err := MakeReplicator("/tmp/kenny.sock", "/tmp/kenny/checkpoint")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting server @ %s", address)
	http.HandleFunc("/replicate", replicator.Replicate)
	log.Fatal(http.ListenAndServe(address, nil))
}
