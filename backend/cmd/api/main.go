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
	postService := services.NewPostService(db, cfg)
	activityService := services.NewActivityService(db)
	reactionService := services.NewReactionService(db)
	commentService := services.NewCommentService(db)
	streakService := services.NewStreakService(db, cfg)
	challengeService := services.NewChallengeService(db)
	goalService := services.NewGoalService(db, cfg)
	auditService := services.NewAuditService(db)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, auditService, cfg)
	userHandler := handlers.NewUserHandler(userService, auditService)
	postHandler := handlers.NewPostHandler(postService, streakService, challengeService, goalService, auditService)
	activityHandler := handlers.NewActivityHandler(activityService)
	reactionHandler := handlers.NewReactionHandler(reactionService)
	commentHandler := handlers.NewCommentHandler(commentService)
	streakHandler := handlers.NewStreakHandler(streakService)
	challengeHandler := handlers.NewChallengeHandler(challengeService)
	goalHandler := handlers.NewGoalHandler(goalService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService, cfg)
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

	// API v1 routes
	v1 := router.Group("/api/v1")

	// Public routes (with rate limiting)
	auth := v1.Group("/auth")
	auth.Use(rateLimiter.LoginRateLimit())
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/password/reset-request", authHandler.RequestPasswordReset)
		auth.POST("/password/reset", authHandler.ResetPassword)
	}

	// Protected routes
	protected := v1.Group("")
	protected.Use(authMiddleware.Authenticate())
	protected.Use(rateLimiter.APIRateLimit())
	{
		// Auth
		protected.POST("/auth/logout", authHandler.Logout)
		protected.POST("/auth/password/change", authHandler.ChangePassword)

		// Users
		protected.GET("/users/me", userHandler.GetCurrentUser)
		protected.PUT("/users/me", userHandler.UpdateCurrentUser)
		protected.GET("/users/:id", userHandler.GetUser)
		protected.GET("/users/:id/streak", streakHandler.GetUserStreak)
		protected.GET("/users/:id/goals", goalHandler.GetUserGoals)

		// Activities
		protected.GET("/activities", activityHandler.ListActivities)
		protected.POST("/activities", activityHandler.CreateActivity)
		protected.DELETE("/activities/:id", activityHandler.DeleteActivity)

		// Posts
		protected.GET("/posts", postHandler.ListPosts)
		protected.POST("/posts", rateLimiter.PostRateLimit(), postHandler.CreatePost)
		protected.GET("/posts/:id", postHandler.GetPost)
		protected.PUT("/posts/:id", postHandler.UpdatePost)
		protected.DELETE("/posts/:id", postHandler.DeletePost)
		protected.POST("/posts/:id/images", rateLimiter.UploadRateLimit(), postHandler.UploadImage)
		protected.DELETE("/posts/:id/images/:image_id", postHandler.DeleteImage)

		// Reactions
		protected.GET("/posts/:id/reactions", reactionHandler.GetReactions)
		protected.POST("/posts/:id/reactions", reactionHandler.AddReaction)
		protected.DELETE("/posts/:id/reactions", reactionHandler.RemoveReaction)

		// Comments
		protected.GET("/posts/:id/comments", commentHandler.GetComments)
		protected.POST("/posts/:id/comments", commentHandler.CreateComment)
		protected.PUT("/comments/:id", commentHandler.UpdateComment)
		protected.DELETE("/comments/:id", commentHandler.DeleteComment)

		// Streaks
		protected.GET("/streaks/leaderboard", streakHandler.GetLeaderboard)

		// Challenges
		protected.GET("/challenges", challengeHandler.ListChallenges)
		protected.POST("/challenges", challengeHandler.CreateChallenge)
		protected.GET("/challenges/:id", challengeHandler.GetChallenge)
		protected.POST("/challenges/:id/join", challengeHandler.JoinChallenge)
		protected.DELETE("/challenges/:id/leave", challengeHandler.LeaveChallenge)
		protected.DELETE("/challenges/:id", challengeHandler.DeleteChallenge)

		// Goals
		protected.POST("/goals", goalHandler.CreateGoal)
		protected.PUT("/goals/:id", goalHandler.UpdateGoal)
		protected.DELETE("/goals/:id", goalHandler.DeleteGoal)

		// Device tokens (for push notifications)
		protected.POST("/devices", authHandler.RegisterDevice)
		protected.DELETE("/devices/:token", authHandler.UnregisterDevice)
	}

	// Admin routes
	admin := v1.Group("/admin")
	admin.Use(authMiddleware.Authenticate())
	admin.Use(authMiddleware.RequireAdmin())
	{
		admin.POST("/users", userHandler.CreateUser)
		admin.GET("/users", userHandler.ListUsers)
		admin.DELETE("/users/:id", userHandler.DeleteUser)
		admin.GET("/audit-logs", userHandler.GetAuditLogs)
	}

	// Image serving (protected)
	router.GET("/images/:id", authMiddleware.Authenticate(), postHandler.ServeImage)

	// Start server
	log.Printf("Starting server on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
