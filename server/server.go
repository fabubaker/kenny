package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/checkpoint-restore/go-criu/v5/rpc"
	"github.com/fabubaker/kenny/server/handler"
	"github.com/fabubaker/kenny/server/store"
	"google.golang.org/protobuf/proto"
)

func main() {
	store := &store.Store{
		Table: make(map[string]string),
	}

	handler := &handler.Handler{
		Store: store,
	}

	portPtr := flag.String("port", "8080", "Port to listen on")

	flag.Parse()

	address := ":" + *portPtr

	/* CRIU TEST */
	addr, err := net.ResolveUnixAddr("unix", "/home/fabubaker/Waterloo/Courses/CS_654/research_project/server/criu_service.socket")
	if err != nil {
		fmt.Printf("Failed to resolve: %v\n", err)
		os.Exit(1)
	}

	socket, err := net.DialUnix("unixpacket", nil, addr)
	if err != nil {
		panic(err)
	}

	img, err := os.Open("./checkpoint")
	if err != nil {
		log.Fatal("can't open image dir: %w", err)
		return
	}

	opts := &rpc.CriuOpts{
		LogLevel:     proto.Int32(4),
		LogFile:      proto.String("pre-dump.log"),
		LeaveRunning: proto.Bool(true),
		ImagesDirFd:  proto.Int32(int32(img.Fd())),
		ShellJob:     proto.Bool(true),
	}

	t := rpc.CriuReqType_DUMP
	req := rpc.CriuReq{
		Type: &t,
		Opts: opts,
	}

	reqB, err := proto.Marshal(&req)
	if err != nil {
		log.Fatal(err)
		return
	}

	_, err = socket.Write(reqB)
	if err != nil {
		log.Fatal(err)
		return
	}

	respB := make([]byte, 2*4096)
	n, err := socket.Read(respB)
	if err != nil {
		log.Fatal(err)
		return
	}

	resp := &rpc.CriuResp{}
	err = proto.Unmarshal(respB[:n], resp)
	if err != nil {
		log.Fatal(err)
		return
	}

	if !resp.GetSuccess() {
		log.Fatal(err)
		return
	}

	/* */

	log.Printf("Starting server @ %s", address)
	log.Fatal(http.ListenAndServe(address, handler))
}
