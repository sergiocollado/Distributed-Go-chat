package main

import (
	"fmt"
	"os"

	"github.com/sergiocollado/Distributed-Go-chat/internal/chatroom"
)

func main() {
	fmt.Println("Starting server from cmd/server...")
	chatroom.StartServer()
	os.Exit(0)
}
