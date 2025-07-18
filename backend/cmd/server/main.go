package main

import (
	"log"
	"os"

	"brew-detective-backend/internal/database"
	"brew-detective-backend/internal/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Firestore
	if err := database.InitFirestore(); err != nil {
		log.Fatalf("Failed to initialize Firestore: %v", err)
	}
	defer database.CloseFirestore()

	// Initialize Gin router
	router := gin.Default()

	// Configure CORS for GitHub Pages
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"https://kporras07.github.io", // Replace with your GitHub Pages domain
		"http://localhost:3000",       // For local development
		"http://localhost:8080",       // For local development
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	config.AllowCredentials = true

	router.Use(cors.New(config))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "brew-detective-backend"})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Coffee cases
		api.GET("/cases", handlers.GetCases)
		api.GET("/cases/:id", handlers.GetCaseByID)

		// Submissions
		api.POST("/submissions", handlers.SubmitCase)

		// Leaderboard
		api.GET("/leaderboard", handlers.GetLeaderboard)

		// User profile
		api.GET("/users/:id", handlers.GetUserProfile)
		api.PUT("/users/:id", handlers.UpdateUserProfile)

		// Orders
		api.POST("/orders", handlers.CreateOrder)
		api.GET("/orders/:id", handlers.GetOrder)
		api.PUT("/orders/:id/status", handlers.UpdateOrderStatus)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}