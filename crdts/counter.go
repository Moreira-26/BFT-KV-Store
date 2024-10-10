package crdts

import (
	"crypto/ed25519"
	"crypto/sha256"
)

const (
	CRDT_COUNTER = "counter"
)

type Counter struct {
	Value int `json:"value"`
}

func NewCounterOp(secretkey ed25519.PrivateKey, val int) (op []byte, id []byte, err error) {
	op, err = SignOperation(secretkey, Operation{
		Op:    "new",
		Preds: make([]string, 0),
		Crdt:  Counter{Value: val},
		Type:  CRDT_COUNTER,
	})
	if err != nil {
		return
	}

	opId := sha256.Sum256(op)
	return op, opId[:], nil
}

func IncCounterOp(secretkey ed25519.PrivateKey, val int, preds []SignedOperation) ([]byte, error) {
	var hashed_preds []string = make([]string, len(preds))
	for idx, pred := range preds {
		hashed_preds[idx] = HashOperation(pred)
	}

	return SignOperation(secretkey, Operation{
		Op:    "inc",
		Preds: hashed_preds,
		Crdt:  Counter{Value: val},
		Type:  CRDT_COUNTER,
	})
}

func DecCounterOp(secretkey ed25519.PrivateKey, val int, preds []SignedOperation) ([]byte, error) {
	var hashed_preds []string = make([]string, len(preds))
	for idx, pred := range preds {
		hashed_preds[idx] = HashOperation(pred)
	}

	return SignOperation(secretkey, Operation{
		Op:    "dec",
		Preds: hashed_preds,
		Crdt:  Counter{Value: val},
		Type:  CRDT_COUNTER,
	})
}
