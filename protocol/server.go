package protocol

import (
	"bftkvstore/context"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

func Router(ctx *context.AppContext, conn net.Conn, msg Message) {
	switch msg.Header {
	// server api
	case PING:
		pingMsg(conn)
	case CONNECT:
		connectMsg(ctx, conn, msg.Content)
	case Q_CONNECT:
		qConnectMsg(ctx, conn, msg.Content)

	// user api
	case NEW:
		newMsg(conn, msg.Content)
	default:
		// SendString(conn, CmdNotFoundError(header))
	}
}

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
		log.Println("Error parsing connect message", err)
		sendString(conn, BadArgumentsError())
		return
	}

	log.Println("Received CONN message with arguments", data)

	if data.Address == "" || data.Port == "" {
		sendString(conn, BadArgumentsError())
		return
	}

	log.Printf("Trying to connect to %s:%s\n", data.Address, data.Port)

	connected := ConnectTo(ctx.Address, ctx.Port, data.Address, data.Port)

	var response string = ""
	if connected {
		response = fmt.Sprintf("Connected to %s:%s successfully", data.Address, data.Port)
		ctx.AddNewNode(data.Address, data.Port)
	} else {
		response = fmt.Sprintf("Failed to connect to %s:%s", data.Address, data.Port)
	}
	log.Print(response)
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
		log.Println("Error parsing connect message", err)
		NewMessage(NO).Send(conn)
		return
	}

	log.Printf("Received request to connect from %s:%s\n", data.Address, data.Port)
	ctx.AddNewNode(data.Address, data.Port)

	NewMessage(OK).Send(conn)
}

func newMsg(conn net.Conn, body []byte) {
	// TODO: Expects a type
	sendString(conn, NotYetImplementedError())
}
