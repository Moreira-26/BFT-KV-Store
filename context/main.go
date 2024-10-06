package context

import (
	"bftkvstore/storage"
	"crypto/rsa"
	"sync"
)

type AppContext struct {
	lock      sync.Mutex
	Secretkey *rsa.PrivateKey
	Publickey *rsa.PublicKey
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
