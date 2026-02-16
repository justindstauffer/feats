package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jstauff/feats-api/internal/middleware"
	"github.com/jstauff/feats-api/internal/models"
	"github.com/jstauff/feats-api/internal/services"
	"github.com/jstauff/feats-api/internal/websocket"
)

// PostWorkflow coordinates non-CRUD side effects for post lifecycle events.
type PostWorkflow struct {
	streakService    *services.StreakService
	challengeService *services.ChallengeService
	goalService      *services.GoalService
	groupService     *services.GroupService
	pushService      *services.PushService
	wsHub            *websocket.Hub
}

func NewPostWorkflow(
	streakService *services.StreakService,
	challengeService *services.ChallengeService,
	goalService *services.GoalService,
	groupService *services.GroupService,
	pushService *services.PushService,
	wsHub *websocket.Hub,
) *PostWorkflow {
	return &PostWorkflow{
		streakService:    streakService,
		challengeService: challengeService,
		goalService:      goalService,
		groupService:     groupService,
		pushService:      pushService,
		wsHub:            wsHub,
	}
}

func (w *PostWorkflow) HandlePostCreated(c *gin.Context, post *models.Post, groupID, userID, activityTypeID string, postService *services.PostService) {
	// Update streak for this group.
	w.streakService.UpdateStreakForActivity(groupID, userID, time.Now())

	// Update challenge progress and create completion posts.
	completedChallengeIDs, _ := w.challengeService.UpdateProgressForActivity(groupID, userID, activityTypeID)
	for _, challengeID := range completedChallengeIDs {
		challenge, err := w.challengeService.GetChallengeByID(groupID, challengeID)
		if err == nil {
			postService.CreateChallengeCompletionPost(groupID, userID, challenge.Title)
		}
	}

	// Update goal progress for this group.
	w.goalService.UpdateProgressForActivity(groupID, userID, activityTypeID)

	// Broadcast post.created event via WebSocket.
	if w.wsHub != nil {
		user, _ := middleware.GetCurrentUser(c)
		payload := websocket.PostCreatedPayload{
			PostID:         post.ID,
			UserID:         post.UserID,
			UserName:       user.Name,
			ActivityTypeID: post.ActivityTypeID,
			ActivityName:   post.ActivityType.Name,
			ActivityIcon:   getActivityIcon(post.ActivityType.Icon),
			Description:    getDescription(post.Description),
			ImageCount:     len(post.Images),
		}
		if event, err := websocket.NewEvent(websocket.EventPostCreated, groupID, userID, payload); err == nil {
			w.wsHub.BroadcastToGroup(event)
		}
	}

	// Send push notification to other group members.
	if w.pushService != nil && w.groupService != nil {
		user, _ := middleware.GetCurrentUser(c)
		memberIDs, err := w.groupService.GetMemberUserIDs(groupID)
		if err == nil && len(memberIDs) > 1 {
			activityName := ""
			if post.ActivityType.Name != "" {
				activityName = post.ActivityType.Name
			}
			go w.pushService.NotifyNewPost(memberIDs, userID, user.Name, activityName, post.ID)
		}
	}
}

func (w *PostWorkflow) HandlePostDeleted(postID, groupID, userID string) {
	if w.wsHub == nil {
		return
	}

	payload := websocket.PostDeletedPayload{PostID: postID}
	if event, err := websocket.NewEvent(websocket.EventPostDeleted, groupID, userID, payload); err == nil {
		w.wsHub.BroadcastToGroup(event)
	}
}
