package protocol

import (
	"fmt"
	"log"
	"net"
)

const (
	// server api
	PING    = "PING" // Just a PING message
	CONNECT = "CONN" // Expects 2 arguments -> address and port
	// user api
	NEW = "/new" // Adds a new key to the database, expects a type
	INC = "/inc" // Increments a value in the database
)

func parseMsg(msg string, conn net.Conn) {
	var header = msg[:4]
	var body = func() string { return msg[4:] }

	switch header {
	case PING:
		pingMsg(conn)
	case NEW:
		newMsg(conn, body())
	case CONNECT:
		connectMsg(conn, body())
	}

	if conn.Close() != nil {
		log.Println("Failed to close a connection")
	}
}

func pingMsg(conn net.Conn) {
	conn.Write([]byte("PONG"))
}

func connectMsg(conn net.Conn, body string) {
	// TODO: Connect to another node
	fmt.Println(body)
	conn.Write([]byte("Not yet Implemented"))
}

func newMsg(conn net.Conn, body string) {
	// TODO: Expects a type
	conn.Write([]byte("Not yet Implemented"))
}
