package protocol

import (
	"encoding/json"
	"log"
	"net"
)

func sendString(conn net.Conn, msg string) (n int, e error) {
	return conn.Write([]byte(msg))
}

type MessageHeader string

const (
	// server api
	PING      MessageHeader = "PING" // Just a PING message
	PONG      MessageHeader = "PONG" // Just a PONG message
	CONNECT   MessageHeader = "CONN" // Expects 2 arguments -> address and port
	Q_CONNECT MessageHeader = "CON?" // Asks a node if he wants to connect
	OK        MessageHeader = "R_OK"
	NO        MessageHeader = "R_NO"

	// user api
	NEW MessageHeader = "/new" // Adds a new key to the database, expects a type
	INC MessageHeader = "/inc" // Increments a value in the database
)

var EMPTYBODY struct{} = struct{}{}

type Message struct {
	Header  MessageHeader
	Content []byte
}

func NewMessage(header MessageHeader) Message {
	return Message{
		header,
		make([]byte, 0),
	}
}

func MessageFromPayload(payload []byte) Message {
	var header = MessageHeader(payload[:4])
	var content = payload[4:]

	return Message{
		Header:  header,
		Content: content,
	}
}

func (msg Message) AddContent(content interface{}) (Message, error) {
	// serialize
	serialized, err := json.Marshal(content)
	if err != nil {
		log.Println("Failed to serialize content")
		return msg, err
	}

	msg.Content = serialized

	return msg, nil
}

func (msg Message) Send(conn net.Conn) (e error) {
	var payload []byte = []byte(msg.Header)

	payload = append(payload, msg.Content...)

	_, e = conn.Write(payload)

	return e
}

func (msg Message) SendAwaitRead(conn net.Conn) (data []byte, e error) {
	e = msg.Send(conn)

	if e != nil {
		return nil, e
	}

	return ReadFromConnection(conn)
}
