package crdts

import (
	"bftkvstore/logger"
	"bftkvstore/set"
	"bftkvstore/utils"
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
	Type  CRDT_TYPE
	Nonce string
}

type SignedOperation []byte

func HashOperationFromString(opStr string) string {
	op, _ := hex.DecodeString(opStr)
	hash := sha256.Sum256(op)
	return hex.EncodeToString(hash[:])
}

func HashOperation(op SignedOperation) string {
	hash := sha256.Sum256(op)
	return hex.EncodeToString(hash[:])
}

func IsValid(payload []byte) bool {
	if len(payload) < 96 {
		return false
	}
	publickey := payload[:32]
	signature := payload[32:96]
	content := payload[96:]

	return ed25519.Verify(publickey, content, signature)
}

func ReadOperation(payload []byte) (op Operation, err error) {
	if !IsValid(payload) {
		return op, errors.New("The operation provided is not valid")
	}

	content := payload[96:]
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
	Type         CRDT_TYPE
}

// TODO: Check if the new operation exists
func CalculateOperations(signedops []SignedOperation, crdtType CRDT_TYPE) OpCalcResult {
	validOperationsMap := make(map[string]Operation)
	signedOperationsMap := make(map[string]SignedOperation)

	for _, signedop := range signedops {
		readOp, err := ReadOperation(signedop)

		if readOp.Type != crdtType {
			logger.Alert("Found operation of wrong type when calculating:", readOp.Type, "!=", crdtType, readOp)
			continue
		}

		if err == nil {
			validOperationsMap[HashOperation(signedop)] = readOp
			signedOperationsMap[HashOperation(signedop)] = signedop
		}
	}
	predecessorsMissing := make([]string, 0)

	// This loop removes every operation that has an invalid op as predecessor
	iteration := 0
	for {
		iteration += 1
		keysToDelete := make(map[string]bool)

		for key, validOpVal := range validOperationsMap {
			for _, pred := range validOpVal.Preds {
				_, exists := validOperationsMap[pred]

				if !exists {
					logger.Alert("Invalid predecessor, deleting", key)
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
				delete(validOperationsMap, key)
			}
		}
	}

	hashGraph := make(map[string]graphNode)

	for key, _ := range validOperationsMap {
		_, exists := hashGraph[key]

		if !exists {
			propagateNode(hashGraph, validOperationsMap, key, 0)
		}
	}

	// nodes with tier 0 are the most recent
	var heads []SignedOperation = make([]SignedOperation, 0)

	var reducer opReducerI
	switch crdtType {
	case CRDT_COUNTER:
		reducer = &counterReducer{result: 0}
	case CRDT_GSET:
		reducer = &gSetReducer{result: make(map[any]bool)}
	case CRDT_2PSET:
		reducer = &twoPhaseSetReducer{result: make(map[any]bool)}
	}

	for k, v := range hashGraph {
		if v.tier == 0 {
			heads = append(heads, signedOperationsMap[k])
		}

		reducer.add(v)
	}

	return OpCalcResult{Heads: heads, Value: reducer.value(), PredsMissing: predecessorsMissing, Type: crdtType}
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

type opReducerI interface {
	add(node graphNode)
	value() any
}

type counterReducer struct{ result float64 }

func (r *counterReducer) add(node graphNode) {
	switch node.value.Op {
	case "inc":
		r.result += node.value.Crdt.(map[string]interface{})["value"].(float64)
	case "dec":
		r.result -= node.value.Crdt.(map[string]interface{})["value"].(float64)
	}
}

type gSetReducer struct{ result map[any]bool }

func (r *gSetReducer) add(node graphNode) {
	if node.value.Op == "add" {
		val := node.value.Crdt.(map[string]interface{})["value"]
		r.result[val] = true
	}
}

type twoPhaseSetReducer struct{ result map[any]bool }

func (r *twoPhaseSetReducer) add(node graphNode) {
	switch node.value.Op {
	case "add":
		val := node.value.Crdt.(map[string]interface{})["value"]
		res, exists := r.result[val]
		if !exists {
			res = true
		}
		r.result[val] = true && res
	case "rmv":
		val := node.value.Crdt.(map[string]interface{})["value"]
		r.result[val] = false
	}
}

func (r *counterReducer) value() any {
	return r.result
}
func (r *gSetReducer) value() any {
	keys := make([]any, 0)
	for k, _ := range r.result {
		keys = append(keys, k)
	}
	return keys
}
func (r *twoPhaseSetReducer) value() any {
	keys := make([]any, 0)
	for k, v := range r.result {
		if v {
			keys = append(keys, k)
		}
	}
	return keys
}

func CalculateOperationsTopologicalOrder(ops []string) []string {
	opStringMap := make(map[string]string)
	signedops := utils.Map(ops, func(elem string) []byte {
		r, _ := hex.DecodeString(elem)
		opStringMap[HashOperation(r)] = elem
		return r
	})
	validOperationsMap := make(map[string]Operation)

	for _, signedop := range signedops {
		readOp, err := ReadOperation(signedop)

		if err == nil {
			validOperationsMap[HashOperation(signedop)] = readOp
		}
	}

	hashGraph := make(map[string]graphNode)
	keys := set.New[string]()

	for key, op := range validOperationsMap {
		keys = set.Add(keys, key)
		hashGraph[key] = graphNode{
			value: op,
			succs: set.New[string](),
			preds: set.New[string](),
			tier:  0,
		}
	}

	for key, op := range validOperationsMap {
		for _, pred := range op.Preds {
			if !set.Has(keys, pred) {
				continue
			}
			// add key to succs to pred node
			predN := hashGraph[pred]
			predN.succs = set.Add(predN.succs, key)
			hashGraph[pred] = predN

			// add key to succs to pred node
			succN := hashGraph[key]
			succN.preds = set.Add(succN.preds, pred)
			hashGraph[key] = succN
		}

	}

	res := utils.Map(opTopologicalSort(hashGraph), func(elem string) string {
		return opStringMap[elem]
	})

	if len(res) != len(keys) {
		logger.Fatal("Something bad happened during topological sort!", len(res), len(keys), len(hashGraph))
	}

	return res
}

func opTopologicalSort(hashGraph map[string]graphNode) []string {
	keys := set.New[string]()
	for k, _ := range hashGraph {
		keys = set.Add(keys, k)
	}

	// add vertices with no incoming edge to queue Q
	queue := utils.Filter(keys, func(elem string) bool {
		val, _ := hashGraph[elem]
		return len(val.preds) == 0
	})

	list := make([]string, 0)

	for len(queue) != 0 {
		u := queue[0]
		queue = queue[1:]
		list = append(list, u)

		toDel := make([]int, 0)
		uNode := hashGraph[u]
		for idx, succ := range uNode.succs {
			toDel = append(toDel, idx)

			succVal := hashGraph[succ]
			succVal.preds = set.Remove(succVal.preds, u)
			hashGraph[succ] = succVal

			// NOTE: This might be unnecessary
			uVal := hashGraph[u]
			uVal.succs = set.Remove(uVal.succs, succ)
			hashGraph[u] = uVal

			if len(hashGraph[succ].preds) == 0 {
				queue = append(queue, succ)
			}
		}
	}

	return list
}
