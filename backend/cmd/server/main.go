package main

import (
	"log"
	"os"

	"brew-detective-backend/internal/auth"
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

	// Initialize Auth
	auth.InitAuth()

	// Initialize Gin router
	router := gin.Default()

	// Configure CORS for GitHub Pages
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"https://kporras07.github.io", // Replace with your GitHub Pages domain
		"http://localhost:3000",       // For local development
		"http://localhost:8080",       // For local development
		"http://127.0.0.1:8080",       // Alternative localhost
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	config.AllowCredentials = true

	router.Use(cors.New(config))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "brew-detective-backend"})
	})

	// Firestore test endpoint
	router.GET("/test/firestore", handlers.TestFirestore)

	// Auth routes
	authRoutes := router.Group("/auth")
	{
		authRoutes.GET("/google", handlers.GoogleLogin)
		authRoutes.GET("/google/callback", handlers.GoogleCallback)
		authRoutes.POST("/logout", handlers.Logout)
	}

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes
		api.GET("/cases", handlers.GetCases)
		api.GET("/cases/active", handlers.GetActiveCase)
		api.GET("/cases/:id", handlers.GetCaseByID)
		api.GET("/leaderboard", handlers.GetLeaderboard)
		api.GET("/catalog", handlers.GetAllCatalog)
		api.GET("/catalog/:category", handlers.GetCatalogByCategory)

		// Protected routes
		protected := api.Group("/")
		protected.Use(auth.AuthMiddleware())
		{
			// User profile
			protected.GET("/profile", handlers.GetProfile)
			protected.GET("/users/:id", handlers.GetUserProfile)
			protected.PUT("/users/:id", handlers.UpdateUserProfile)

			// Submissions
			protected.POST("/submissions", handlers.SubmitCase)
			protected.GET("/submissions", handlers.GetUserSubmissions)

			// Orders
			protected.POST("/orders", handlers.CreateOrder)
			protected.GET("/orders/:id", handlers.GetOrder)
			protected.PUT("/orders/:id/status", handlers.UpdateOrderStatus)
		}

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(auth.AdminMiddleware())
		{
			// Catalog management
			admin.GET("/catalog", handlers.GetAllCatalogItems)
			admin.POST("/catalog", handlers.CreateCatalogItem)
			admin.PUT("/catalog/:id", handlers.UpdateCatalogItem)
			admin.DELETE("/catalog/:id", handlers.DeleteCatalogItem)

			// Case management
			admin.GET("/cases", handlers.GetAllCases)
			admin.GET("/cases/:id", handlers.GetCaseByID)
			admin.POST("/cases", handlers.CreateCase)
			admin.PUT("/cases/:id", handlers.UpdateCase)
			admin.DELETE("/cases/:id", handlers.DeleteCase)
		}
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