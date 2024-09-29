package main

import (
	"bftkvstore/context"
	"bftkvstore/protocol"
	"bftkvstore/utils"
	"log"
	"os"
)

var serverPort string
var serverHostname string

func main() {
	argsWithoutProg := os.Args[1:]

	serverHostname = utils.GetOutboundIP().String()
	if len(argsWithoutProg) > 0 {
		serverPort = argsWithoutProg[0]
	} else {
		serverPort = "8089"
	}

	log.Printf("Server started: %s:%s\n", serverHostname, serverPort)

	var ctx context.AppContext = context.AppContext{
		Address: serverHostname,
		Port: serverPort,
		Nodes: make([]struct {
			Address string
			Port    string
		}, 0),
	}

	go protocol.ReceiverStart(&ctx, serverPort)

	for {
		// wait for end
	}
}
