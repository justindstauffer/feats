package websocket

import "testing"

func TestSubscribeToGroup_DeniedWhenAuthorizerRejects(t *testing.T) {
	hub := NewHub()
	hub.SetSubscriptionAuthorizer(func(userID, groupID string) bool {
		return false
	})

	client := &Client{
		userID: "u1",
		groups: map[string]bool{},
		send:   make(chan []byte, 1),
	}

	ok := hub.subscribeToGroup(client, "g1")
	if ok {
		t.Fatal("expected subscribe to be denied")
	}

	if client.groups["g1"] {
		t.Fatal("client should not be subscribed when denied")
	}

	if _, exists := hub.groupSubscriptions["g1"]; exists {
		t.Fatal("hub should not track denied subscription")
	}
}

func TestSubscribeToGroup_AllowsWhenAuthorizerApproves(t *testing.T) {
	hub := NewHub()
	hub.SetSubscriptionAuthorizer(func(userID, groupID string) bool {
		return userID == "u1" && groupID == "g1"
	})

	client := &Client{
		userID: "u1",
		groups: map[string]bool{},
		send:   make(chan []byte, 1),
	}

	ok := hub.subscribeToGroup(client, "g1")
	if !ok {
		t.Fatal("expected subscribe to be allowed")
	}

	if !client.groups["g1"] {
		t.Fatal("client should be subscribed when allowed")
	}

	subscribers, exists := hub.groupSubscriptions["g1"]
	if !exists {
		t.Fatal("expected group subscription map to exist")
	}
	if !subscribers[client] {
		t.Fatal("expected client to be in group subscriptions")
	}
}
