package crdts

import (
	"crypto/ed25519"
)

func SignOperation(secretkey ed25519.PrivateKey, content []byte) []byte {
	var publickey ed25519.PublicKey = secretkey.Public().(ed25519.PublicKey)
	var signed_content []byte = ed25519.Sign(secretkey, content)

	var signed_op []byte = make([]byte, 0)
	signed_op = append(signed_op, publickey...)
	return append(signed_op, signed_content...)
}
