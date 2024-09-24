package protocol

import (
	"bufio"
	"fmt"
	"net"
)

func handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)

	var msg []byte

	data := make([]byte, 1000)

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

	parseMsg(string(msg), conn)
}

func ReceiverStart(port string) {
	ln, err := net.Listen("tcp", ":"+port)

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
