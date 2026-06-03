package main

import (
	"log"
	"os"

	"brew-detective-backend/internal/auth"
	"brew-detective-backend/internal/database"
	"brew-detective-backend/internal/handlers"
	"brew-detective-backend/internal/store"

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

	// Create store and handler
	s := store.NewFirestoreStore(database.FirestoreClient)
	a := &auth.GoogleAuthenticator{}
	h := handlers.NewHandler(s, a)

	// Initialize Gin router
	router := gin.Default()

	// Configure CORS for GitHub Pages
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"https://brewdetective.coffee",
		"http://localhost:3000",
		"http://localhost:8080",
		"http://127.0.0.1:8080",
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	config.AllowCredentials = true

	router.Use(cors.New(config))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "brew-detective-backend"})
	})

	// Auth routes
	authRoutes := router.Group("/auth")
	{
		authRoutes.GET("/google", h.GoogleLogin)
		authRoutes.GET("/google/callback", h.GoogleCallback)
		authRoutes.POST("/logout", h.Logout)
	}

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes
		api.GET("/cases/public", h.GetCasesPublic)
		api.GET("/cases/active/public", h.GetActiveCasePublic)
		api.GET("/cases/:id/public", h.GetCaseByIDPublic)
		api.GET("/leaderboard", h.GetLeaderboard)
		api.GET("/leaderboard/current", h.GetCurrentCaseLeaderboard)
		api.GET("/catalog", h.GetAllCatalog)
		api.GET("/catalog/:category", h.GetCatalogByCategory)

		// Protected routes
		protected := api.Group("/")
		protected.Use(auth.AuthMiddleware())
		{
			protected.GET("/profile", h.GetProfile)
			protected.GET("/users/:id", h.GetUserProfile)
			protected.PUT("/users/:id", h.UpdateUserProfile)

			protected.POST("/submissions", h.SubmitCase)
			protected.GET("/submissions", h.GetUserSubmissions)

			protected.POST("/orders", h.CreateOrder)
			protected.GET("/orders/:id", h.GetOrder)
			protected.PUT("/orders/:id/status", h.UpdateOrderStatus)
		}

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(auth.AdminMiddleware())
		{
			admin.GET("/catalog", h.GetAllCatalogItems)
			admin.POST("/catalog", h.CreateCatalogItem)
			admin.PUT("/catalog/:id", h.UpdateCatalogItem)
			admin.DELETE("/catalog/:id", h.DeleteCatalogItem)

			admin.GET("/cases", h.GetAllCases)
			admin.GET("/cases/active", h.GetActiveCase)
			admin.GET("/cases/:id", h.GetCaseByID)
			admin.GET("/cases/list", h.GetCases)
			admin.POST("/cases", h.CreateCase)
			admin.PUT("/cases/:id", h.UpdateCase)
			admin.DELETE("/cases/:id", h.DeleteCase)

			admin.GET("/orders", h.GetAllOrders)
			admin.POST("/orders", h.CreateOrder)
			admin.PUT("/orders/:id/status", h.UpdateOrderStatus)

			admin.GET("/users", h.GetAllUsers)
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
