package context

import (
	"bftkvstore/storage"
	"crypto/ed25519"
	"sync"
)

type AppContext struct {
	lock      sync.Mutex
	Secretkey ed25519.PrivateKey
	Address   string
	Port      string
	Nodes     []struct {
		Address string
		Port    string
	}
	Storage storage.Storage
}

func (ctx *AppContext) AddNewNode(address string, port string) {
	ctx.lock.Lock()
	defer ctx.lock.Unlock()

	for _, node := range ctx.Nodes {
		if node.Address == address && node.Port == port {
			return
		}
	}

	ctx.Nodes = append(ctx.Nodes, struct {
		Address string
		Port    string
	}{address, port})
}
