package protocol

import (
	"bftkvstore/logger"
	"fmt"
	"net"
)

func ConnectTo(ownAddress string, ownPort string, targetAddress string, targetPort string) bool {
	conn, err := net.Dial("tcp", targetAddress+":"+targetPort)

	if err != nil {
		logger.Alert(fmt.Sprintf("Failed to Dial %s:%s", targetAddress, targetPort))
		return false
	}

	res, err := NewMessage(Q_CONNECT).AddContent(struct {
		Address string `json:"address"`
		Port    string `json:"port"`
	}{ownAddress, ownPort}).SendAwaitRead(conn)
	if err != nil {
		logger.Alert(fmt.Sprintf("Failed to send Connect request to %s:%s", targetAddress, targetPort))
		return false
	}

	msg_parsed, ok := MessageFromPayload(res)

	return ok && msg_parsed.header == OK
}
