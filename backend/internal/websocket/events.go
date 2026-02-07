package websocket

import (
	"encoding/json"
	"time"
)

// EventType represents the type of WebSocket event
type EventType string

const (
	// Post events
	EventPostCreated EventType = "post.created"
	EventPostDeleted EventType = "post.deleted"

	// Reaction events
	EventReactionAdded   EventType = "reaction.added"
	EventReactionRemoved EventType = "reaction.removed"

	// Comment events
	EventCommentCreated EventType = "comment.created"
	EventCommentDeleted EventType = "comment.deleted"

	// Challenge events
	EventChallengeCreated  EventType = "challenge.created"
	EventChallengeJoined   EventType = "challenge.joined"
	EventChallengeLeft     EventType = "challenge.left"
	EventChallengeProgress EventType = "challenge.progress"

	// Group member events
	EventMemberJoined EventType = "member.joined"
	EventMemberLeft   EventType = "member.left"

	// Streak events
	EventStreakUpdated EventType = "streak.updated"
)

// Event represents a WebSocket event to be broadcast
type Event struct {
	Type      EventType       `json:"type"`
	GroupID   string          `json:"group_id"`
	UserID    string          `json:"user_id,omitempty"` // The user who triggered the event
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
}

// NewEvent creates a new event with the given type, group, user, and payload
func NewEvent(eventType EventType, groupID, userID string, payload interface{}) (*Event, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &Event{
		Type:      eventType,
		GroupID:   groupID,
		UserID:    userID,
		Payload:   payloadJSON,
		Timestamp: time.Now(),
	}, nil
}

// PostCreatedPayload contains data for post.created events
type PostCreatedPayload struct {
	PostID         string `json:"post_id"`
	UserID         string `json:"user_id"`
	UserName       string `json:"user_name"`
	ActivityTypeID string `json:"activity_type_id"`
	ActivityName   string `json:"activity_name"`
	ActivityIcon   string `json:"activity_icon,omitempty"`
	Description    string `json:"description,omitempty"`
	ImageCount     int    `json:"image_count"`
}

// PostDeletedPayload contains data for post.deleted events
type PostDeletedPayload struct {
	PostID string `json:"post_id"`
}

// ReactionPayload contains data for reaction events
type ReactionPayload struct {
	PostID       string `json:"post_id"`
	UserID       string `json:"user_id"`
	UserName     string `json:"user_name"`
	ReactionType string `json:"reaction_type"`
}

// CommentCreatedPayload contains data for comment.created events
type CommentCreatedPayload struct {
	CommentID string  `json:"comment_id"`
	PostID    string  `json:"post_id"`
	UserID    string  `json:"user_id"`
	UserName  string  `json:"user_name"`
	Content   string  `json:"content"`
	ParentID  *string `json:"parent_id,omitempty"`
}

// CommentDeletedPayload contains data for comment.deleted events
type CommentDeletedPayload struct {
	CommentID string `json:"comment_id"`
	PostID    string `json:"post_id"`
}

// ChallengeCreatedPayload contains data for challenge.created events
type ChallengeCreatedPayload struct {
	ChallengeID  string `json:"challenge_id"`
	Name         string `json:"name"`
	CreatorID    string `json:"creator_id"`
	CreatorName  string `json:"creator_name"`
	ActivityName string `json:"activity_name"`
	TargetCount  int    `json:"target_count"`
}

// ChallengeJoinedPayload contains data for challenge.joined events
type ChallengeJoinedPayload struct {
	ChallengeID   string `json:"challenge_id"`
	ChallengeName string `json:"challenge_name"`
	UserID        string `json:"user_id"`
	UserName      string `json:"user_name"`
}

// ChallengeLeftPayload contains data for challenge.left events
type ChallengeLeftPayload struct {
	ChallengeID   string `json:"challenge_id"`
	ChallengeName string `json:"challenge_name"`
	UserID        string `json:"user_id"`
}

// ChallengeProgressPayload contains data for challenge.progress events
type ChallengeProgressPayload struct {
	ChallengeID   string `json:"challenge_id"`
	ChallengeName string `json:"challenge_name"`
	UserID        string `json:"user_id"`
	UserName      string `json:"user_name"`
	Progress      int    `json:"progress"`
	TargetCount   int    `json:"target_count"`
	Completed     bool   `json:"completed"`
}

// MemberJoinedPayload contains data for member.joined events
type MemberJoinedPayload struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
}

// MemberLeftPayload contains data for member.left events
type MemberLeftPayload struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
}

// StreakUpdatedPayload contains data for streak.updated events
type StreakUpdatedPayload struct {
	UserID        string `json:"user_id"`
	UserName      string `json:"user_name"`
	CurrentStreak int    `json:"current_streak"`
	LongestStreak int    `json:"longest_streak"`
}
