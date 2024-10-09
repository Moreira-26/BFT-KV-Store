package crdts

import "crypto/ed25519"

type Counter struct {
	value int
}

func NewCounterOp(secretkey ed25519.PrivateKey, crdt Counter) ([]byte, error) {
	return SignOperation(secretkey, Operation{
		Op:    "new",
		Preds: make([]string, 0),
		Crdt:  crdt,
		Type:  "counter",
	})
}

func ModifyCounterOp(secretkey ed25519.PrivateKey, crdt Counter, preds []SignedOperation) ([]byte, error) {
	var hashed_preds []string = make([]string, len(preds))
	for idx, pred := range preds {
		hashed_preds[idx] = HashOperation(pred)
	}

	return SignOperation(secretkey, Operation{
		Op:    "modify",
		Preds: hashed_preds,
		Crdt:  crdt,
		Type:  "counter",
	})
}
