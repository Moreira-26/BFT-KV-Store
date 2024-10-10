package protocol

import (
	"bftkvstore/context"
	"bftkvstore/crdts"
	"encoding/hex"
	"encoding/json"
	"log"
	"net"
)

func newMsg(ctx *context.AppContext, conn net.Conn, body []byte) {
	type newMsgBody struct {
		Type string `json:"type"`
	}

	var data newMsgBody
	err := json.Unmarshal(body, &data)
	if err != nil {
		log.Println("Error parsing new crdt message", err)
		NewMessage(NO).Send(conn)
		return
	}

	if data.Type == crdts.CRDT_COUNTER {
		op, opId, err := crdts.NewCounterOp(ctx.Secretkey, 0)
		if err != nil {
			log.Println("Failed to create new counter operation", err)
			NewMessage(NO).Send(conn)
			return
		}

		assignErr := ctx.Storage.Assign(hex.EncodeToString(opId), op)
		if assignErr != nil {
			log.Println("Failed to store new counter operation", err)
			NewMessage(NO).Send(conn)
			return
		}

		msg, err := NewMessage(OK).AddContent(struct {
			Key string `json:"key"`
		}{Key: hex.EncodeToString(opId[:])})
		if err != nil {
			log.Println(err)
		} else {
			msg.Send(conn)
		}
	}
}

func readMsg(ctx *context.AppContext, conn net.Conn, body []byte) {
	type readMsgBody struct {
		Key string `json:"key"`
	}

	var data readMsgBody
	err := json.Unmarshal(body, &data)
	if err != nil {
		log.Println("Error parsing read message", err)
		NewMessage(NO).Send(conn)
		return
	}

	resultObject, err := ctx.Storage.Get(data.Key)
	if err != nil {
		log.Println("Error getting item from key", err)
		NewMessage(NO).Send(conn)
		return
	}

	msg, err := NewMessage(OK).AddContent(struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
		Type  string      `json:"type"`
	}{Key: data.Key, Value: resultObject.Value, Type: "counter"})
	if err != nil {
		log.Println(err)
	} else {
		msg.Send(conn)
	}
}

func incMsg(ctx *context.AppContext, conn net.Conn, body []byte) {
	type readMsgBody struct {
		Key   string `json:"key"`
		Value int    `json:"value"`
	}

	var data readMsgBody
	err := json.Unmarshal(body, &data)
	if err != nil {
		log.Println("Error parsing read message", err)
		NewMessage(NO).Send(conn)
		return
	}

	resultObject, err := ctx.Storage.Get(data.Key)
	if err != nil {
		log.Println("Error getting item from key", err)
		NewMessage(NO).Send(conn)
		return
	}

	switch resultObject.Type {
	case crdts.CRDT_COUNTER:
		failed := false

		incOp, err := crdts.IncCounterOp(ctx.Secretkey, data.Value, resultObject.Heads)
		if err != nil {
			failed = true
		}

		err = ctx.Storage.Append(data.Key, incOp)
		if err != nil {
			failed = true
		} else {
			NewMessage(OK).Send(conn)
		}

		if failed {
			log.Println("Failed to create counter inc operation on key", data.Key)
		} else {
			return
		}
	}

	NewMessage(NO).Send(conn)
}
