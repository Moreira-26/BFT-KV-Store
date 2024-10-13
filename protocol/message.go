package protocol

import (
	"bftkvstore/logger"
	"encoding/json"
	"errors"
	"net"
)

type MessageHeader string

const (
	// server api
	PING      MessageHeader = "PING" // Just a PING message
	PONG      MessageHeader = "PONG" // Just a PONG message
	CONNECT   MessageHeader = "CONN" // Expects 2 arguments -> address and port
	Q_CONNECT MessageHeader = "CON?" // Asks a node if he wants to connect
	OK        MessageHeader = "R_OK"
	NO        MessageHeader = "R_NO"
	ERR		  MessageHeader = "R_ER"
	MSGS      MessageHeader = "MSGS"
	NEEDS     MessageHeader = "NEED"

	// user api
	API_NEW MessageHeader = "/new" // Adds a new key to the database, expects a type
	API_GET MessageHeader = "/get" // Gets the value to a key in the database
	API_INC MessageHeader = "/inc" // Increments a value in the database
	API_DEC MessageHeader = "/dec" // Decrements a value in the database
	API_ADD MessageHeader = "/add" // Adds a value to the database
	API_RMV MessageHeader = "/rmv" // Removes a value from the database
)

var EMPTYBODY struct{} = struct{}{}

type Message struct {
	header    MessageHeader
	content   []byte
	malformed error
}

func NewMessage(header MessageHeader) Message {
	return Message{
		header,
		make([]byte, 0),
		nil,
	}
}

func MessageFromPayload(payload []byte) (msg Message, ok bool) {
	if len(payload) < 4 {
		msg.malformed = errors.New("Failed to parse payload into message due to the payload being too short")
		logger.Alert(msg.malformed)
		return msg, false
	}

	var header = MessageHeader(payload[:4])
	var content = payload[4:]

	return Message{
		header:    header,
		content:   content,
		malformed: nil,
	}, true
}

func (msg Message) AddContent(content interface{}) Message {
	if serialized, err := json.Marshal(content); err != nil {
		msg.malformed = errors.New("Failed to serialize content")
	} else {
		msg.content = serialized
	}

	return msg
}

func (msg Message) IsMalformed() bool {
	return msg.malformed != nil
}

func (msg Message) Send(conn net.Conn) (e error) {
	if msg.IsMalformed() {
		return msg.malformed
	}

	var payload []byte = []byte(msg.header)

	payload = append(payload, msg.content...)

	_, e = conn.Write(payload)

	return e
}

func (msg Message) SendAwaitRead(conn net.Conn) (data []byte, e error) {
	if e := msg.Send(conn); e != nil {
		return nil, e
	}

	return ReadFromConnection(conn)
}
