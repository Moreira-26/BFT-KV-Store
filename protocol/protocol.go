package protocol

import (
	"bufio"
	"fmt"
	"net"
)

const (
	PORT = "8089"
)

func handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)

	var msg []byte

	data := make([]byte, 64)

	for {
		sz, err := reader.Read(data)

		if err != nil {
			fmt.Println(err)
			return
		}

		for i := range sz {
			msg = append(msg, data[i])
		}

		// Check if there is any more data to read
		if reader.Buffered() == 0 {
			break
		}
	}

	// TODO: Response
	conn.Write([]byte("Hello World"))

	conn.Close()
}

func Start() {
	// TODO: How nodes will communicate with each other

	ln, err := net.Listen("tcp", ":"+PORT)

	if err != nil {
		// handle error
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		go handleConnection(conn)
	}
}
