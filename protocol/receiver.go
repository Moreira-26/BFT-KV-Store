package protocol

import (
	"bftkvstore/context"
	"bufio"
	"encoding/binary"
	"net"
)

const READER_SIZE = 1000

func ReadFromConnection(conn net.Conn) (msg []byte, err error) {
	reader := bufio.NewReader(conn)

	headerAndSize := make([]byte, 6)

	sz, err := reader.Read(headerAndSize)
	if err != nil || sz != 6 {
		return nil, err
		// NOTE: Perhaps discard the buffer
	}

	msg = append(msg, headerAndSize[0:4]...)

	contentSize := int(binary.BigEndian.Uint16(headerAndSize[4:]))

	for contentSize > 0 {
		data := make([]byte, min(READER_SIZE, contentSize))

		sz, err := reader.Read(data)
		contentSize -= sz

		if err != nil {
			return nil, err
		}

		for i := range sz {
			msg = append(msg, data[i])
		}

		// Check if there is any more data to read
		if reader.Buffered() == 0 {
			break
		}
	}

	return msg, nil
}

func handleConnection(ctx *context.AppContext, conn net.Conn) {
	payload, err := ReadFromConnection(conn)
	if err != nil {
		return
	}

	msg, ok := MessageFromPayload(payload)
	if ok {
		Router(ctx, conn, msg)
	}
}

func ReceiverStart(ctx *context.AppContext, port string) {
	ln, err := net.Listen("tcp", ":"+port)

	if err != nil {
		// handle error
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			continue
		}
		go handleConnection(ctx, conn)
	}
}
