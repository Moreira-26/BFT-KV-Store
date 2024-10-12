package protocol

import (
	"bftkvstore/context"
	"net"
)

func Router(ctx *context.AppContext, conn net.Conn, msg Message) {
	switch msg.header {
	// server api
	case PING:
		pingMsg(conn)
	case CONNECT:
		connectMsg(ctx, conn, msg.content)
	case Q_CONNECT:
		qConnectMsg(ctx, conn, msg.content)

	// user api
	case API_NEW:
		newMsg(ctx, conn, msg.content)
	case API_GET:
		readMsg(ctx, conn, msg.content)
	case API_INC, API_DEC, API_ADD, API_RMV:
		opMsg(msg.header, ctx, conn, msg.content)
	default:
	}
}
