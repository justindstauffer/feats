package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients mapped by userID
	clients map[string]map[*Client]bool

	// Group subscriptions: groupID -> set of clients
	groupSubscriptions map[string]map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast events to group members
	broadcast chan *Event

	// Mutex for thread-safe access
	mu sync.RWMutex

	// Optional authorization callback for dynamic group subscriptions.
	// Returns true when userID is allowed to subscribe to groupID.
	subscriptionAuthorizer func(userID, groupID string) bool
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:            make(map[string]map[*Client]bool),
		groupSubscriptions: make(map[string]map[*Client]bool),
		register:           make(chan *Client),
		unregister:         make(chan *Client),
		broadcast:          make(chan *Event, 256),
	}
}

// SetSubscriptionAuthorizer configures server-side authorization for subscribe actions.
func (h *Hub) SetSubscriptionAuthorizer(authorizer func(userID, groupID string) bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.subscriptionAuthorizer = authorizer
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case event := <-h.broadcast:
			h.broadcastEvent(event)
		}
	}
}

// registerClient adds a new client to the hub
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Add to user's client set
	if _, ok := h.clients[client.userID]; !ok {
		h.clients[client.userID] = make(map[*Client]bool)
	}
	h.clients[client.userID][client] = true

	// Subscribe to all groups the user is a member of
	for groupID := range client.groups {
		if _, ok := h.groupSubscriptions[groupID]; !ok {
			h.groupSubscriptions[groupID] = make(map[*Client]bool)
		}
		h.groupSubscriptions[groupID][client] = true
	}

	log.Printf("WebSocket client connected: user=%s, groups=%d", client.userID, len(client.groups))
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Remove from user's client set
	if clients, ok := h.clients[client.userID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.clients, client.userID)
		}
	}

	// Remove from all group subscriptions
	for groupID := range client.groups {
		if clients, ok := h.groupSubscriptions[groupID]; ok {
			delete(clients, client)
			if len(clients) == 0 {
				delete(h.groupSubscriptions, groupID)
			}
		}
	}

	close(client.send)
	log.Printf("WebSocket client disconnected: user=%s", client.userID)
}

// broadcastEvent sends an event to all clients subscribed to the event's group
func (h *Hub) broadcastEvent(event *Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.groupSubscriptions[event.GroupID]
	if !ok {
		return
	}

	message, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return
	}

	for client := range clients {
		select {
		case client.send <- message:
		default:
			// Client's send buffer is full, close the connection
			h.mu.RUnlock()
			h.unregisterClient(client)
			h.mu.RLock()
		}
	}
}

// subscribeToGroup adds a client to a group subscription
func (h *Hub) subscribeToGroup(client *Client, groupID string) bool {
	h.mu.RLock()
	authorizer := h.subscriptionAuthorizer
	h.mu.RUnlock()

	if authorizer != nil && !authorizer(client.userID, groupID) {
		log.Printf("WebSocket subscription denied: user=%s, group=%s", client.userID, groupID)
		return false
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	client.groups[groupID] = true

	if _, ok := h.groupSubscriptions[groupID]; !ok {
		h.groupSubscriptions[groupID] = make(map[*Client]bool)
	}
	h.groupSubscriptions[groupID][client] = true

	log.Printf("Client subscribed to group: user=%s, group=%s", client.userID, groupID)
	return true
}

// unsubscribeFromGroup removes a client from a group subscription
func (h *Hub) unsubscribeFromGroup(client *Client, groupID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(client.groups, groupID)

	if clients, ok := h.groupSubscriptions[groupID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.groupSubscriptions, groupID)
		}
	}

	log.Printf("Client unsubscribed from group: user=%s, group=%s", client.userID, groupID)
}

// BroadcastToGroup sends an event to all clients in a group
// This is the main method called by handlers to broadcast events
func (h *Hub) BroadcastToGroup(event *Event) {
	h.broadcast <- event
}

// BroadcastToUser sends a message to all connections of a specific user
func (h *Hub) BroadcastToUser(userID string, event *Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.clients[userID]
	if !ok {
		return
	}

	message, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return
	}

	for client := range clients {
		select {
		case client.send <- message:
		default:
			// Client's send buffer is full
		}
	}
}

// GetConnectedUsers returns the number of connected users
func (h *Hub) GetConnectedUsers() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetConnectedClients returns the total number of connected clients
func (h *Hub) GetConnectedClients() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	count := 0
	for _, clients := range h.clients {
		count += len(clients)
	}
	return count
}

// GetGroupSubscribers returns the number of clients subscribed to a group
func (h *Hub) GetGroupSubscribers(groupID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.groupSubscriptions[groupID])
}

// Register adds a client to the hub
func (h *Hub) Register(client *Client) {
	h.register <- client
}
