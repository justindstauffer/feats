package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jstauff/feats-api/internal/config"
	"github.com/jstauff/feats-api/internal/database"
	"github.com/jstauff/feats-api/internal/handlers"
	"github.com/jstauff/feats-api/internal/middleware"
	"github.com/jstauff/feats-api/internal/services"
	"github.com/jstauff/feats-api/internal/websocket"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate required config
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	if len(cfg.JWTSecret) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 characters")
	}

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Connect to database
	debug := cfg.GinMode == "debug"
	db, err := database.Connect(cfg.DatabasePath, debug)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Create indexes
	if err := database.CreateIndexes(db); err != nil {
		log.Fatalf("Failed to create indexes: %v", err)
	}

	// Seed default data
	if err := database.Seed(db); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	// Ensure storage directories exist
	if err := os.MkdirAll(cfg.StoragePath+"/images", 0755); err != nil {
		log.Fatalf("Failed to create storage directory: %v", err)
	}
	if err := os.MkdirAll(cfg.StoragePath+"/profiles", 0755); err != nil {
		log.Fatalf("Failed to create profiles directory: %v", err)
	}

	// Initialize services
	authService := services.NewAuthService(db, cfg)
	userService := services.NewUserService(db, cfg)
	betaInviteService := services.NewBetaInviteService(db)
	groupService := services.NewGroupService(db, cfg)
	postService := services.NewPostService(db, cfg)
	activityService := services.NewActivityService(db)
	reactionService := services.NewReactionService(db)
	commentService := services.NewCommentService(db)
	streakService := services.NewStreakService(db, cfg)
	challengeService := services.NewChallengeService(db)
	goalService := services.NewGoalService(db, cfg)
	auditService := services.NewAuditService(db)

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, userService, betaInviteService, auditService, cfg)
	userHandler := handlers.NewUserHandler(userService, auditService, authService)
	groupHandler := handlers.NewGroupHandler(groupService, auditService, wsHub)
	betaInviteHandler := handlers.NewBetaInviteHandler(betaInviteService)
	postHandler := handlers.NewPostHandler(postService, streakService, challengeService, goalService, auditService, cfg, wsHub)
	activityHandler := handlers.NewActivityHandler(activityService)
	reactionHandler := handlers.NewReactionHandler(reactionService, wsHub)
	commentHandler := handlers.NewCommentHandler(commentService, wsHub)
	streakHandler := handlers.NewStreakHandler(streakService)
	challengeHandler := handlers.NewChallengeHandler(challengeService, wsHub)
	goalHandler := handlers.NewGoalHandler(goalService)
	wsHandler := handlers.NewWebSocketHandler(wsHub, authService, groupService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService, cfg)
	groupMiddleware := middleware.NewGroupMiddleware(groupService)
	rateLimiter := middleware.NewRateLimiter(cfg)
	securityMiddleware := middleware.NewSecurityMiddleware()

	// Create router
	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(securityMiddleware.SecurityHeaders())
	router.Use(middleware.CORS(cfg))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// WebSocket endpoint (authentication via query param)
	router.GET("/ws", wsHandler.HandleWebSocket)

	// API v1 routes
	v1 := router.Group("/api/v1")

	// Public routes (with rate limiting)
	auth := v1.Group("/auth")
	auth.Use(rateLimiter.LoginRateLimit())
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/password/reset-request", authHandler.RequestPasswordReset)
		auth.POST("/password/reset", authHandler.ResetPassword)
	}

	// Protected routes (user-level, not group-specific)
	protected := v1.Group("")
	protected.Use(authMiddleware.Authenticate())
	protected.Use(rateLimiter.APIRateLimit())
	{
		// Auth
		protected.POST("/auth/logout", authHandler.Logout)
		protected.POST("/auth/password/change", authHandler.ChangePassword)

		// Users (profile management - not group-scoped)
		protected.GET("/users/me", userHandler.GetCurrentUser)
		protected.PUT("/users/me", userHandler.UpdateCurrentUser)
		protected.GET("/users/:id", userHandler.GetUser)

		// Device tokens (for push notifications)
		protected.POST("/devices", authHandler.RegisterDevice)
		protected.DELETE("/devices/:token", authHandler.UnregisterDevice)

		// Invite redemption (outside group context)
		protected.POST("/invites/redeem", rateLimiter.InviteRedeemRateLimit(), groupHandler.RedeemInvite)
	}

	// Group management routes (requires authentication but not group membership)
	groups := v1.Group("/groups")
	groups.Use(authMiddleware.Authenticate())
	groups.Use(rateLimiter.APIRateLimit())
	{
		groups.POST("", groupHandler.CreateGroup)
		groups.GET("", groupHandler.ListGroups)
	}

	// Group-scoped routes (requires group membership)
	group := v1.Group("/groups/:gid")
	group.Use(authMiddleware.Authenticate())
	group.Use(groupMiddleware.RequireGroupMember())
	group.Use(rateLimiter.APIRateLimit())
	{
		// Group info
		group.GET("", groupHandler.GetGroup)
		group.POST("/leave", groupHandler.LeaveGroup)
		group.GET("/members", groupHandler.ListMembers)

		// User data within group context
		group.GET("/users/:id/streak", streakHandler.GetUserStreak)
		group.GET("/users/:id/goals", goalHandler.GetUserGoals)

		// Activities (system-wide + group custom)
		group.GET("/activities", activityHandler.ListActivities)
		group.POST("/activities", activityHandler.CreateActivity)
		group.DELETE("/activities/:id", activityHandler.DeleteActivity)

		// Posts
		group.GET("/posts", postHandler.ListPosts)
		group.POST("/posts", rateLimiter.PostRateLimit(), postHandler.CreatePost)
		group.GET("/posts/:id", postHandler.GetPost)
		group.PUT("/posts/:id", postHandler.UpdatePost)
		group.DELETE("/posts/:id", postHandler.DeletePost)
		group.POST("/posts/:id/images", rateLimiter.UploadRateLimit(), postHandler.UploadImage)
		group.DELETE("/posts/:id/images/:image_id", postHandler.DeleteImage)

		// Reactions
		group.GET("/posts/:id/reactions", reactionHandler.GetReactions)
		group.POST("/posts/:id/reactions", reactionHandler.AddReaction)
		group.DELETE("/posts/:id/reactions", reactionHandler.RemoveReaction)

		// Comments
		group.GET("/posts/:id/comments", commentHandler.GetComments)
		group.POST("/posts/:id/comments", commentHandler.CreateComment)
		group.PUT("/comments/:id", commentHandler.UpdateComment)
		group.DELETE("/comments/:id", commentHandler.DeleteComment)

		// Streaks
		group.GET("/streaks/leaderboard", streakHandler.GetLeaderboard)

		// Challenges
		group.GET("/challenges", challengeHandler.ListChallenges)
		group.POST("/challenges", challengeHandler.CreateChallenge)
		group.GET("/challenges/:id", challengeHandler.GetChallenge)
		group.POST("/challenges/:id/join", challengeHandler.JoinChallenge)
		group.DELETE("/challenges/:id/leave", challengeHandler.LeaveChallenge)
		group.DELETE("/challenges/:id", challengeHandler.DeleteChallenge)

		// Goals
		group.POST("/goals", goalHandler.CreateGoal)
		group.PUT("/goals/:id", goalHandler.UpdateGoal)
		group.DELETE("/goals/:id", goalHandler.DeleteGoal)
	}

	// Group admin routes (requires group admin)
	groupAdmin := v1.Group("/groups/:gid")
	groupAdmin.Use(authMiddleware.Authenticate())
	groupAdmin.Use(groupMiddleware.RequireGroupAdmin())
	groupAdmin.Use(rateLimiter.APIRateLimit())
	{
		groupAdmin.PUT("", groupHandler.UpdateGroup)
		groupAdmin.DELETE("", groupHandler.DeleteGroup)
		groupAdmin.PUT("/members/:uid", groupHandler.UpdateMember)
		groupAdmin.DELETE("/members/:uid", groupHandler.RemoveMember)
		groupAdmin.POST("/invites", groupHandler.CreateInvite)
		groupAdmin.GET("/invites", groupHandler.ListInvites)
		groupAdmin.DELETE("/invites/:iid", groupHandler.RevokeInvite)
	}

	// System admin routes
	admin := v1.Group("/admin")
	admin.Use(authMiddleware.Authenticate())
	admin.Use(authMiddleware.RequireAdmin())
	{
		admin.POST("/users", userHandler.CreateUser)
		admin.GET("/users", userHandler.ListUsers)
		admin.DELETE("/users/:id", userHandler.DeleteUser)
		admin.GET("/audit-logs", userHandler.GetAuditLogs)

		// Beta invites
		admin.POST("/beta-invites", betaInviteHandler.CreateBetaInvite)
		admin.GET("/beta-invites", betaInviteHandler.ListBetaInvites)
		admin.GET("/beta-invites/:id", betaInviteHandler.GetBetaInvite)
		admin.DELETE("/beta-invites/:id", betaInviteHandler.DeleteBetaInvite)
	}

	// Image serving (protected)
	router.GET("/images/:id", authMiddleware.Authenticate(), postHandler.ServeImage)

	// Start server
	log.Printf("Starting server on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
