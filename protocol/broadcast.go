package protocol

import (
	"bftkvstore/context"
	"bftkvstore/crdts"
	"bftkvstore/logger"
	"bftkvstore/set"
	"bftkvstore/utils"
	"encoding/hex"
	"fmt"
	"net"
)

type connectionVariables struct {
	sent    set.Set[string]
	recvd   set.Set[string]
	missing set.Set[string]
	mconn   set.Set[string]
}

type connectionData struct {
	conn net.Conn
	vars *connectionVariables
}

var connections map[string]*connectionData = make(map[string]*connectionData)
var M set.Set[string] = set.New[string]()

func BroadcastReceiver(ctx *context.AppContext) {
	for {
		// check for new nodes
		if len(ctx.NewNodes) > 0 && ctx.Lock.TryLock() {
			if len(ctx.NewNodes) > 0 {
				node := ctx.NewNodes[0]

				name := fmt.Sprintf("%s:%s", node.Address, node.Port)
				connections[name] = &connectionData{
					conn: node.Conn,
					vars: nil,
				}
				on_connection_to_another_replica(ctx, connections[name])
				ctx.NewNodes = ctx.NewNodes[1:]
			}
			ctx.Lock.Unlock()
		}
		for name, connData := range connections {
			if connData.conn == nil {
				// TODO: remove from connections
				continue
			}
			if payload, err := ReadFromConnection(connData.conn); err != nil {
				continue
			} else {
				if len(payload) < 4 {
					continue
				}

				header := MessageHeader(payload[:4])
				logger.Debug("from main:", header, "from:", name)

				switch header {
				case MSGS:
					on_receiving_msgs(ctx, connData, payload[4:])
				case NEEDS:
					on_receiving_needs(ctx, connData, payload[4:])
				}
			}
		}
	}
}

type msgsDTO struct {
	Key      string   `json:"key"`
	Messages []string `json:"messages"`
}

// Algorithm 1 A Byzantine causal broadcast algorithm.
func broadcast(ctx *context.AppContext, key string, m crdts.SignedOperation) {
	fmt.Println("Issued broadcast", key)

	// TODO: Add a lock here because it is atomic
	M = set.Add(M, hex.EncodeToString(m))

	msg := NewMessage(MSGS).AddContent(msgsDTO{
		Key:      key,
		Messages: []string{hex.EncodeToString(m)},
	})
	if msg.IsMalformed() { // should never happen
		logger.Fatal("Broadcast message is malformed")
		return
	}

	for _, connData := range connections {
		if connData.conn != nil {
			if err := msg.Send(connData.conn); err != nil {
				logger.Alert("Failed to send Message during broadcast", err)
			}
		}
	}
}

func on_connection_to_another_replica(ctx *context.AppContext, connData *connectionData) {
	// connection-local variables
	connData.vars = &connectionVariables{
		sent:    set.New[string](),
		recvd:   set.New[string](),
		missing: set.New[string](),
		mconn:   M,
	}

	// TODO: Send heads
	// send ⟨heads : heads(Mconn)⟩ via current connection
}

func on_receiving_heads() {
	/*
	 on receiving ⟨heads : hs⟩ via a connection do
	 HandleMissing({ℎ ∈ hs | 𝑚 ∈ Mconn. 𝐻 (𝑚) = ℎ})
	*/
}

func on_receiving_msgs(ctx *context.AppContext, connData *connectionData, body []byte) {
	data, err := unmarshallJson[msgsDTO](body)
	if err != nil {
		logger.Error("Failed to parse msgs JSON", err)
		return
	}

	signedOps := make([]crdts.Operation, 0)
	for _, msg := range data.Messages {
		msgBytes, err := hex.DecodeString(msg)
		if err != nil {
			logger.Alert("Failed decode the msgs operation", err)
			continue
		}

		// ReadOperation checks if it is valid
		signedOp, err := crdts.ReadOperation(msgBytes)
		if err != nil {
			logger.Alert("Failed read the msgs operation", err)
			continue
		}
		signedOps = append(signedOps, signedOp)
		connData.vars.recvd = set.Add(connData.vars.recvd, msg)
	}

	predsToCheck := set.New[string]()
	for _, signedOp := range signedOps {
		for _, pred := range signedOp.Preds {
			predsToCheck = set.Add(predsToCheck, pred)
		}
	}

	toHash := set.Union(connData.vars.mconn, connData.vars.recvd)
	hashes := utils.Map(toHash, crdts.HashOperationFromString)

	unresolved := set.Diff(predsToCheck, hashes)

	handleMissing(ctx, connData, data.Key, unresolved)
}

func on_receiving_needs(ctx *context.AppContext, connData *connectionData, body []byte) {
	data, err := unmarshallJson[msgsDTO](body)
	if err != nil {
		logger.Error("Failed to parse msgs JSON", err)
		return
	}
	key := data.Key
	needs := set.FromSlice(data.Messages)

	reply := utils.Filter(set.Diff(connData.vars.mconn, connData.vars.sent), func(el string) bool {
		op, _ := hex.DecodeString(el)
		if set.Has(needs, crdts.HashOperation(op)) {
			return true
		} else {
			return false
		}
	})

	connData.vars.sent = set.Union(connData.vars.sent, reply)

	if len(reply) != 0 { // otherwise it gets stuck in a loop
		NewMessage(MSGS).AddContent(
			msgsDTO{
				Key:      key,
				Messages: reply,
			}).Send(connData.conn)
	}
}

func handleMissing(ctx *context.AppContext, connData *connectionData, key string, hashes set.Set[string]) {
	connData.vars.missing = set.Diff(
		set.Union(set.FromSlice(connData.vars.missing), set.FromSlice(hashes)),
		set.FromSlice(utils.Map[string, string](connData.vars.recvd, crdts.HashOperationFromString)),
	)

	fmt.Println("missing", connData.vars.missing)
	if len(connData.vars.missing) == 0 {
		// TODO: Add lock here because it is atomic
		msgs := set.Diff(set.FromSlice(connData.vars.recvd), M)

		fmt.Println("msgs would be", utils.Map(set.Diff(set.FromSlice(connData.vars.recvd), M), func(element string) crdts.Operation {
			toString, _ := hex.DecodeString(element)
			op, _ := crdts.ReadOperation(toString)
			return op
		}))

		M = set.Union(M, connData.vars.recvd)

		// TODO: deliver all of the messages in msgs in topologically sorted order
		var orderedMsgs []string = crdts.CalculateOperationsTopologicalOrder(msgs)

		fmt.Println("after sort msgs are", utils.Map(orderedMsgs, func(element string) crdts.Operation {
			toString, _ := hex.DecodeString(element)
			op, _ := crdts.ReadOperation(toString)
			return op
		}))

		for _, msg := range orderedMsgs {
			signedOp, _ := hex.DecodeString(msg)
			op, _ := crdts.ReadOperation(signedOp)
			var err error
			if op.Op == "new" {
				err = ctx.Storage.Assign(key, signedOp)
			} else {
				err = ctx.Storage.Append(key, signedOp)
			}

			if err != nil {
				logger.Error("Could not append operation", op, "with key", key, "reason:", err)
			}
		}
	} else {
		logger.Debug("Sending needs with", connData.vars.missing)
		msg := NewMessage(NEEDS).AddContent(
			msgsDTO{
				Key:      key,
				Messages: connData.vars.missing,
			})

		msg.Send(connData.conn)
	}
}
