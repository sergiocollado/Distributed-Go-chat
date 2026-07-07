package chatroom

import (
	"fmt"
	"time"
)

func NewChatRoom(dataDir string) (*ChatRoom, error) {
	cr := &ChatRoom{
		clients:       make(map[*Client]bool),
		join:          make(chan *Client),
		leave:         make(chan *Client),
		broadcast:     make(chan string),
		listUsers:     make(chan *Client),
		directMessage: make(chan DirectMessage),
		sessions:      make(map[string]*SessionInfo),
		messages:      make([]Message, 0),
		startTime:     time.Now(),
		dataDir:       dataDir,
	}

	// Restore from snapshot if available
	if err := cr.loadSnapshot(); err != nil {
		fmt.Printf("Failed to load snapshot: %v\n", err)
	}

	// Initialize WAL(Write-Ahead Log) for new messages
	if err := cr.initializePersistence(); err != nil {
		return nil, err
	}

	// Start background snapshot worker
	go cr.periodicSnapshots()

	return cr, nil
}

func (cr *ChatRoom) periodicSnapshots() {
	// this function runs in a separate goroutine
	// it wakes up every 5 minutes and checks if
	// it is needed to create a snapshot
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {

		// the routine gets the Lock to count the messages and releases it.
		// The lock is not adquired during the createSnapshot() because it
		// would take much more time and that would block the message broadcasting.
		// Why 5 minutes and 100 messages? this are tunnable paremetes, get the best ones for you.
		cr.messageMu.Lock()
		messageCount := len(cr.messages)
		cr.messageMu.Unlock()

		if messageCount > 100 {
			if err := cr.createSnapshot(); err != nil {
				fmt.Printf("Snapshot failed: %v\n", err)
			}
		}
	}
}

func (cr *ChatRoom) Run() {
	fmt.Println("ChatRoom heartbeat started...")
	go cr.cleanupInactiveClients()

	for {
		// The loop blocks on the select statement, waiting for data
		// on any of the five channels. When data arrives on any channel,
		// that case executes.
		// After the case completes, the loop goes back to waiting.
		select {
		case client := <-cr.join:
			cr.handleJoin(client)

		case client := <-cr.leave:
			cr.handleLeave(client)

		case message := <-cr.broadcast:
			cr.handleBroadcast(message)

		case client := <-cr.listUsers:
			cr.sendUserList(client)

		case dm := <-cr.directMessage:
			cr.handleDirectMessage(dm)
		}
	}
}
