package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type UpdatePNCounter struct {
	ID           string   // Hash of the update
	NodeID       string   // ID of the node that generated the update
	Predecessors []string // Set of predecessor hashes
	OpType       string   //Type of the operation inc or dec
	Value        int      // Value
}

type PNCounter struct {
	ID          string
	CreatorNode string
	Positives   map[string]int             // Positive increments by node
	Negatives   map[string]int             // Negative decrements by node
	Updates     map[string]UpdatePNCounter // Update history, indexed by hash
	Heads       map[string]bool            // Current heads of the hash graph
	mu          sync.Mutex                 // Mutex for thread safety
}

// NewPNCounter creates a new PNCounter instance
func NewPNCounter(creatorNode string) *PNCounter {
	return &PNCounter{
		ID:          uuid.New().String(),
		CreatorNode: creatorNode,
		Positives:   make(map[string]int),
		Negatives:   make(map[string]int),
		Updates:     make(map[string]UpdatePNCounter),
		Heads:       make(map[string]bool),
	}
}

func NewPNCounterID(creatorNode string, pnCounterID string) *PNCounter {
	return &PNCounter{
		ID:          pnCounterID,
		CreatorNode: creatorNode,
		Positives:   make(map[string]int),
		Negatives:   make(map[string]int),
		Updates:     make(map[string]UpdatePNCounter),
		Heads:       make(map[string]bool),
	}
}

// Generate an update with its hash ID
func (p *PNCounter) generateUpdate(nodeId string, OpType string, Value int) UpdatePNCounter {
	p.mu.Lock()
	defer p.mu.Unlock()

	predecessors := p.getCurrentHeads()
	update := UpdatePNCounter{
		NodeID:       nodeId,
		OpType:       OpType,
		Value:        Value,
		Predecessors: predecessors,
	}

	// Create a hash of the update for its ID
	hash := p.hashUpdate(update)
	update.ID = hash

	// Store update in history and adjust heads
	p.Updates[hash] = update
	p.updateHeads(hash, predecessors)

	//Apply update locally
	if update.OpType == "inc" {
		p.Positives[nodeId] += update.Value
	} else if update.OpType == "dec" {
		p.Negatives[nodeId] += update.Value
	}

	return update
}

// Helper function to hash an update using SHA-256
func (p *PNCounter) hashUpdate(update UpdatePNCounter) string {
	data := fmt.Sprintf("%s:%s:%d:%v", update.NodeID, update.OpType, update.Value, update.Predecessors)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// Update the set of heads after adding a new update
func (p *PNCounter) updateHeads(newHash string, predecessors []string) {
	p.Heads[newHash] = true
	for _, pred := range predecessors {
		delete(p.Heads, pred)
	}
}

// Get the current heads of the hash graph
func (p *PNCounter) getCurrentHeads() []string {
	heads := make([]string, 0, len(p.Heads))
	for head := range p.Heads {
		heads = append(heads, head)
	}
	return heads
}

// Receive update from another node and applies effect to the current state
func (p *PNCounter) effect(update UpdatePNCounter) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if update already exists
	if _, exists := p.Updates[update.ID]; exists {
		return
	}

	// Validate the update (verify predecessors exist)
	for _, pred := range update.Predecessors {
		if _, exists := p.Updates[pred]; !exists {
			return // missing predecessor
		}
	}

	// Apply the update
	p.Updates[update.ID] = update
	p.updateHeads(update.ID, update.Predecessors)
	if update.OpType == "inc" {
		p.Positives[update.NodeID] += update.Value
	} else if update.OpType == "dec" {
		p.Negatives[update.NodeID] += update.Value
	}

}

// GetValue calculates the current Value of the PNCounter
func (p *PNCounter) GetValue() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	totalPos, totalNeg := 0, 0
	for _, inc := range p.Positives {
		totalPos += inc
	}
	for _, dec := range p.Negatives {
		totalNeg += dec
	}
	return totalPos - totalNeg
}

// func main() {
// 	//Node1 Creates PNCounter
// 	node1PN := NewPNCounter("node1")

// 	//Node2 Receives Create
// 	node2PN := NewPNCounterID("node1", node1PN.ID)

// 	//Node1 Create update
// 	update1 := node1PN.generateUpdate("node1", "inc", 5) // node1 increments by 5
// 	//Node2 Create update
// 	update2 := node2PN.generateUpdate("node2", "dec", 3) // node2 decrements by 3

// 	//Simulates node1 --update1-> node2
// 	node2PN.effect(update1)

// 	//Simulates node2 --update2-> node1
// 	node1PN.effect(update2)
// 	node1PN.effect(update2)

// 	fmt.Printf("Node1 Value: %d\n", node1PN.GetValue())
// 	fmt.Printf("Node2 Value: %d\n", node2PN.GetValue())
// }
