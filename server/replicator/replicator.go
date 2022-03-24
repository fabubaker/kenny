package replicator

import (
	"fmt"
	"net"
	"os"

	"github.com/checkpoint-restore/go-criu/v5/rpc"
	"google.golang.org/protobuf/proto"
)

type Replicator struct {
	socket *net.UnixConn
	req    []byte
}

func MakeReplicator(socketPath string, checkpointDir string) (*Replicator, error) {
	addr, err := net.ResolveUnixAddr("unixpacket", socketPath)
	if err != nil {
		return nil, err
	}

	socket, err := net.DialUnix("unixpacket", nil, addr)
	if err != nil {
		return nil, err
	}

	dir, err := os.Open(checkpointDir)
	if err != nil {
		return nil, err
	}

	opts := &rpc.CriuOpts{
		LogLevel:     proto.Int32(4),
		LogFile:      proto.String("pre-dump.log"),
		LeaveRunning: proto.Bool(true),
		ImagesDirFd:  proto.Int32(int32(dir.Fd())),
		ShellJob:     proto.Bool(true),
	}

	t := rpc.CriuReqType_DUMP
	req := rpc.CriuReq{
		Type: &t,
		Opts: opts,
	}

	mReq, err := proto.Marshal(&req)
	if err != nil {
		return nil, err
	}

	return &Replicator{
		socket: socket,
		req:    mReq,
	}, nil
}

func (r *Replicator) Replicate() error {
	// Make a dump request to the CRIU service
	_, err := r.socket.Write(r.req)
	if err != nil {
		return err
	}

	mResp := make([]byte, 2*4096)
	bytesRead, err := r.socket.Read(mResp)
	if err != nil {
		return err
	}

	resp := &rpc.CriuResp{}
	err = proto.Unmarshal(mResp[:bytesRead], resp)
	if err != nil {
		return err
	}

	if !resp.GetSuccess() {
		return fmt.Errorf(
			"operation failed (msg:%s err:%d)",
			resp.GetCrErrmsg(), resp.GetCrErrno(),
		)
	}

	return nil
}
