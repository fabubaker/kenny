package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/checkpoint-restore/go-criu/v5/rpc"
	"google.golang.org/protobuf/proto"
)

type Replicator struct {
	addr *net.UnixAddr
	mReq []byte
	pid  int
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

	opts := &rpc.CriuOpts{
		Pid:          proto.Int32(int32(pid)),
		LogLevel:     proto.Int32(4),
		LogFile:      proto.String("dump.log"),
		LeaveRunning: proto.Bool(true),
		ImagesDirFd:  proto.Int32(int32(dir.Fd())),
		ShellJob:     proto.Bool(true),
		TcpClose:     proto.Bool(true),
	}

	t := rpc.CriuReqType_DUMP
	req := rpc.CriuReq{
		Type: &t,
		Opts: opts,
	}

	mReq, err := proto.Marshal(&req)
	if err != nil {
		log.Fatal(err)
	}

	return &Replicator{
		addr: addr,
		mReq: mReq,
	}, nil
}

func (r *Replicator) Checkpoint() {
	log.Printf("Checkpointing...")

	socket, err := net.DialUnix("unixpacket", nil, r.addr)
	if err != nil {
		log.Fatal(err)
	}

	// Make a dump request to the CRIU service
	_, err = socket.Write(r.mReq)
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

	time.AfterFunc(500*time.Millisecond, r.Checkpoint)
}

func main() {
	pidPtr := flag.Int("pid", 0, "PID of the process to checkpoint")
	portPtr := flag.String("port", "9090", "Port to listen on")
	flag.Parse()

	pid := *pidPtr
	address := ":" + *portPtr

	replicator, err := MakeReplicator("/tmp/kenny.sock", "/tmp/kenny/checkpoint", pid)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting replicator @ %s", address)
	timer := time.AfterFunc(500*time.Millisecond, replicator.Checkpoint)
	<-timer.C // Block forever here
}
