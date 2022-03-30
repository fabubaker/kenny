package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/checkpoint-restore/go-criu/v5/rpc"
	"google.golang.org/protobuf/proto"
)

type Replicator struct {
	criuAddr   *net.UnixAddr
	dumpReq    []byte
	restoreReq []byte
}

func MakeReplicator(socketPath string, checkpointDir string, pid int) (*Replicator, error) {
	addr, err := net.ResolveUnixAddr("unixpacket", socketPath)
	if err != nil {
		return nil, err
	}

	dir, err := os.Open(checkpointDir)
	if err != nil {
		return nil, err
	}

	// Generate a marshaled dump request
	dumpOpts := &rpc.CriuOpts{
		Pid:          proto.Int32(int32(pid)),
		LogLevel:     proto.Int32(4),
		LogFile:      proto.String("dump.log"),
		LeaveRunning: proto.Bool(true),
		ImagesDirFd:  proto.Int32(int32(dir.Fd())),
		ShellJob:     proto.Bool(true),
		TcpClose:     proto.Bool(true),
	}

	dumpType := rpc.CriuReqType_DUMP
	dumpReq := rpc.CriuReq{
		Type: &dumpType,
		Opts: dumpOpts,
	}

	mDumpReq, err := proto.Marshal(&dumpReq)
	if err != nil {
		log.Fatal(err)
	}

	// Generate a marshaled restore request
	restoreOpts := &rpc.CriuOpts{
		LogLevel:    proto.Int32(4),
		LogFile:     proto.String("restore.log"),
		ImagesDirFd: proto.Int32(int32(dir.Fd())),
		ShellJob:    proto.Bool(true),
		TcpClose:    proto.Bool(true),
	}

	restoreType := rpc.CriuReqType_RESTORE
	restoreReq := rpc.CriuReq{
		Type: &restoreType,
		Opts: restoreOpts,
	}

	mRestoreReq, err := proto.Marshal(&restoreReq)
	if err != nil {
		log.Fatal(err)
	}

	return &Replicator{
		criuAddr:   addr,
		dumpReq:    mDumpReq,
		restoreReq: mRestoreReq,
	}, nil
}

func (r *Replicator) sendAndRecv(msg []byte) {
	socket, err := net.DialUnix("unixpacket", nil, r.criuAddr)
	if err != nil {
		log.Fatal(err)
	}

	_, err = socket.Write(msg)
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
}

func (r *Replicator) Checkpoint() {
	log.Printf("Checkpointing...")

	r.sendAndRecv(r.dumpReq)

	time.AfterFunc(500*time.Millisecond, r.Checkpoint)
}

func (r *Replicator) Restore() {
	log.Printf("Restoring...")

	r.sendAndRecv(r.restoreReq)
}

func (r *Replicator) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("Received request: %s", req.URL.Path)

	switch req.URL.Path {
	case "/restore":
		r.Restore()
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func main() {
	pidPtr := flag.Int("pid", 0, "PID of the process to checkpoint")
	portPtr := flag.String("port", "9090", "Port to listen on")
	activePtr := flag.Bool("active", true, "Run in active mode")
	flag.Parse()

	pid := *pidPtr
	address := ":" + *portPtr
	active := *activePtr

	replicator, err := MakeReplicator("/tmp/kenny.sock", "/tmp/kenny/checkpoint", pid)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting replicator @ %s", address)
	if active {
		time.AfterFunc(500*time.Millisecond, replicator.Checkpoint)
	}
	log.Fatal(http.ListenAndServe(address, replicator))
}
