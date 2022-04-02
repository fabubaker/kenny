package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/checkpoint-restore/go-criu/v5/rpc"
	"google.golang.org/protobuf/proto"
)

const CHECKPOINT_INTERVAL_MS = 500

type Replicator struct {
	criuAddr          *net.UnixAddr
	dumpReq           rpc.CriuReq
	restoreReq        rpc.CriuReq
	checkpointDir     string
	checkpointCounter int
}

func MakeReplicator(socketPath string, checkpointDir string, pid int) (*Replicator, error) {
	addr, err := net.ResolveUnixAddr("unixpacket", socketPath)
	if err != nil {
		return nil, err
	}

	// Generate a skeleton for a marshaled dump request
	dumpOpts := &rpc.CriuOpts{
		Pid:          proto.Int32(int32(pid)),
		LogLevel:     proto.Int32(4),
		LogFile:      proto.String("dump.log"),
		LeaveRunning: proto.Bool(true),
		ShellJob:     proto.Bool(true),
		TcpClose:     proto.Bool(true),
	}

	dumpType := rpc.CriuReqType_DUMP
	dumpReq := rpc.CriuReq{
		Type: &dumpType,
		Opts: dumpOpts,
	}

	// Generate a skeleton for a marshaled restore request
	restoreOpts := &rpc.CriuOpts{
		LogLevel: proto.Int32(4),
		LogFile:  proto.String("restore.log"),
		ShellJob: proto.Bool(true),
		TcpClose: proto.Bool(true),
	}

	restoreType := rpc.CriuReqType_RESTORE
	restoreReq := rpc.CriuReq{
		Type: &restoreType,
		Opts: restoreOpts,
	}

	return &Replicator{
		criuAddr:          addr,
		dumpReq:           dumpReq,
		restoreReq:        restoreReq,
		checkpointDir:     checkpointDir,
		checkpointCounter: 0,
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

// @param iterative: whether to checkpoint iteratively
func (r *Replicator) Checkpoint(iterative bool) {
	log.Printf("Checkpointing...")

	r.checkpointCounter++
	dir := fmt.Sprintf("%s/%d", r.checkpointDir, r.checkpointCounter)
	err := os.Mkdir(dir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	dirfh, err := os.Open(dir)
	if err != nil {
		log.Fatal(err)
	}

	r.dumpReq.Opts.ImagesDirFd = proto.Int32(int32(dirfh.Fd()))

	if iterative {
		prevDir := fmt.Sprintf("%s/%d", r.checkpointDir, r.checkpointCounter-1)
		prevDirRel, err := filepath.Rel(dir, prevDir)
		if err != nil {
			log.Fatal(err)
		}

		r.dumpReq.Opts.ParentImg = proto.String(prevDirRel)
	}

	mReq, err := proto.Marshal(&r.dumpReq)
	if err != nil {
		log.Fatal(err)
	}

	r.sendAndRecv(mReq)

	time.AfterFunc(CHECKPOINT_INTERVAL_MS*time.Millisecond, func() { r.Checkpoint(false) })
}

func (r *Replicator) Restore() {
	log.Printf("Restoring...")

	dir := fmt.Sprintf("%s/%d", r.checkpointDir, r.checkpointCounter)
	dirfh, err := os.Open(dir)
	if err != nil {
		log.Fatal(err)
	}

	r.dumpReq.Opts.ImagesDirFd = proto.Int32(int32(dirfh.Fd()))

	mReq, err := proto.Marshal(&r.dumpReq)
	if err != nil {
		log.Fatal(err)
	}

	r.sendAndRecv(mReq)
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
		time.AfterFunc(CHECKPOINT_INTERVAL_MS*time.Millisecond, func() { replicator.Checkpoint(false) })
	}
	log.Fatal(http.ListenAndServe(address, replicator))
}
