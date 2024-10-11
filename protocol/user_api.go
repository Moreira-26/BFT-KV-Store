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
		Type crdts.CRDT_TYPE `json:"type"`
	}

	var data newMsgBody
	err := json.Unmarshal(body, &data)
	if err != nil {
		log.Println("Error parsing new crdt message", err)
		NewMessage(NO).Send(conn)
		return
	}

	crdtType := data.Type

	op, opId, err := crdts.NewCRDT(crdtType, ctx.Secretkey)
	if err != nil {
		log.Println("Failed to create new", crdtType, "operation", err)
		NewMessage(NO).Send(conn)
		return
	}

	assignErr := ctx.Storage.Assign(hex.EncodeToString(opId), op)
	if assignErr != nil {
		log.Println("Failed to store new", crdtType, "operation", err)
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
		log.Println("Error getting item from key:", err)
		NewMessage(NO).Send(conn)
		return
	}

	msg, err := NewMessage(OK).AddContent(struct {
		Key   string          `json:"key"`
		Value interface{}     `json:"value"`
		Type  crdts.CRDT_TYPE `json:"type"`
	}{Key: data.Key, Value: resultObject.Value, Type: resultObject.Type})
	if err != nil {
		log.Println(err)
	} else {
		msg.Send(conn)
	}
}

func storeOperation(ctx *context.AppContext, conn net.Conn, key string, op []byte) bool {
	if err := ctx.Storage.Append(key, op); err != nil {
		log.Println("Failed to create operation on key", key, "due to:", err)
		return false
	} else {
		return true
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

	invalid := false
	var incOp []byte

	switch resultObject.Type {
	case crdts.CRDT_COUNTER:
		incOp, err = crdts.IncCounterOp(ctx.Secretkey, data.Value, resultObject.Heads)
	default:
		invalid = true
	}

	if !invalid && err == nil && storeOperation(ctx, conn, data.Key, incOp) {
		NewMessage(OK).Send(conn)
	} else {
		NewMessage(NO).Send(conn)
	}
}

func decMsg(ctx *context.AppContext, conn net.Conn, body []byte) {
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

	invalid := false
	var decOp []byte

	switch resultObject.Type {
	case crdts.CRDT_COUNTER:
		decOp, err = crdts.DecCounterOp(ctx.Secretkey, data.Value, resultObject.Heads)
	default:
		invalid = true
	}

	if !invalid && err == nil && storeOperation(ctx, conn, data.Key, decOp) {
		NewMessage(OK).Send(conn)
	} else {
		NewMessage(NO).Send(conn)
	}
}

func addMsg(ctx *context.AppContext, conn net.Conn, body []byte) {
	type readMsgBody struct {
		Key   string `json:"key"`
		Value any    `json:"value"`
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

	invalid := false
	var addOp []byte

	switch resultObject.Type {
	case crdts.CRDT_GSET:
		addOp, err = crdts.AddGSetOp(ctx.Secretkey, data.Value, resultObject.Heads)
	default:
		invalid = true
	}

	if !invalid && err == nil && storeOperation(ctx, conn, data.Key, addOp) {
		NewMessage(OK).Send(conn)
	} else {
		NewMessage(NO).Send(conn)
	}
}
