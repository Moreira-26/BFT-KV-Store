package protocol

import (
	"bftkvstore/context"
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
	case API_NEW:
		newMsg(ctx, conn, msg.Content)
	case API_GET:
		readMsg(ctx, conn, msg.Content)
	case API_INC:
		incMsg(ctx, conn, msg.Content)
	case API_DEC:
		decMsg(ctx, conn, msg.Content)
	case API_ADD:
		addMsg(ctx, conn, msg.Content)
	case API_RMV:
		rmvMsg(ctx, conn, msg.Content)
	default:
	}
}

