package protocol

import (
	"net"
)

func ConnectTo(ownAddress string, ownPort string, targetAddress string, targetPort string) (conn net.Conn, ok bool) {
	conn, err := net.Dial("tcp", targetAddress+":"+targetPort)
	if err != nil {
		ok = false
		return
	}

	res, err := NewMessage(Q_CONNECT).AddContent(struct {
		Address string `json:"address"`
		Port    string `json:"port"`
	}{ownAddress, ownPort}).SendAwaitRead(conn)
	if err != nil {
		conn.Close()
		ok = false
		return
	}

	msg_parsed, ok := MessageFromPayload(res)
	if ok && msg_parsed.header == OK {
		return conn, true
	} else {
		conn.Close()
		ok = false
		return
	}

}
