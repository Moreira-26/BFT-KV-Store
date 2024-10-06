package main

import (
	"bftkvstore/config"
	"bftkvstore/context"
	"bftkvstore/protocol"
	"bftkvstore/utils"
	"flag"
	"fmt"
	"log"
)

var serverHostname string
var serverPortPtr *string
var configPathPtr *string

func init() {
	serverHostname = utils.GetOutboundIP().String()
	serverPortPtr = flag.String("port", "8089", "specifies which port must be used by the application")
	configPathPtr = flag.String("config", "bftkvstore.toml", "specifies the path for a configuration file")
}

func main() {
	flag.Parse()
	flagset := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })

	var nodeConfig config.ConfigData

	if flagset["config"] { // config was set
		var err error
		nodeConfig, err = config.ReadConfig(*configPathPtr)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		nodeConfig = config.WriteConfig()
	}

	var serverPort string = *serverPortPtr

	log.Printf("Server started: %s:%s\n", serverHostname, serverPort)

	var ctx context.AppContext = context.AppContext{
		Secretkey: nodeConfig.Sk,
		Publickey: nodeConfig.Pk,
		Address:   serverHostname,
		Port:      serverPort,
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
