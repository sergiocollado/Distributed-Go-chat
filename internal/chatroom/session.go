package chatroom

/*
Session management lets users reconnect to your server after network interruptions
without needing to create a new account or re-enter credentials. You'll implement
this using cryptographically secure tokens that persist across connections.
*/

import (
	"fmt"
	"time"

	"github.com/sergiocollado/Distributed-Go-chat/pkg/token"
)

func (cr *ChatRoom) createSession(username string) *SessionInfo {
	cr.sessionsMu.Lock()
	defer cr.sessionsMu.Unlock()

	tok := token.GenerateToken()

	session := &SessionInfo{
		Username:       username,
		ReconnectToken: tok,
		LastSeen:       time.Now(),
		CreatedAt:      time.Now(),
	}

	cr.sessions[username] = session

	fmt.Printf("Created session for %s (token: %s...)\n", username, tok[:8])

	return session
}

func (cr *ChatRoom) validateReconnectToken(username, token string) bool {
	cr.sessionsMu.Lock()
	defer cr.sessionsMu.Unlock()

	session, exists := cr.sessions[username]
	if !exists {
		return false
	}

	if session.ReconnectToken != token {
		return false
	}

	if time.Since(session.LastSeen) > 1*time.Hour {
		delete(cr.sessions, username)
		return false
	}

	session.LastSeen = time.Now()

	return true
}

func (cr *ChatRoom) updateSessionActivity(username string) {
	cr.sessionsMu.Lock()
	defer cr.sessionsMu.Unlock()

	if session, exists := cr.sessions[username]; exists {
		session.LastSeen = time.Now()
	}
}

func (cr *ChatRoom) isUsernameConnected(username string) bool {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	for client := range cr.clients {
		if client.username == username {
			return true
		}
	}

	return false
}

func (cr *ChatRoom) cleanupInactiveClients() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		cr.mu.Lock()
		var toRemove []*Client

		for client := range cr.clients {
			if client.isInactive(5 * time.Minute) {
				fmt.Printf("Removing inactive: %s\n", client.username)
				toRemove = append(toRemove, client)
			}
		}
		cr.mu.Unlock()

		for _, client := range toRemove {
			cr.leave <- client
		}
	}
}
