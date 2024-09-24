package main

import (
	"bftkvstore/protocol"
	"bftkvstore/utils"
	"fmt"
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

	if serverPort == "" {
		fmt.Println("A server port must be specified")
		os.Exit(1)
	}

	log.Printf("Server started: %s:%s\n", serverHostname, serverPort)

	go protocol.ReceiverStart(serverPort)

	for {
		// wait for end
	}
}
