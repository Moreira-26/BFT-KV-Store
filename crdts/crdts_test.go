package crdts

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"
)

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

func TestCounterOperations(t *testing.T) {
	keys := make(map[string][]byte)
	for _, name := range [...]string{"john", "alice"} {
		_, keys[name], _ = ed25519.GenerateKey(rand.Reader)
	}

	op0, _, err := NewCounterOp(keys["john"], 0)
	op1, err := IncCounterOp(keys["alice"], 4, []SignedOperation{op0})
	op2, err := DecCounterOp(keys["john"], 3, []SignedOperation{op0})
	op3, err := IncCounterOp(keys["alice"], 5, []SignedOperation{op1, op2})
	op4, err := IncCounterOp(keys["alice"], 7, []SignedOperation{op3})
	op5, err := IncCounterOp(keys["john"], 1, []SignedOperation{op4})
	op6, err := DecCounterOp(keys["john"], 3, []SignedOperation{op4})
	checkErr(t, err)

	result := CalculateOperations([]SignedOperation{op0, op2, op1, op4, op3, op6, op5, op6}, CRDT_COUNTER)

	fmt.Println("heads:", result.Heads)
	fmt.Println("value:", result.Value)
	fmt.Println("predecessors missing:", result.PredsMissing)

	t.Fail()
}

func checkErr(t *testing.T, err error) {
	if err != nil {
		fmt.Println("ERROR:", err)
		t.Fail()
	}
}
