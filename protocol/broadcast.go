package protocol

import (
	"bftkvstore/context"
	"bftkvstore/crdts"
	"bftkvstore/logger"
	"bftkvstore/set"
	"bftkvstore/utils"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"syscall"
	// "github.com/bits-and-blooms/bloom/v3"
)

type connectionVariables struct {
	sent    set.Set[string]
	recvd   set.Set[string]
	missing set.Set[string]
	mconn   set.Set[string]
	// oldHeads set.Set[string]
}

type connectionData struct {
	conn   net.Conn
	vars   *connectionVariables
	ch     chan []byte
	hangup chan bool
}

var connections map[string]*connectionData = make(map[string]*connectionData)
var M set.Set[string] = set.New[string]()

func isNetConnClosedErr(err error) bool {
	switch {
	case
		errors.Is(err, net.ErrClosed),
		errors.Is(err, io.EOF),
		errors.Is(err, syscall.EPIPE):
		return true
	default:
		return false
	}
}

func listenToConnection(connData *connectionData) {
	for {
		msg, err := ReadFromConnection(connData.conn)

		if err != nil {
			if isNetConnClosedErr(err) {
				logger.Alert("Connection Closed")
				connData.hangup <- true
				return
			} else {
				logger.Error("Error in listen", err)
				connData.hangup <- true
				return
			}
		}

		connData.ch <- msg
	}
}

func BroadcastReceiver(ctx *context.AppContext) {
	for {
		// check for new nodes
		if len(ctx.NewNodes) > 0 && ctx.Lock.TryLock() {
			if len(ctx.NewNodes) > 0 {
				node := ctx.NewNodes[0]

				name := fmt.Sprintf("%s:%s", node.Address, node.Port)

				val, exists := connections[name]

				if exists {
					val.conn = node.Conn
					val.ch = make(chan []byte)
					val.hangup = make(chan bool)
					connections[name] = val
				} else {
					connections[name] = &connectionData{
						conn:   node.Conn,
						vars:   nil,
						ch:     make(chan []byte),
						hangup: make(chan bool),
					}
				}
				go listenToConnection(connections[name])

				if !exists {
					onConnectionToAnotherReplica(ctx, connections[name])
				}

				ctx.NewNodes = ctx.NewNodes[1:]
			}
			ctx.Lock.Unlock()
		}
		for name, connData := range connections {
			select {
			case payload := <-connData.ch:
				if len(payload) < 4 {
					continue
				}

				header := MessageHeader(payload[:4])

				switch header {
				case MSGS:
					onReceivingMsgs(ctx, connData, payload[4:])
				case NEEDS:
					onReceivingNeeds(ctx, connData, payload[4:])
				case HEADS:
					onReceivingHeads(ctx, connData, payload[4:])
				}
			case <-connData.hangup:
				connData.conn.Close()
				delete(connections, name)
				continue
			default:
				continue
			}
		}
	}
}

type msgsDTO struct {
	Key      string   `json:"key"`
	Messages []string `json:"messages"`
}

func broadcast(ctx *context.AppContext, key string, m crdts.SignedOperation) {
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

func onConnectionToAnotherReplica(ctx *context.AppContext, connData *connectionData) {
	// connection-local variables
	connData.vars = &connectionVariables{
		sent:    set.New[string](),
		recvd:   set.New[string](),
		missing: set.New[string](),
		mconn:   M,
		// oldHeads: set.New[string](),
	}

	heads := ctx.Storage.GetHeads()

	for key, hds := range heads {
		NewMessage(HEADS).AddContent(msgsDTO{
			Key:      key,
			Messages: utils.Map(hds, crdts.HashOperation),
		}).Send(connData.conn)
	}
}

/*
	func onConnectionToAnotherReplicaAlg2(ctx *context.AppContext, connData *connectionData) {
		oldHeads := set.New[string]()
		if connData.vars != nil {
			oldHeads = connData.vars.oldHeads
		}
		// connection-local variables
		connData.vars = &connectionVariables{
			sent:     set.New[string](),
			recvd:    set.New[string](),
			missing:  set.New[string](),
			oldHeads: oldHeads,
			mconn:    M,
		}

		// NOTE: I don't know what these parameters are, only that the first is the number of bytes
		bloomFilter := bloom.New(1000, 4)
		for _, head := range oldHeads {
			headBytes, _ := hex.DecodeString(head)
			bloomFilter.Add(headBytes)
		}

		// replace with get heads from Mconn
		heads := utils.Map(ctx.Storage.GetHeads(), hex.EncodeToString)
		filter := bloomFilter.BitSet().Bytes()

		// send ⟨heads : heads(Mconn), oldHeads : oldHeads, filter : filter

		for key, hds := range heads {
			NewMessage(HEADS).AddContent(struct {
				Heads    []string
				OldHeads []string
				Filter   []uint64
			}{
				Heads: heads,
				Messages: utils.Map(hds, func(el crdts.SignedOperation) string {
					return crdts.HashOperation(el)
				}),
			}).Send(connData.conn)
		}
	}

	func messagesSince(connData *connectionData, oldHeads []string) set.Set[string] {
		oldHeadsHashes := set.FromSlice(utils.Map(oldHeads, crdts.HashOperationFromString))
		known := utils.Filter(connData.vars.mconn, func(elem string) bool {
			return set.Has(oldHeadsHashes, crdts.HashOperationFromString(elem))
		})

		mConnHashMap := make(map[string]string)

		utils.Map(connData.vars.mconn, func(elem string) bool {
			mConnHashMap[crdts.HashOperationFromString(elem)] = elem
			return true
		})

		predsOfKnown := set.New[string]()
		predsOfKnownQueue := known

		for len(predsOfKnownQueue) > 0 {
			predToCheck := predsOfKnown[0]
			predsOfKnown = predsOfKnown[1:]

			opBytes, _ := hex.DecodeString(predToCheck)
			op, err := crdts.ReadOperation(opBytes)
			if err != nil {
				continue
			}

			for _, predHash := range op.Preds {
				val, exists := mConnHashMap[predHash]

				if exists {
					predsOfKnownQueue = append(predsOfKnownQueue, val)
				}
			}

			set.Add(predsOfKnown, predToCheck)
		}

		return set.Diff(connData.vars.mconn, predsOfKnown)
	}
*/
func onReceivingHeads(ctx *context.AppContext, connData *connectionData, body []byte) {
	data, err := unmarshallJson[msgsDTO](body)
	if err != nil {
		logger.Error("Failed to parse msgs JSON", err)
		return
	}

	key := data.Key
	hs := set.FromSlice(data.Messages)

	mConnHashes := utils.Map(connData.vars.mconn, crdts.HashOperationFromString)
	missing := set.Diff(hs, mConnHashes)

	handleMissing(ctx, connData, key, missing)
}

func onReceivingMsgs(ctx *context.AppContext, connData *connectionData, body []byte) {
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

func onReceivingNeeds(ctx *context.AppContext, connData *connectionData, body []byte) {
	data, err := unmarshallJson[msgsDTO](body)
	if err != nil {
		logger.Error("Failed to parse msgs JSON", err)
		return
	}
	key := data.Key
	needs := set.FromSlice(data.Messages)

	reply := utils.Filter(set.Diff(connData.vars.mconn, connData.vars.sent), func(el string) bool {
		if set.Has(needs, crdts.HashOperationFromString(el)) {
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

	if len(connData.vars.missing) == 0 {
		msgs := set.Diff(set.FromSlice(connData.vars.recvd), M)

		M = set.Union(M, connData.vars.recvd)
		connData.vars.mconn = set.Union(connData.vars.mconn, connData.vars.recvd)

		var orderedMsgs []string = crdts.CalculateOperationsTopologicalOrder(msgs)

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
		NewMessage(NEEDS).AddContent(
			msgsDTO{
				Key:      key,
				Messages: connData.vars.missing,
			}).Send(connData.conn)
	}
}
