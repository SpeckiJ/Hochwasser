package rpc

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
)

// const handshake_magick = "Sæl!"

func ConnectHevring(ránAddress string) {
	rpc.Register(new(Hevring))

	fmt.Printf("[hevring] greeting Rán at %s\n", ránAddress)
	conn, err := net.Dial("tcp", ránAddress)
	if err != nil {
		log.Fatal(err)
	}
	go rpc.ServeConn(conn)
	fmt.Printf("[hevring] awaiting task from Rán\n")
}

type Hevring struct {}

type FlutAck struct{ Ok bool }

func (h *Hevring) Flut(job RánJob, reply *FlutAck) error {
	fmt.Printf("[hevring] Rán gave us /w o r k/! %v\n", job)
	reply.Ok = true
	return nil
}

func (h *Hevring) Status(x int, reply *FlutAck) error {
	reply.Ok = true
	return nil
}

func (h *Hevring) Stop(x int, reply *FlutAck) error {
	reply.Ok = true
	return nil
}
