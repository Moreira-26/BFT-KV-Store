package crdts

import (
	"crypto/ed25519"
)

type GSet struct {
	Value interface{} `json:"value"`
}

func IncGSetOp(secretkey ed25519.PrivateKey, val int, preds []SignedOperation) ([]byte, error) {
	var hashed_preds []string = make([]string, len(preds))
	for idx, pred := range preds {
		hashed_preds[idx] = HashOperation(pred)
	}

	return SignOperation(secretkey, Operation{
		Op:    "inc",
		Preds: hashed_preds,
		Crdt:  GSet{Value: val},
		Type:  CRDT_GSET,
	})
}

func DecGSetOp(secretkey ed25519.PrivateKey, val int, preds []SignedOperation) ([]byte, error) {
	var hashed_preds []string = make([]string, len(preds))
	for idx, pred := range preds {
		hashed_preds[idx] = HashOperation(pred)
	}

	return SignOperation(secretkey, Operation{
		Op:    "dec",
		Preds: hashed_preds,
		Crdt:  GSet{Value: val},
		Type:  CRDT_GSET,
	})
}
