package main

import (
	"fmt"

	"github.com/sergiocollado/Distributed-Go-chat//chatroom/internal/chatroom"
)

func main() {
	fmt.Println("Starting client from cmd/client...")
	chatroom.StartClient()
}
