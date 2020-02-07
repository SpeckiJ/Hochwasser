package rpc

import (
	"fmt"
	"log"
	"net/rpc"
)

func ConnectHevring(ránAddress string) {
	fmt.Printf("[hevring] connecting to %s\n", ránAddress)

	client, err := rpc.Dial("tcp", ránAddress)
	if err != nil {
		log.Fatal(err)
	}

	job := RánJob{}
	err = client.Call("Rán.Hello", RánHelloReq{}, &job)
	if err != nil {
		log.Fatal(err)
	}

	if (job == RánJob{}) {
		fmt.Printf("[hevring] Rán has no job for us. :(\n")
	} else {
		fmt.Printf("[hevring] Rán gave us /w o r k/! %v\n", job)
	}
}
