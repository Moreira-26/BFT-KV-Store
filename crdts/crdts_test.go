package crdts

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

type MyCounter struct {
	Val int
}

type MyOperation struct {
	Op    string
	Preds []string
	Crdt  interface{}
	Type  string
}

func hashOperation(op []byte) string {
	hash := sha256.Sum256(op)
	return hex.EncodeToString(hash[:])
}

func newCounterOp(secretkey ed25519.PrivateKey, crdt MyCounter) ([]byte, error) {
	return SignOp(secretkey, MyOperation{
		Op:    "new",
		Preds: make([]string, 0),
		Crdt:  crdt,
		Type:  "counter",
	})
}

func modifyCounterOp(secretkey ed25519.PrivateKey, crdt MyCounter, preds []signedOperation) ([]byte, error) {
	var hashed_preds []string = make([]string, len(preds))
	for idx, pred := range preds {
		hashed_preds[idx] = hashOperation(pred)
	}

	return SignOp(secretkey, MyOperation{
		Op:    "modify",
		Preds: hashed_preds,
		Crdt:  crdt,
		Type:  "counter",
	})
}

func readOperation(payload []byte) (op MyOperation, err error) {
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

func SignOp(secretkey ed25519.PrivateKey, operation MyOperation) ([]byte, error) {
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

type signedOperation []byte

type GraphNode struct {
	value MyOperation
	preds []string
	succs []string
}

func CalculateOperations(signedops []signedOperation) {
	validOperations := make(map[string]MyOperation)
	for _, signedop := range signedops {
		readOp, err := readOperation(signedop)
		if err == nil {
			validOperations[hashOperation(signedop)] = readOp
		}
	}

	// This loop removes every operation that has an invalid op as predecessor
	deleted := false
	for {
		deleted = false

		for key, validOpVal := range validOperations {
			for _, pred := range validOpVal.Preds {
				_, exists := validOperations[pred]

				if !exists {
					fmt.Println("A precedence does not exist, deleting", key)
					delete(validOperations, key)
					deleted = true
					break
				}
			}
		}

		if !deleted {
			break
		}
	}

	fmt.Println(len(validOperations))

	// Build Graph

	// Check what nodes are not fully connected

	// Remove every node after

	// check if every predecessor exists
}

func TestUtf8(t *testing.T) {
	var bytesmsg []byte = []byte{
		125, 170, 60, 45, 155, 136, 62, 192,
		188, 163, 87, 131, 178, 98, 21, 208,
		196, 70, 234, 122, 51, 5, 19, 123,
		199, 134, 150, 125, 119, 138, 149, 217,
		65, 21, 172, 178, 84, 236, 86, 153,
		188, 155, 12, 22, 61, 65, 252, 227,
		69, 146, 240, 64, 101, 139, 42, 64,
		73, 77, 77, 145, 136, 24, 5, 103,
	}

	var encodedbytesmsg string = hex.EncodeToString(bytesmsg)

	decodedbytesmsg, err := hex.DecodeString(encodedbytesmsg)
	if err != nil {
		t.FailNow()
	}

	if len(decodedbytesmsg) != len(bytesmsg) {
		t.FailNow()
	}

	equal := true
	for i := range len(bytesmsg) {
		equal = equal && decodedbytesmsg[i] == bytesmsg[i]
	}
	if !equal {
		t.FailNow()
	}

}

func TestHelloName(t *testing.T) {
	keys := make(map[string][]byte)
	for _, name := range [...]string{"john", "alice"} {
		_, keys[name], _ = ed25519.GenerateKey(rand.Reader)
	}

	var userACounter MyCounter = MyCounter{Val: 0}

	op0, err := newCounterOp(keys["john"], userACounter)
	op1, err := modifyCounterOp(keys["alice"], MyCounter{Val: 4}, []signedOperation{op0})
	op2, err := modifyCounterOp(keys["john"], MyCounter{Val: 3}, []signedOperation{op0})
	op3, err := modifyCounterOp(keys["alice"], MyCounter{Val: 5}, []signedOperation{op1, op2})
	op4, err := modifyCounterOp(keys["alice"], MyCounter{Val: 7}, []signedOperation{op3})
	op5, err := modifyCounterOp(keys["john"], MyCounter{Val: 1}, []signedOperation{op4})
	checkErr(t, err)

	CalculateOperations([]signedOperation{op0, op4, op1, op3, op5})

	t.Fail()
}

func checkErr(t *testing.T, err error) {
	if err != nil {
		fmt.Println("ERROR:", err)
		t.Fail()
	}
}
