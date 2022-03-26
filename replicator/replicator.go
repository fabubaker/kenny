package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/checkpoint-restore/go-criu/v5/rpc"
	"google.golang.org/protobuf/proto"
)

type Replicator struct {
	addr *net.UnixAddr
	opts *rpc.CriuOpts
}

func MakeReplicator(socketPath string, checkpointDir string) (*Replicator, error) {
	addr, err := net.ResolveUnixAddr("unixpacket", socketPath)
	if err != nil {
		return nil, err
	}

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
		addr: addr,
		opts: opts,
	}, nil
}

func (r *Replicator) Checkpoint(w http.ResponseWriter, httpreq *http.Request) {
	log.Printf("Received %v", httpreq)

	pidStr := httpreq.URL.Query().Get("pid")
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		log.Fatal(err)
	}

	pidAddr := proto.Int32(int32(pid))
	r.opts.Pid = pidAddr

	t := rpc.CriuReqType_DUMP
	req := rpc.CriuReq{
		Type: &t,
		Opts: r.opts,
	}

	mReq, err := proto.Marshal(&req)
	if err != nil {
		log.Fatal(err)
	}

	socket, err := net.DialUnix("unixpacket", nil, r.addr)
	if err != nil {
		log.Fatal(err)
	}

	// Make a dump request to the CRIU service
	_, err = socket.Write(mReq)
	if err != nil {
		log.Fatal(err)
	}

	mResp := make([]byte, 2*4096)
	bytesRead, err := socket.Read(mResp)
	if err != nil {
		log.Fatal(err)
	}

	resp := &rpc.CriuResp{}
	err = proto.Unmarshal(mResp[:bytesRead], resp)
	if err != nil {
		log.Fatal(err)
	}

	if !resp.GetSuccess() {
		fmt.Printf(
			"operation failed (msg:%s err:%d)",
			resp.GetCrErrmsg(), resp.GetCrErrno(),
		)
		return
	}

	w.WriteHeader(http.StatusOK)
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
	http.HandleFunc("/checkpoint", replicator.Checkpoint)
	log.Fatal(http.ListenAndServe(address, nil))
}
