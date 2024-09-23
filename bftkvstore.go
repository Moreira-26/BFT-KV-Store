package main

import (
	"bftkvstore/api"
	"bftkvstore/protocol"
	"fmt"
)

func main() {
	fmt.Println("Hello, World!")

	go protocol.Start()

	api.Start()
}
