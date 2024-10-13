package main

import (
	"bftkvstore/config"
	"bftkvstore/context"
	"bftkvstore/logger"
	"bftkvstore/protocol"
	"bftkvstore/utils"
	"flag"
	"fmt"
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
		logger.Alert(fmt.Sprintf("No configuration exists on path %s. Creating a new one.", configPath))
		nodeConfig = config.WriteConfig(configPath)
		logger.Info(fmt.Sprintf("Configuration %s created successfully", configPath))
	} else if !configFolderInfo.IsDir() {
		logger.Fatal(fmt.Sprintf("The provided configuration path %s should be a folder.", *configPathPtr))
	} else {
		var err error
		nodeConfig, err = config.ReadConfig(configPath)
		if err != nil {
			logger.Fatal(err)
		}
		logger.Info(fmt.Sprintf("Configuration %s read successfully", configPath))
	}

	logger.Info(fmt.Sprintf("Server started: %s:%s", serverHostname, serverPort))

	var ctx context.AppContext = context.New(nodeConfig.Sk, serverHostname, serverPort)

	go protocol.ReceiverStart(&ctx, serverPort)

	protocol.BroadcastReceiver(&ctx)
}
