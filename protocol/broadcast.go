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

var M set.Set[string] = set.New[string]()

type connectionVariables struct {
	sent    []string
	recvd   set.Set[string]
	missing []string
	mconn   set.Set[string]
}

var connVars connectionVariables = connectionVariables{
	sent:    make([]string, 0),
	recvd:   set.New[string](),
	missing: make([]string, 0),
	mconn:   M,
}

type msgsDTO struct {
	Key      string   `json:"key"`
	Messages []string `json:"messages"`
}

// Algorithm 1 A Byzantine causal broadcast algorithm.
func broadcast(ctx *context.AppContext, key string, m crdts.SignedOperation) {
	// should have already have m
	ctx.Storage.Append(key, m)

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

	for _, node := range ctx.Nodes {
		if node.Conn != nil {
			if err := msg.Send(node.Conn); err != nil {
				logger.Alert("Failed to send Message during broadcast", err)
			}
		}
	}
}

func on_connection_to_another_replica() {
	// connection-local variables

	/*
		connVars := connectionVariables{
			sent:    make([]string, 0),
			recvd:   set.New[string](),
			missing: make([]string, 0),
			mconn:   M,
		}
	*/

	// send ⟨heads : heads(Mconn)⟩ via current connection
}

func on_receiving_heads() {
	/*
	 on receiving ⟨heads : hs⟩ via a connection do
	 HandleMissing({ℎ ∈ hs | 𝑚 ∈ Mconn. 𝐻 (𝑚) = ℎ})
	*/
}

func on_receiving_msgs(ctx *context.AppContext, conn net.Conn, body []byte) {
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
		connVars.recvd = set.Add(connVars.recvd, msg)
	}

	// unresolved := {ℎ | ∃(𝑣, hs, sig) ∈ recvd . ℎ ∈ hs ∧ 𝑚 ∈ (Mconn ∪ recvd ). 𝐻 (𝑚) = ℎ}
	predsToCheck := set.New[string]()
	for _, signedOp := range signedOps {
		for _, pred := range signedOp.Preds {
			predsToCheck = set.Add(predsToCheck, pred)
		}
	}

	toHash := set.Union(connVars.mconn, connVars.recvd)
	hashes := utils.Map(toHash, crdts.HashOperationFromString)

	unresolved := set.Diff(predsToCheck, hashes)

	handleMissing(unresolved)

	/*
	 on receiving ⟨msgs : new ⟩ via a connection do
	 recvd := recvd ∪ {(𝑣, hs, sig) ∈ new | check((𝑣, hs), sig)}
	 unresolved := {ℎ | ∃(𝑣, hs, sig) ∈ recvd . ℎ ∈ hs ∧
	 𝑚 ∈ (Mconn ∪ recvd ). 𝐻 (𝑚) = ℎ}
	 HandleMissing(unresolved )
	*/
}

func on_receiving_needs() {
	/*
	 on receiving ⟨needs : hashes⟩ via a connection do
	 reply := {𝑚 ∈ Mconn | 𝐻 (𝑚) ∈ hashes ∧ 𝑚 ∉ sent }
	 sent := sent ∪ reply
	 send ⟨msgs : reply⟩ via current connection
	*/
}

func handleMissing(hashes set.Set[string]) {
	fmt.Println("missing", hashes)

	connVars.missing = set.Diff(
		set.Union(set.FromSlice(connVars.missing), set.FromSlice(hashes)),
		set.FromSlice(utils.Map[string, string](connVars.recvd, crdts.HashOperationFromString)),
	)

	if len(connVars.missing) == 0 {
		// TODO: Add lock here because it is atomic
		// msgs := set.Diff(set.FromSlice(connVars.recvd), M)
		M = set.Union(M, connVars.recvd)
		// TODO: deliver all of the messages in msgs in topologically sorted order

	} else {
		// TODO: The Key situation
		msg := NewMessage(NEEDS).AddContent(
			msgsDTO{
				Key:      "This is wrong",
				Messages: connVars.missing,
			})

		logger.Debug("Should eventually send to conn", msg)
		// TODO: Send
	}
}
