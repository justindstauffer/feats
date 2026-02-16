package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jstauff/feats-api/internal/config"
	"github.com/jstauff/feats-api/internal/database"
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
	if err := initStorage(cfg); err != nil {
		log.Fatalf("Failed to create storage directory: %v", err)
	}
	appServices := initServices(db, cfg)

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	appHandlers := initHandlers(appServices, cfg, wsHub)
	appMiddleware := initMiddleware(appServices.auth, appServices.group, cfg)
	router := setupRouter(cfg, appHandlers, appMiddleware)

	// Start server
	log.Printf("Starting server on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
