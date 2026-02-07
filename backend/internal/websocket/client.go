package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// Client represents a single WebSocket connection
type Client struct {
	hub *Hub

	// The websocket connection
	conn *websocket.Conn

	// User information
	userID string

	// Groups the user is subscribed to
	groups map[string]bool

	// Buffered channel of outbound messages
	send chan []byte
}

// NewClient creates a new client
func NewClient(hub *Hub, conn *websocket.Conn, userID string, groups []string) *Client {
	groupMap := make(map[string]bool)
	for _, g := range groups {
		groupMap[g] = true
	}

	return &Client{
		hub:    hub,
		conn:   conn,
		userID: userID,
		groups: groupMap,
		send:   make(chan []byte, 256),
	}
}

// ReadPump pumps messages from the websocket connection to the hub
// The application runs readPump in a per-connection goroutine
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for user %s: %v", c.userID, err)
			}
			break
		}

		// Handle incoming messages (e.g., subscribe/unsubscribe commands)
		c.handleMessage(message)
	}
}

// WritePump pumps messages from the hub to the websocket connection
// A goroutine running writePump is started for each connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming client messages
func (c *Client) handleMessage(message []byte) {
	var msg struct {
		Action  string   `json:"action"`
		Groups  []string `json:"groups,omitempty"`
		GroupID string   `json:"group_id,omitempty"`
	}

	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Invalid message from user %s: %v", c.userID, err)
		return
	}

	switch msg.Action {
	case "subscribe":
		// Subscribe to a group
		if msg.GroupID != "" {
			c.hub.subscribeToGroup(c, msg.GroupID)
		}
	case "unsubscribe":
		// Unsubscribe from a group
		if msg.GroupID != "" {
			c.hub.unsubscribeFromGroup(c, msg.GroupID)
		}
	}
}

// IsSubscribedTo checks if the client is subscribed to a group
func (c *Client) IsSubscribedTo(groupID string) bool {
	return c.groups[groupID]
}
