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
		newMsg(conn, msg.Content)
	default:
		// SendString(conn, CmdNotFoundError(header))
	}
}

