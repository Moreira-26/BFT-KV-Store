package protocol

import (
	"encoding/json"
	"log"
	"net"
)

func newMsg(conn net.Conn, body []byte) {
	// TODO: Expects a type
	type newMsgBody struct {
		Type string `json:"type"`
	}

	var data newMsgBody

	err := json.Unmarshal(body, &data)
	if err != nil {
		log.Println("Error parsing connect message", err)
		NewMessage(NO).Send(conn)
		return
	}

	msg, err := NewMessage(OK).AddContent(struct {
		Key string `json:"key"`
	}{"TO BE IMPLEMENTED"})

	if err != nil {
		log.Println(err)
	} else {
		msg.Send(conn)
	}

	sendString(conn, NotYetImplementedError())
}
