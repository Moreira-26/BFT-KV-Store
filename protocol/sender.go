package protocol

import (
	"log"
	"net"
)

func ConnectTo(ownAddress string, ownPort string, targetAddress string, targetPort string) bool {
	conn, err := net.Dial("tcp", targetAddress+":"+targetPort)

	if err != nil {
		log.Printf("Failed to Dial %s:%s\n", targetAddress, targetPort)
		return false
	}

	msg, err := NewMessage(Q_CONNECT).AddContent(struct {
		Address string `json:"address"`
		Port    string `json:"port"`
	}{ownAddress, ownPort})
	if err != nil {
		return false
	}

	res, err := msg.SendAwaitRead(conn)

	if err != nil {
		log.Printf("Failed to send Connect request to %s:%s\n", targetAddress, targetPort)
		return false
	}

	return MessageFromPayload(res).Header == OK
}
