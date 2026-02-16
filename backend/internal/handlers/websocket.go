package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jstauff/feats-api/internal/services"
	ws "github.com/jstauff/feats-api/internal/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, you should check the origin
		// For now, allow all origins
		return true
	},
}

type WebSocketHandler struct {
	hub          *ws.Hub
	authService  *services.AuthService
	groupService *services.GroupService
}

func NewWebSocketHandler(hub *ws.Hub, authService *services.AuthService, groupService *services.GroupService) *WebSocketHandler {
	if hub != nil {
		hub.SetSubscriptionAuthorizer(func(userID, groupID string) bool {
			return groupService.IsGroupMember(groupID, userID)
		})
	}

	return &WebSocketHandler{
		hub:          hub,
		authService:  authService,
		groupService: groupService,
	}
}

// HandleWebSocket upgrades HTTP connection to WebSocket
// Authentication is done via query parameter: ?token=<jwt_token>
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Get token from query parameter
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
		return
	}

	// Validate token
	claims, err := h.authService.ValidateAccessToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// Get user
	user, err := h.authService.GetUserByID(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	// Check if user is locked
	if user.IsLocked() {
		c.JSON(http.StatusForbidden, gin.H{"error": "account locked"})
		return
	}

	// Get user's groups
	groups, err := h.groupService.ListUserGroups(user.ID)
	if err != nil {
		log.Printf("Failed to get user groups for WebSocket: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get groups"})
		return
	}

	groupIDs := make([]string, len(groups))
	for i, g := range groups {
		groupIDs[i] = g.ID
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}

	// Create client and register with hub
	client := ws.NewClient(h.hub, conn, user.ID, groupIDs)

	// Register client
	h.hub.Register(client)

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()
}

// GetStats returns WebSocket connection statistics
func (h *WebSocketHandler) GetStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"connected_users":   h.hub.GetConnectedUsers(),
		"connected_clients": h.hub.GetConnectedClients(),
	})
}
