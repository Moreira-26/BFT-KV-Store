package crdts

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
)

type Operation struct {
	Op    string
	Preds []string
	Crdt  interface{}
	Type  string
}

type SignedOperation []byte

func HashOperation(op []byte) string {
	hash := sha256.Sum256(op)
	return hex.EncodeToString(hash[:])
}

func ReadOperation(payload []byte) (op Operation, err error) {
	publickey := payload[:32]
	signature := payload[32:96]
	content := payload[96:]

	if !ed25519.Verify(publickey, content, signature) {
		return op, errors.New("Operation cannot be verified")
	}

	err = json.Unmarshal(content, &op)
	if err != nil {
		return op, err
	}

	return op, nil
}

func SignOperation(secretkey ed25519.PrivateKey, operation Operation) ([]byte, error) {
	opjson, err := json.Marshal(operation)

	if err != nil {
		return opjson, err
	}

	var publickey ed25519.PublicKey = secretkey.Public().(ed25519.PublicKey)
	var signature []byte = ed25519.Sign(secretkey, opjson)

	var signed_op []byte = make([]byte, 0)
	signed_op = append(signed_op, publickey...)
	signed_op = append(signed_op, signature...)
	return append(signed_op, opjson...), nil
}
