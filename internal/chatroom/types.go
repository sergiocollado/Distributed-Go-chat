package chatroom

import (
	"net"
	"os"
	"sync"
	"time"
)

// Message represents a single chat message with metadata
type Message struct {
	ID        int       `json:"id"` // uniquely identifies each message and ensures messages are in order
	From      string    `json:"from"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"` // when the messages are sent
	Channel   string    `json:"channel"`   // "global" or "private:username"
}

// When the message is written/readed (Marshal/unMarshal) from the WAL, it should be something like:
//{"id":1,"from":"Alice","content":"Hello world","timestamp":"2024-02-06T10:00:00Z","channel":"global"}
//{"id":2,"from":"Bob","content":"Hi Alice!","timestamp":"2024-02-06T10:00:05Z","channel":"global"}

// Client represents a connected user
type Client struct {
	conn         net.Conn    // TCP connection
	username     string      // Display name
	outgoing     chan string // this a channel of strings! - Buffered channel for writes, this means we can buffer X strings without blocking
	lastActive   time.Time   // this will allow us to identify idle users, and we could disconnect them to free resources
	messagesSent int         // Statistics
	messagesRecv int
	isSlowClient bool // Testing flag

	reconnectToken string
	mu             sync.Mutex // Protects statistics fields: messagesSent and messagesRecv
}

// ChatRoom is the central coordinator
type ChatRoom struct {
	// Communication channels

	// this channeles are unbuffered, capacity 0, because we want synchronization. When data is sent to
	// an unbuffered channel, you block it till someone receives. This ensures event loop processes events in order.

	join          chan *Client // when a new client connects, we sent the client to the join channel.
	leave         chan *Client
	broadcast     chan string // when a client sends a message, we sent the client to the broadcast channel
	listUsers     chan *Client
	directMessage chan DirectMessage

	// State
	clients       map[*Client]bool
	mu            sync.Mutex // protects the client's map
	totalMessages int
	startTime     time.Time

	// Message history
	messages      []Message
	messageMu     sync.Mutex // protects the session's map
	nextMessageID int

	// why two mutexes: mu and messageMu? efficiency, if we use one mutex for everything, broadcasting a message would
	// lock all the data, preventing new clients from joining. Separate mutexes mean diferent operations can happen concurrently

	// Persistence
	walFile *os.File   // write ahead log File
	walMu   sync.Mutex // because writing to the disk is slow. We don't want to keep the main mutex while waiting for IO operations to disk
	dataDir string

	// Sessions
	sessions   map[string]*SessionInfo
	sessionsMu sync.Mutex
}

// SessionInfo tracks reconnection data
type SessionInfo struct {
	Username       string
	ReconnectToken string
	LastSeen       time.Time
	CreatedAt      time.Time
}

// DirectMessage represents a private message
type DirectMessage struct {
	toClient *Client
	message  string
}
