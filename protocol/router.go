package protocol

import (
	"bftkvstore/context"
	"bftkvstore/logger"
	"net"
)

func Router(ctx *context.AppContext, conn net.Conn, msg Message) {
	closeConnection := true

	switch msg.header {
	// server api
	case PING:
		pingMsg(conn)
	case CONNECT:
		connectMsg(ctx, conn, msg.content)
	case Q_CONNECT:
		closeConnection = !qConnectMsg(ctx, conn, msg.content)

	// user api
	case API_NEW:
		newMsg(ctx, conn, msg.content)
	case API_GET:
		readMsg(ctx, conn, msg.content)
	case API_INC, API_DEC, API_ADD, API_RMV:
		opMsg(msg.header, ctx, conn, msg.content)
	default:
	}

	if closeConnection {
		if err := conn.Close(); err != nil {
			logger.Alert("Failed to close a connection", err)
		}
	}
}
