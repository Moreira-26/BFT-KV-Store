package crdts

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

type CRDT_TYPE string

const (
	CRDT_COUNTER CRDT_TYPE = "counter"
	CRDT_GSET    CRDT_TYPE = "gset"
	CRDT_2PSET   CRDT_TYPE = "2pset"
)

func isValidCrdtType(crdtType CRDT_TYPE) bool {
	switch crdtType {
	case CRDT_COUNTER, CRDT_GSET, CRDT_2PSET:
		return true
	}

	return false
}

func genRandomToken(size int) (string, error) {
	bytes := make([]byte, size)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func createCrdtOp(crdtType CRDT_TYPE, secretkey ed25519.PrivateKey) (op []byte, id []byte, err error) {
	// create nonce
	var nonce string
	nonce, err = genRandomToken(8)
	if err != nil {
		return
	}

	op, err = SignOperation(secretkey, Operation{
		Op:    "new",
		Preds: make([]string, 0),
		Crdt:  nil,
		Nonce: nonce,
		Type:  crdtType,
	})

	if err != nil {
		return
	}

	opId := sha256.Sum256(op)
	return op, opId[:], nil
}

func NewCRDT(crdtType CRDT_TYPE, secretkey ed25519.PrivateKey) (op []byte, id []byte, err error) {
	if isValidCrdtType(crdtType) {
		op, id, err = createCrdtOp(crdtType, secretkey)
	} else {
		err = errors.New(fmt.Sprint("There is no crdt of type", crdtType))
	}

	return
}
