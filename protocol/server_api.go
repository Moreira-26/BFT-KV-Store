package protocol

import (
	"bftkvstore/context"
	"bftkvstore/logger"
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

	data, err := unmarshallJson[connectMsgBody](body)
	if err != nil {
		logger.Error("Error parsing connect message", err.Error())
		NewMessage(ERR).Send(conn)
		return
	}

	if data.Address == "" || data.Port == "" {
		NewMessage(ERR).Send(conn)
		return
	}

	logger.Info(fmt.Sprintf("Trying to connect to %s:%s", data.Address, data.Port))
	serverConn, connected := ConnectTo(ctx.Address, ctx.Port, data.Address, data.Port)

	if connected {
		NewMessage(OK).Send(conn)
		ctx.AddNewNode(data.Address, data.Port, serverConn)
	} else {
		NewMessage(NO).Send(conn)
	}
}

func qConnectMsg(ctx *context.AppContext, conn net.Conn, body []byte) (ok bool) {
	type connectMsgBody struct {
		Address string `json:"address"`
		Port    string `json:"port"`
	}

	data, err := unmarshallJson[connectMsgBody](body)
	if err != nil {
		logger.Error("parsing connect message", err)
		NewMessage(NO).Send(conn)
		return false
	}

	logger.Info(fmt.Sprintf("Received request to connect from %s:%s", data.Address, data.Port))
	ctx.AddNewNode(data.Address, data.Port, conn)

	NewMessage(OK).Send(conn)
	return true
}
