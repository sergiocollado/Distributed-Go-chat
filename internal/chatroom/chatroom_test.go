package chatroom

import (
	"strings"
	"testing"
	"time"
)

/*
Unit test:
run the unit tests with: go test ./internal/chatroom -v
// -v is the verbose flag

The test creates two clients structures, and dont use real TCP connections.
The test sends both clients to the join channel, waits 100 milliseconds for
 them to be processed, then broadcasts a message.
*/

func TestBroadcast(t *testing.T) {
	cr, _ := NewChatRoom("./testdata") // creates a chatroom
	defer cr.shutdown()

	go cr.Run() // runs the chatroom

	// Create mock clients
	client1 := &Client{
		username: "Alice",
		outgoing: make(chan string, 10),
	}

	client2 := &Client{
		username: "Bob",
		outgoing: make(chan string, 10),
	}

	// Join clients
	cr.join <- client1
	cr.join <- client2
	time.Sleep(100 * time.Millisecond)

	// Broadcast message
	cr.broadcast <- "[Alice]: Hello!"

	// Verify both receive it
	select {
	case msg := <-client1.outgoing:
		if !strings.Contains(msg, "Hello!") {
			t.Fatal("Client1 didn't receive correct message")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Client1 didn't receive message")
	}

	select {
	case msg := <-client2.outgoing:
		if !strings.Contains(msg, "Hello!") {
			t.Fatal("Client2 didn't receive correct message")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Client2 didn't receive message")
	}
}
