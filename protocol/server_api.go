package protocol

import (
	"bftkvstore/context"
	"bftkvstore/logger"
	"encoding/json"
	"fmt"
	"net"
)

func pingMsg(conn net.Conn) {
	NewMessage(PONG).Send(conn)
}

func connectMsg(ctx *context.AppContext, conn net.Conn, body []byte) {
	type connectMsgBody struct {
		Address string `json:"address"`
		Port    string `json:"port"`
	}

	var data connectMsgBody

	err := json.Unmarshal(body, &data)

	if err != nil {
		logger.Error("Error parsing connect message", err.Error())
		sendString(conn, BadArgumentsError())
		return
	}

	logger.Info("Received CONN message with arguments", fmt.Sprint(data))

	if data.Address == "" || data.Port == "" {
		sendString(conn, BadArgumentsError())
		return
	}

	logger.Info(fmt.Sprintf("Trying to connect to %s:%s", data.Address, data.Port))

	connected := ConnectTo(ctx.Address, ctx.Port, data.Address, data.Port)

	var response string = ""
	if connected {
		response = fmt.Sprintf("Connected to %s:%s successfully", data.Address, data.Port)
		ctx.AddNewNode(data.Address, data.Port)
	} else {
		response = fmt.Sprintf("Failed to connect to %s:%s", data.Address, data.Port)
	}
	logger.Debug(response)
	sendString(conn, response)
}

func qConnectMsg(ctx *context.AppContext, conn net.Conn, body []byte) {
	type connectMsgBody struct {
		Address string `json:"address"`
		Port    string `json:"port"`
	}

	var data connectMsgBody

	err := json.Unmarshal(body, &data)
	if err != nil {
		logger.Error("parsing connect message", err.Error())
		NewMessage(NO).Send(conn)
		return
	}

	logger.Info(fmt.Sprintf("Received request to connect from %s:%s", data.Address, data.Port))
	ctx.AddNewNode(data.Address, data.Port)

	NewMessage(OK).Send(conn)
}
