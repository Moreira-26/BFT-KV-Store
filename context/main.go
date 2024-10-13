package context

import (
	"bftkvstore/storage"
	"crypto/ed25519"
	"net"
	"sync"
)

type Node struct {
	Address string
	Port    string
	Conn    net.Conn
}

type AppContext struct {
	Lock      sync.Mutex
	Secretkey ed25519.PrivateKey
	Address   string
	Port      string
	NewNodes  []Node
	Storage   storage.Storage
}

func New(secretkey ed25519.PrivateKey, hostname string, port string) AppContext {
	return AppContext{
		Secretkey: secretkey,
		Address:   hostname,
		Port:      port,
		NewNodes:  make([]Node, 0),
		Storage:   storage.Init(),
	}
}

func (ctx *AppContext) AddNewNode(address string, port string, conn net.Conn) {
	ctx.Lock.Lock()
	defer ctx.Lock.Unlock()

	for _, node := range ctx.NewNodes {
		if node.Address == address && node.Port == port {
			return
		}
	}

	ctx.NewNodes = append(ctx.NewNodes, Node{Address: address, Port: port, Conn: conn})
}
