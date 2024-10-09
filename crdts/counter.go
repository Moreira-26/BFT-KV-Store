package crdts

import "crypto/ed25519"

const (
	CRDT_COUNTER = "counter"
)

type Counter struct {
	value int
}

func NewCounterOp(secretkey ed25519.PrivateKey, val int) ([]byte, error) {
	return SignOperation(secretkey, Operation{
		Op:    "new",
		Preds: make([]string, 0),
		Crdt:  Counter{value: val},
		Type:  CRDT_COUNTER,
	})
}

func IncCounterOp(secretkey ed25519.PrivateKey, val int, preds []SignedOperation) ([]byte, error) {
	var hashed_preds []string = make([]string, len(preds))
	for idx, pred := range preds {
		hashed_preds[idx] = HashOperation(pred)
	}

	return SignOperation(secretkey, Operation{
		Op:    "inc",
		Preds: hashed_preds,
		Crdt:  Counter{value: val},
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
		Crdt:  Counter{value: val},
		Type: CRDT_COUNTER,
	})
}
