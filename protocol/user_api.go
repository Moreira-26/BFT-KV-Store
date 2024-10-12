package protocol

import (
	"bftkvstore/context"
	"bftkvstore/crdts"
	"bftkvstore/logger"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
)

func newMsg(ctx *context.AppContext, conn net.Conn, body []byte) {
	type newMsgBody struct {
		Type crdts.CRDT_TYPE `json:"type"`
	}

	data, err := unmarshallJson[newMsgBody](body)
	if err != nil {
		NewMessage(NO).Send(conn)
		return
	}

	crdtType := data.Type

	op, opId, err := crdts.NewCRDT(crdtType, ctx.Secretkey)
	if err != nil {
		logger.Error("Failed to create new", crdtType, "operation", err)
		NewMessage(NO).Send(conn)
		return
	}

	assignErr := ctx.Storage.Assign(hex.EncodeToString(opId), op)
	if assignErr != nil {
		logger.Error("Failed to store new", crdtType, "operation", err)
		NewMessage(NO).Send(conn)
		return
	}

	err = NewMessage(OK).AddContent(struct {
		Key string `json:"key"`
	}{Key: hex.EncodeToString(opId[:])}).Send(conn)
	if err != nil {
		logger.Error(err)
	}
}

func readMsg(ctx *context.AppContext, conn net.Conn, body []byte) {
	type readMsgBody struct {
		Key string `json:"key"`
	}
	data, err := unmarshallJson[readMsgBody](body)
	if err != nil {
		NewMessage(NO).Send(conn)
		return
	}

	resultObject, err := ctx.Storage.Get(data.Key)
	if err != nil {
		logger.Alert("Error getting item from key: ", err)
		NewMessage(NO).Send(conn)
		return
	}

	err = NewMessage(OK).AddContent(struct {
		Key   string          `json:"key"`
		Value interface{}     `json:"value"`
		Type  crdts.CRDT_TYPE `json:"type"`
	}{
		Key:   data.Key,
		Value: resultObject.Value,
		Type:  resultObject.Type,
	}).Send(conn)
	if err != nil {
		logger.Error(err)
	}
}

func opMsg(opType MessageHeader, ctx *context.AppContext, conn net.Conn, body []byte) {
	type readMsgBody struct {
		Key   string `json:"key"`
		Value any    `json:"value"`
	}

	data, err := unmarshallJson[readMsgBody](body)
	if err != nil {
		NewMessage(NO).Send(conn)
		return
	}

	resultObject, err := ctx.Storage.Get(data.Key)
	if err != nil {
		logger.Alert("Error getting item from key: ", err)
		NewMessage(NO).Send(conn)
		return
	}

	op, err := getOperation(opType, resultObject.Type, ctx.Secretkey, data.Value, resultObject.Heads)
	if err == nil && storeOperation(ctx, conn, data.Key, op) {
		NewMessage(OK).Send(conn)
	} else {
		logger.Alert(err, data.Value)
		NewMessage(NO).Send(conn)
	}
}

func getOperation(opType MessageHeader, crdtType crdts.CRDT_TYPE, secretkey ed25519.PrivateKey, value interface{}, heads []crdts.SignedOperation) ([]byte, error) {
	switch crdtType {
	case crdts.CRDT_COUNTER:
		switch opType {
		case API_INC:
			if v, ok := value.(float64); ok == true {
				return crdts.IncCounterOp(secretkey, int(math.Round(v)), heads)
			} else {
				goto wrongValueType
			}
		case API_DEC:
			if v, ok := value.(float64); ok == true {
				return crdts.DecCounterOp(secretkey, int(math.Round(v)), heads)
			} else {
				goto wrongValueType
			}
		default:
			goto invalid
		}
	case crdts.CRDT_GSET:
		switch opType {
		case API_ADD:
			return crdts.AddGSetOp(secretkey, value, heads)
		default:
			goto invalid
		}
	case crdts.CRDT_2PSET:
		switch opType {
		case API_ADD:
			return crdts.AddTwoPhaseSetOp(secretkey, value, heads)
		case API_RMV:
			return crdts.RemoveTwoPhaseSetOp(secretkey, value, heads)
		default:
			goto invalid
		}
	default:
		goto invalid
	}

invalid:
	return []byte{}, errors.New(fmt.Sprint("No operation of type", opType, "exists for the CRDT", crdtType))

wrongValueType:
	return []byte{}, errors.New("Provided the wrong value type for the operation")
}

func unmarshallJson[T interface{}](body []byte) (T, error) {
	var data T
	err := json.Unmarshal(body, &data)
	if err != nil {
		logger.Error("Error parsing message json", err)
	}
	return data, err
}

func storeOperation(ctx *context.AppContext, conn net.Conn, key string, op []byte) bool {
	if err := ctx.Storage.Append(key, op); err != nil {
		logger.Error("Failed to create operation on key", key, "due to:", err)
		return false
	} else {
		return true
	}
}
