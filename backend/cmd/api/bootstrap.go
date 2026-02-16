package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jstauff/feats-api/internal/config"
	"github.com/jstauff/feats-api/internal/handlers"
	"github.com/jstauff/feats-api/internal/middleware"
	"github.com/jstauff/feats-api/internal/services"
	"github.com/jstauff/feats-api/internal/websocket"
	"gorm.io/gorm"
)

type appServices struct {
	auth       *services.AuthService
	user       *services.UserService
	betaInvite *services.BetaInviteService
	group      *services.GroupService
	post       *services.PostService
	activity   *services.ActivityService
	reaction   *services.ReactionService
	comment    *services.CommentService
	streak     *services.StreakService
	challenge  *services.ChallengeService
	goal       *services.GoalService
	audit      *services.AuditService
	push       *services.PushService
}

type appHandlers struct {
	auth       *handlers.AuthHandler
	user       *handlers.UserHandler
	group      *handlers.GroupHandler
	betaInvite *handlers.BetaInviteHandler
	post       *handlers.PostHandler
	activity   *handlers.ActivityHandler
	reaction   *handlers.ReactionHandler
	comment    *handlers.CommentHandler
	streak     *handlers.StreakHandler
	challenge  *handlers.ChallengeHandler
	goal       *handlers.GoalHandler
	push       *handlers.PushHandler
	ws         *handlers.WebSocketHandler
}

type appMiddleware struct {
	auth     *middleware.AuthMiddleware
	group    *middleware.GroupMiddleware
	rate     *middleware.RateLimiter
	security *middleware.SecurityMiddleware
}

func initStorage(cfg *config.Config) error {
	if err := os.MkdirAll(cfg.StoragePath+"/images", 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(cfg.StoragePath+"/profiles", 0755); err != nil {
		return err
	}
	return nil
}

func initServices(db *gorm.DB, cfg *config.Config) *appServices {
	return &appServices{
		auth:       services.NewAuthService(db, cfg),
		user:       services.NewUserService(db, cfg),
		betaInvite: services.NewBetaInviteService(db),
		group:      services.NewGroupService(db, cfg),
		post:       services.NewPostService(db, cfg),
		activity:   services.NewActivityService(db),
		reaction:   services.NewReactionService(db),
		comment:    services.NewCommentService(db),
		streak:     services.NewStreakService(db, cfg),
		challenge:  services.NewChallengeService(db),
		goal:       services.NewGoalService(db, cfg),
		audit:      services.NewAuditService(db),
		push:       services.NewPushService(db, cfg),
	}
}

func initHandlers(s *appServices, cfg *config.Config, wsHub *websocket.Hub) *appHandlers {
	postWorkflow := handlers.NewPostWorkflow(s.streak, s.challenge, s.goal, s.group, s.push, wsHub)

	return &appHandlers{
		auth:       handlers.NewAuthHandler(s.auth, s.user, s.betaInvite, s.audit, cfg),
		user:       handlers.NewUserHandler(s.user, s.audit, s.auth),
		group:      handlers.NewGroupHandler(s.group, s.audit, wsHub),
		betaInvite: handlers.NewBetaInviteHandler(s.betaInvite),
		post:       handlers.NewPostHandler(s.post, s.audit, cfg, postWorkflow),
		activity:   handlers.NewActivityHandler(s.activity),
		reaction:   handlers.NewReactionHandler(s.reaction, s.push, wsHub),
		comment:    handlers.NewCommentHandler(s.comment, s.push, wsHub),
		streak:     handlers.NewStreakHandler(s.streak),
		challenge:  handlers.NewChallengeHandler(s.challenge, s.push, wsHub),
		goal:       handlers.NewGoalHandler(s.goal),
		push:       handlers.NewPushHandler(s.push),
		ws:         handlers.NewWebSocketHandler(wsHub, s.auth, s.group),
	}
}

func initMiddleware(auth *services.AuthService, group *services.GroupService, cfg *config.Config) *appMiddleware {
	return &appMiddleware{
		auth:     middleware.NewAuthMiddleware(auth, cfg),
		group:    middleware.NewGroupMiddleware(group),
		rate:     middleware.NewRateLimiter(cfg),
		security: middleware.NewSecurityMiddleware(),
	}
}

func setupRouter(cfg *config.Config, h *appHandlers, m *appMiddleware) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(m.security.SecurityHeaders())
	router.Use(middleware.CORS(cfg))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.GET("/ws", h.ws.HandleWebSocket)

	v1 := router.Group("/api/v1")

	auth := v1.Group("/auth")
	auth.Use(m.rate.LoginRateLimit())
	{
		auth.POST("/register", h.auth.Register)
		auth.POST("/login", h.auth.Login)
		auth.POST("/refresh", h.auth.RefreshToken)
		auth.POST("/password/reset-request", h.auth.RequestPasswordReset)
		auth.POST("/password/reset", h.auth.ResetPassword)
	}

	protected := v1.Group("")
	protected.Use(m.auth.Authenticate())
	protected.Use(m.rate.APIRateLimit())
	{
		protected.POST("/auth/logout", h.auth.Logout)
		protected.POST("/auth/password/change", h.auth.ChangePassword)

		protected.GET("/users/me", h.user.GetCurrentUser)
		protected.PUT("/users/me", h.user.UpdateCurrentUser)
		protected.GET("/users/:id", h.user.GetUser)

		protected.POST("/devices", h.push.RegisterToken)
		protected.DELETE("/devices", h.push.UnregisterToken)
		protected.POST("/invites/redeem", m.rate.InviteRedeemRateLimit(), h.group.RedeemInvite)
	}

	groups := v1.Group("/groups")
	groups.Use(m.auth.Authenticate())
	groups.Use(m.rate.APIRateLimit())
	{
		groups.POST("", h.group.CreateGroup)
		groups.GET("", h.group.ListGroups)
	}

	group := v1.Group("/groups/:gid")
	group.Use(m.auth.Authenticate())
	group.Use(m.group.RequireGroupMember())
	group.Use(m.rate.APIRateLimit())
	{
		group.GET("", h.group.GetGroup)
		group.POST("/leave", h.group.LeaveGroup)
		group.GET("/members", h.group.ListMembers)

		group.GET("/users/:id/streak", h.streak.GetUserStreak)
		group.GET("/users/:id/goals", h.goal.GetUserGoals)

		group.GET("/activities", h.activity.ListActivities)
		group.POST("/activities", h.activity.CreateActivity)
		group.DELETE("/activities/:id", h.activity.DeleteActivity)

		group.GET("/posts", h.post.ListPosts)
		group.POST("/posts", m.rate.PostRateLimit(), h.post.CreatePost)
		group.GET("/posts/:id", h.post.GetPost)
		group.PUT("/posts/:id", h.post.UpdatePost)
		group.DELETE("/posts/:id", h.post.DeletePost)
		group.POST("/posts/:id/images", m.rate.UploadRateLimit(), h.post.UploadImage)
		group.DELETE("/posts/:id/images/:image_id", h.post.DeleteImage)

		group.GET("/posts/:id/reactions", h.reaction.GetReactions)
		group.POST("/posts/:id/reactions", h.reaction.AddReaction)
		group.DELETE("/posts/:id/reactions", h.reaction.RemoveReaction)

		group.GET("/posts/:id/comments", h.comment.GetComments)
		group.POST("/posts/:id/comments", h.comment.CreateComment)
		group.PUT("/comments/:id", h.comment.UpdateComment)
		group.DELETE("/comments/:id", h.comment.DeleteComment)

		group.GET("/streaks/leaderboard", h.streak.GetLeaderboard)

		group.GET("/challenges", h.challenge.ListChallenges)
		group.POST("/challenges", h.challenge.CreateChallenge)
		group.GET("/challenges/:id", h.challenge.GetChallenge)
		group.POST("/challenges/:id/join", h.challenge.JoinChallenge)
		group.DELETE("/challenges/:id/leave", h.challenge.LeaveChallenge)
		group.DELETE("/challenges/:id", h.challenge.DeleteChallenge)

		group.POST("/goals", h.goal.CreateGoal)
		group.PUT("/goals/:id", h.goal.UpdateGoal)
		group.DELETE("/goals/:id", h.goal.DeleteGoal)
	}

	groupAdmin := v1.Group("/groups/:gid")
	groupAdmin.Use(m.auth.Authenticate())
	groupAdmin.Use(m.group.RequireGroupAdmin())
	groupAdmin.Use(m.rate.APIRateLimit())
	{
		groupAdmin.PUT("", h.group.UpdateGroup)
		groupAdmin.DELETE("", h.group.DeleteGroup)
		groupAdmin.PUT("/members/:uid", h.group.UpdateMember)
		groupAdmin.DELETE("/members/:uid", h.group.RemoveMember)
		groupAdmin.POST("/invites", h.group.CreateInvite)
		groupAdmin.GET("/invites", h.group.ListInvites)
		groupAdmin.DELETE("/invites/:iid", h.group.RevokeInvite)
	}

	admin := v1.Group("/admin")
	admin.Use(m.auth.Authenticate())
	admin.Use(m.auth.RequireAdmin())
	{
		admin.POST("/users", h.user.CreateUser)
		admin.GET("/users", h.user.ListUsers)
		admin.DELETE("/users/:id", h.user.DeleteUser)
		admin.GET("/audit-logs", h.user.GetAuditLogs)

		admin.POST("/beta-invites", h.betaInvite.CreateBetaInvite)
		admin.GET("/beta-invites", h.betaInvite.ListBetaInvites)
		admin.GET("/beta-invites/:id", h.betaInvite.GetBetaInvite)
		admin.DELETE("/beta-invites/:id", h.betaInvite.DeleteBetaInvite)
	}

	router.GET("/images/:id", m.auth.Authenticate(), h.post.ServeImage)
	return router
}
