package crdts

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
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

type graphNode struct {
	value Operation
	preds []string
	succs []string
	tier  int
}

type OpCalcResult struct {
	Heads        []SignedOperation
	Value        interface{}
	PredsMissing []string
	Type         string
}

// TODO: Check if the new operation exists
func CalculateOperations(signedops []SignedOperation, crdtType string) OpCalcResult {
	validOperations := make(map[string]Operation)
	signedOperations := make(map[string]SignedOperation)
	for _, signedop := range signedops {
		readOp, err := ReadOperation(signedop)

		if readOp.Type != crdtType {
			log.Println("Found operation of wrong type when calculating:", readOp.Type, "!=", crdtType, readOp)
			continue
		}

		if err == nil {
			validOperations[HashOperation(signedop)] = readOp
			signedOperations[HashOperation(signedop)] = signedop
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
					log.Println("Invalid predecessor, deleting", key)
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

	hashGraph := make(map[string]graphNode)

	for key, _ := range validOperations {
		_, exists := hashGraph[key]

		if !exists {
			propagateNode(hashGraph, validOperations, key, 0)
		}
	}

	// nodes with tier 0 are the most recent
	var heads []SignedOperation = make([]SignedOperation, 0)
	var value float64 = 0

	for k, v := range hashGraph {
		if v.tier == 0 {
			heads = append(heads, signedOperations[k])
		}

		switch crdtType {
		case CRDT_COUNTER:
			{
				if v.value.Op == "new" || v.value.Op == "inc" {
					value += v.value.Crdt.(map[string]interface{})["value"].(float64)
				} else if v.value.Op == "dec" {
					value -= v.value.Crdt.(map[string]interface{})["value"].(float64)
				}
			}
		}
	}

	return OpCalcResult{Heads: heads, Value: value, PredsMissing: predecessorsMissing, Type: crdtType}
}

func propagateNode(graph map[string]graphNode, validOps map[string]Operation, key string, tier int) {
	gNode, exists := graph[key]

	if !exists {
		graph[key] = graphNode{
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
		propagateNode(graph, validOps, predKey, tier+1)

		predGNode, _ := graph[predKey]

		predGNode.succs = append(graph[predKey].succs, key)
		graph[predKey] = predGNode

		gNode, _ = graph[key]
		gNode.preds = append(gNode.preds, predKey)
		graph[key] = gNode
	}
}
