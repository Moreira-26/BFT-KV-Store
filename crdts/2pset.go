package crdts

import (
	"crypto/ed25519"
)

type TwoPhaseSet struct {
	Value any `json:"value"`
}

func AddTwoPhaseSetOp(secretkey ed25519.PrivateKey, val any, preds []SignedOperation) ([]byte, error) {
	var hashed_preds []string = make([]string, len(preds))
	for idx, pred := range preds {
		hashed_preds[idx] = HashOperation(pred)
	}

	return SignOperation(secretkey, Operation{
		Op:    "add",
		Preds: hashed_preds,
		Crdt:  GSet{Value: val},
		Type:  CRDT_2PSET,
	})
}

func RemoveTwoPhaseSetOp(secretkey ed25519.PrivateKey, val any, preds []SignedOperation) ([]byte, error) {
	var hashed_preds []string = make([]string, len(preds))
	for idx, pred := range preds {
		hashed_preds[idx] = HashOperation(pred)
	}

	return SignOperation(secretkey, Operation{
		Op:    "rmv",
		Preds: hashed_preds,
		Crdt:  GSet{Value: val},
		Type:  CRDT_2PSET,
	})
}
