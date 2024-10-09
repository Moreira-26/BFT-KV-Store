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

func newCounterOp(secretkey ed25519.PrivateKey, val int) ([]byte, error) {
	return SignOp(secretkey, MyOperation{
		Op:    "new",
		Preds: make([]string, 0),
		Crdt:  MyCounter{Val: val},
		Type:  "counter",
	})
}

func incCounterOp(secretkey ed25519.PrivateKey, val int, preds []signedOperation) ([]byte, error) {
	var hashed_preds []string = make([]string, len(preds))
	for idx, pred := range preds {
		hashed_preds[idx] = hashOperation(pred)
	}

	return SignOp(secretkey, MyOperation{
		Op:    "inc",
		Preds: hashed_preds,
		Crdt:  MyCounter{Val: val},
		Type:  "counter",
	})
}

func decCounterOp(secretkey ed25519.PrivateKey, val int, preds []signedOperation) ([]byte, error) {
	var hashed_preds []string = make([]string, len(preds))
	for idx, pred := range preds {
		hashed_preds[idx] = hashOperation(pred)
	}

	return SignOp(secretkey, MyOperation{
		Op:    "dec",
		Preds: hashed_preds,
		Crdt:  MyCounter{Val: val},
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
	tier  int
}

type operationCalculationResult struct {
	heads        []string
	value        interface{}
	predsMissing []string
}

func calculateOperations(signedops []signedOperation, crdtType string) operationCalculationResult {
	validOperations := make(map[string]MyOperation)
	for _, signedop := range signedops {
		readOp, err := readOperation(signedop)

		if readOp.Type != crdtType {
			fmt.Println("Found operation of wrong type when calculating")
			continue
		}

		if err == nil {
			validOperations[hashOperation(signedop)] = readOp
		}
	}

	predecessorsMissing := make([]string, 0)

	// This loop removes every operation that has an invalid op as predecessor
	iteration := 0
	for {
		iteration += 1
		keysToDelete := make(map[string]bool)

		for key, validOpVal := range validOperations {
			for _, pred := range validOpVal.Preds {
				_, exists := validOperations[pred]

				if !exists {
					fmt.Println("Invalid predecessor, deleting", key)
					if iteration == 1 {
						predecessorsMissing = append(predecessorsMissing, pred)
					}
					keysToDelete[key] = true
				}
			}
		}

		if len(keysToDelete) == 0 {
			break
		} else {
			for key, _ := range keysToDelete {
				delete(validOperations, key)
			}
		}

	}

	var hashGraph map[string]GraphNode = make(map[string]GraphNode)

	for key, _ := range validOperations {
		_, exists := hashGraph[key]

		if !exists {
			propagNode(hashGraph, validOperations, key, 0)
		}
	}

	// nodes with tier 0 are the ones that are the most recent ones
	var heads []string = make([]string, 0)
	var value float64 = 0

	for k, v := range hashGraph {
		if v.tier == 0 {
			heads = append(heads, k)
		}

		switch crdtType {
		case "counter":
			{
				if v.value.Op == "new" || v.value.Op == "inc" {
					value += v.value.Crdt.(map[string]interface{})["Val"].(float64)
				} else if v.value.Op == "dec" {
					value -= v.value.Crdt.(map[string]interface{})["Val"].(float64)
				}
			}
		}
	}

	return operationCalculationResult{heads: heads, value: value, predsMissing: predecessorsMissing}
}

func propagNode(graph map[string]GraphNode, validOps map[string]MyOperation, key string, tier int) {
	gNode, exists := graph[key]

	if !exists {
		graph[key] = GraphNode{
			value: validOps[key],
			tier:  tier,
		}
	} else {
		if tier > gNode.tier {
			gNode.tier = tier
			graph[key] = gNode
		}
	}

	for _, predKey := range validOps[key].Preds {
		propagNode(graph, validOps, predKey, tier+1)

		predGNode, _ := graph[predKey]

		predGNode.succs = append(graph[predKey].succs, key)
		graph[predKey] = predGNode

		gNode, _ = graph[key]
		gNode.preds = append(gNode.preds, predKey)
		graph[key] = gNode
	}
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

	op0, err := newCounterOp(keys["john"], 0)
	op1, err := incCounterOp(keys["alice"], 4, []signedOperation{op0})
	op2, err := decCounterOp(keys["john"], 3, []signedOperation{op0})
	op3, err := incCounterOp(keys["alice"], 5, []signedOperation{op1, op2})
	op4, err := incCounterOp(keys["alice"], 7, []signedOperation{op3})
	op5, err := incCounterOp(keys["john"], 1, []signedOperation{op4})
	op6, err := decCounterOp(keys["john"], 3, []signedOperation{op4})
	checkErr(t, err)

	result := calculateOperations([]signedOperation{op0, op2, op1, op4, op3, op6, op5, op6}, "counter")

	fmt.Println("heads:", result.heads)
	fmt.Println("value:", result.value)
	fmt.Println("predecessors missing:", result.predsMissing)

	t.Fail()
}

func checkErr(t *testing.T, err error) {
	if err != nil {
		fmt.Println("ERROR:", err)
		t.Fail()
	}
}
