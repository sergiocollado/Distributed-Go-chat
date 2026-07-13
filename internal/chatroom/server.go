package chatroom

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func runServer() {

	// creates the chatroom
	chatRoom, err := NewChatRoom("./chatdata")
	if err != nil {
		fmt.Printf("Failed to initialize: %v\n", err)
		return
	}

	// defers the shutdown (this will create the final WAL snapshot)
	defer chatRoom.shutdown()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived shutdown signal")
		chatRoom.shutdown()
		os.Exit(0)
	}()

	// launches the chatroom in a gorotine
	go chatRoom.Run()

	// listen for TCP connections
	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server started on :9000")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// for each connections creats a new clients, and
		// spawns it handling in a goroutine
		fmt.Println("New connection from:", conn.RemoteAddr())
		go handleClient(conn, chatRoom)
	}
}

func (cr *ChatRoom) shutdown() {
	fmt.Println("\nShutting down...")
	if err := cr.createSnapshot(); err != nil {
		fmt.Printf("Final snapshot failed: %v\n", err)
	}
	if cr.walFile != nil {
		cr.walFile.Close()
	}
	fmt.Println("Shutdown complete")
}

func StartServer() {
	runServer()
}
