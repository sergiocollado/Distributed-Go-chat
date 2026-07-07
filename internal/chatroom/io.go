package chatroom

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"
)

func handleClient(conn net.Conn, chatRoom *ChatRoom) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in handleClient: %v\n", r)
		}
		conn.Close()
	}()

	// Set initial timeout for username entry
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	// The initial 30-second timeout prevents connection exhaustion
	// by disconnecting clients who don't enter a username quickly.

	reader := bufio.NewReader(conn)

	// Prompt for username or reconnection
	conn.Write([]byte("Enter username (or 'reconnect:<username>:<token>'): \n"))

	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Failed to read username:", err)
		return
	}
	input = strings.TrimSpace(input)

	var username string
	var reconnectToken string
	var isReconnecting bool

	// Parse reconnection attempt
	if strings.HasPrefix(input, "reconnect:") {
		parts := strings.Split(input, ":")
		if len(parts) == 3 {
			username = parts[1]
			reconnectToken = parts[2]
			isReconnecting = true
		} else {
			conn.Write([]byte("Invalid format. Use: reconnect:<username>:<token>\n"))
			return
		}
	} else {
		username = input
	}

	// Generate guest name if empty
	if username == "" {
		username = fmt.Sprintf("Guest%d", rand.Intn(1000))
	}

	// Validate reconnection or check for duplicate
	if isReconnecting {
		if chatRoom.validateReconnectToken(username, reconnectToken) {
			fmt.Printf("%s reconnected successfully\n", username)
			conn.Write([]byte(fmt.Sprintf("Welcome back, %s!\n", username)))
		} else {
			conn.Write([]byte("Invalid token or session expired.\n"))
			return
		}
	} else {
		// Prevent duplicate logins
		if chatRoom.isUsernameConnected(username) {
			conn.Write([]byte("Username already connected. Use reconnect if you lost connection.\n"))
			return
		}

		// Create or retrieve session
		chatRoom.sessionsMu.Lock()
		existingSession := chatRoom.sessions[username]
		chatRoom.sessionsMu.Unlock()

		if existingSession != nil {
			token := existingSession.ReconnectToken
			msg := fmt.Sprintf("Tip: Save this token: %s\n", token)
			msg += fmt.Sprintf("To reconnect: reconnect:%s:%s\n", username, token)
			conn.Write([]byte(msg))
		} else {
			session := chatRoom.createSession(username)
			token := session.ReconnectToken
			msg := fmt.Sprintf("Your token: %s\n", token)
			msg += fmt.Sprintf("To reconnect: reconnect:%s:%s\n", username, token)
			conn.Write([]byte(msg))
		}
	}

	// Create client object
	client := &Client{
		conn:           conn,
		username:       username,
		outgoing:       make(chan string, 10), //The buffered outgoing channel prevents slow clients from blocking the broadcaster.
		lastActive:     time.Now(),
		reconnectToken: reconnectToken,       // Token-based reconnection lets users resume their session without complex authentication
		isSlowClient:   rand.Float64() < 0.1, // 10% chance for testing
	}

	// Clear timeout for normal operation
	conn.SetReadDeadline(time.Time{})

	// Notify chatroom
	chatRoom.join <- client

	// Send welcome message
	welcomeMsg := buildWelcomeMessage(username)
	conn.Write([]byte(welcomeMsg))

	// Start read/write loops
	go readMessages(client, chatRoom)
	writeMessages(client) // Blocks until disconnect

	// Update session on disconnect
	chatRoom.updateSessionActivity(username)
	chatRoom.leave <- client
}

func buildWelcomeMessage(username string) string {
	msg := fmt.Sprintf("Welcome, %s!\n", username)
	msg += "Commands:\n"
	msg += "  /users - List all users\n"
	msg += "  /history [N] - Show last N messages\n"
	msg += "  /msg <user> <msg> - Private message\n"
	msg += "  /token - Show your reconnect token\n"
	msg += "  /stats - Show your stats\n"
	msg += "  /quit - Leave\n"
	return msg
}
