package websocket

import "testing"

func TestHandleMessageSubscribe_DeniedByAuthorizer(t *testing.T) {
	hub := NewHub()
	hub.SetSubscriptionAuthorizer(func(userID, groupID string) bool {
		return false
	})

	client := &Client{
		hub:    hub,
		userID: "user-1",
		groups: map[string]bool{},
		send:   make(chan []byte, 1),
	}

	client.handleMessage([]byte(`{"action":"subscribe","group_id":"group-1"}`))

	if client.groups["group-1"] {
		t.Fatalf("expected denied subscription to not update client groups")
	}
	if _, ok := hub.groupSubscriptions["group-1"]; ok {
		t.Fatalf("expected denied subscription to not update hub subscriptions")
	}
}

func TestHandleMessageSubscribe_AllowedByAuthorizer(t *testing.T) {
	hub := NewHub()
	hub.SetSubscriptionAuthorizer(func(userID, groupID string) bool {
		return userID == "user-1" && groupID == "group-1"
	})

	client := &Client{
		hub:    hub,
		userID: "user-1",
		groups: map[string]bool{},
		send:   make(chan []byte, 1),
	}

	client.handleMessage([]byte(`{"action":"subscribe","group_id":"group-1"}`))

	if !client.groups["group-1"] {
		t.Fatalf("expected allowed subscription to update client groups")
	}
	subs, ok := hub.groupSubscriptions["group-1"]
	if !ok || !subs[client] {
		t.Fatalf("expected allowed subscription to update hub subscriptions")
	}
}
