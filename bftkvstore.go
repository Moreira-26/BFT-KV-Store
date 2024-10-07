package main

import (
	"bftkvstore/config"
	"bftkvstore/context"
	"bftkvstore/protocol"
	"bftkvstore/utils"
	"flag"
	"fmt"
	"log"
	"os"
)

var serverHostname string
var serverPortPtr *string
var configPathPtr *string

func init() {
	serverHostname = utils.GetOutboundIP().String()
	serverPortPtr = flag.String("port", "8089", "specifies which port must be used by the application")
	configPathPtr = flag.String("config", ".kvstoreconfig", "specifies the path for a configuration file")
}

func main() {
	flag.Parse()

	var serverPort string = *serverPortPtr
	var configPath string = *configPathPtr

	var nodeConfig config.ConfigData
	configFolderInfo, findConfigFolderErr := os.Stat(*configPathPtr)

	if os.IsNotExist(findConfigFolderErr) {
		log.Printf("No configuration exists on path %s. Creating a new one.\n", configPath)
		nodeConfig = config.WriteConfig(configPath)
		log.Printf("Configuration %s created successfully\n", configPath)
	} else if !configFolderInfo.IsDir() {
		log.Fatalln(fmt.Sprintf("The provided configuration path %s should be a folder.\n", *configPathPtr))
	} else {
		var err error
		nodeConfig, err = config.ReadConfig(configPath)
		if err != nil {
			log.Fatalln(err.Error())
		}
		log.Printf("Configuration %s read successfully\n", configPath)
	}

	log.Printf("Server started: %s:%s\n", serverHostname, serverPort)

	var ctx context.AppContext = context.AppContext{
		Secretkey: nodeConfig.Sk,
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
