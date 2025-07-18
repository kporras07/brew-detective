package handlers

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"brew-detective-backend/internal/auth"
	"brew-detective-backend/internal/database"
	"brew-detective-backend/internal/models"

	"github.com/gin-gonic/gin"
)

func GoogleLogin(c *gin.Context) {
	// Use a simple fixed state for now (in production, use proper state management)
	state := "brew-detective-state"
	url := auth.GetGoogleOauthConfig().AuthCodeURL(state)
	c.JSON(http.StatusOK, gin.H{"auth_url": url})
}

func GoogleCallback(c *gin.Context) {
	queryState := c.Query("state")
	expectedState := "brew-detective-state"
	
	if queryState != expectedState {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "OAuth state mismatch", 
			"expected": expectedState,
			"received": queryState,
		})
		return
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code not found"})
		return
	}

	googleUser, err := auth.GetUserDataFromGoogle(code)
	if err != nil {
		fmt.Printf("Error getting user data from Google: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user data", "details": err.Error()})
		return
	}

	// Check if user exists in database
	var user *models.User
	userRef := database.FirestoreClient.Collection(database.UsersCollection).Doc(googleUser.ID)
	doc, err := userRef.Get(c.Request.Context())

	if err != nil || !doc.Exists() {
		// Create new user
		user = &models.User{
			ID:             googleUser.ID,
			Email:          googleUser.Email,
			Name:           googleUser.Name,
			Picture:        googleUser.Picture,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			Score:          0,
			Level:          "Rookie Detective",
			Badges:         []string{},
			CasesAttempted: 0,
			CasesSolved:    0,
		}

		_, err = userRef.Set(c.Request.Context(), user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user", "details": err.Error()})
			return
		}
	} else {
		// Update existing user
		if err := doc.DataTo(&user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user data"})
			return
		}

		// Update user info from Google
		user.Email = googleUser.Email
		user.Name = googleUser.Name
		user.Picture = googleUser.Picture
		user.UpdatedAt = time.Now()

		_, err = userRef.Set(c.Request.Context(), user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}
	}

	// Generate JWT token
	token, err := auth.GenerateJWT(user.ID, user.Email, user.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Redirect back to frontend with token in URL fragment
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:8080" // Default for local development
	}
	redirectURL := fmt.Sprintf("%s/#token=%s", frontendURL, token)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	userRef := database.FirestoreClient.Collection(database.UsersCollection).Doc(userID.(string))
	doc, err := userRef.Get(c.Request.Context())

	if err != nil || !doc.Exists() {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var user models.User
	if err := doc.DataTo(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user data"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func Logout(c *gin.Context) {
	// For JWT tokens, logout is handled client-side by removing the token
	// We could implement a token blacklist here if needed
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
