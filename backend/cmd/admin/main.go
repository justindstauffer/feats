package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jstauff/feats-api/internal/config"
	"github.com/jstauff/feats-api/internal/database"
	"github.com/jstauff/feats-api/internal/models"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

func main() {
	fmt.Println("=== Feats API - Create Admin User ===")
	fmt.Println()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Connect to database
	db, err := database.Connect(cfg.DatabasePath, false)
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		os.Exit(1)
	}

	// Run migrations
	if err := database.Migrate(db); err != nil {
		fmt.Printf("Error running migrations: %v\n", err)
		os.Exit(1)
	}

	// Seed activity types
	if err := database.Seed(db); err != nil {
		fmt.Printf("Error seeding database: %v\n", err)
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	// Get email
	fmt.Print("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(strings.ToLower(email))

	if email == "" {
		fmt.Println("Email is required")
		os.Exit(1)
	}

	// Check if user exists
	var existing models.User
	if err := db.Where("email = ?", email).First(&existing).Error; err == nil {
		fmt.Println("User with this email already exists")
		os.Exit(1)
	}

	// Get name
	fmt.Print("Name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	if name == "" {
		fmt.Println("Name is required")
		os.Exit(1)
	}

	// Get password
	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Printf("\nError reading password: %v\n", err)
		os.Exit(1)
	}
	fmt.Println()

	password := string(passwordBytes)
	if len(password) < 12 {
		fmt.Println("Password must be at least 12 characters")
		os.Exit(1)
	}

	// Confirm password
	fmt.Print("Confirm Password: ")
	confirmBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Printf("\nError reading password: %v\n", err)
		os.Exit(1)
	}
	fmt.Println()

	if password != string(confirmBytes) {
		fmt.Println("Passwords do not match")
		os.Exit(1)
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cfg.BcryptCost)
	if err != nil {
		fmt.Printf("Error hashing password: %v\n", err)
		os.Exit(1)
	}

	// Create user
	now := time.Now()
	user := models.User{
		ID:                  uuid.New().String(),
		Email:               email,
		PasswordHash:        string(hash),
		Name:                name,
		Role:                models.RoleAdmin,
		PasswordChangedAt:   now,
		ForcePasswordChange: false,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := db.Create(&user).Error; err != nil {
		fmt.Printf("Error creating user: %v\n", err)
		os.Exit(1)
	}

	// Create streak record
	streak := models.Streak{
		ID:            uuid.New().String(),
		UserID:        user.ID,
		CurrentStreak: 0,
		LongestStreak: 0,
		UpdatedAt:     now,
	}
	db.Create(&streak)

	fmt.Println()
	fmt.Println("=== Admin user created successfully! ===")
	fmt.Printf("ID:    %s\n", user.ID)
	fmt.Printf("Email: %s\n", user.Email)
	fmt.Printf("Name:  %s\n", user.Name)
	fmt.Printf("Role:  %s\n", user.Role)
}
