package handlers

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"brew-detective-backend/internal/models"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GoogleLogin(c *gin.Context) {
	url := h.Auth.GenerateOAuthURL()
	c.JSON(http.StatusOK, gin.H{"auth_url": url})
}

func (h *Handler) GoogleCallback(c *gin.Context) {
	queryState := c.Query("state")

	// In a production app with sessions, you'd validate against stored state
	if queryState == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OAuth state parameter missing"})
		return
	}

	// Basic state validation - ensure it's a reasonable hex string
	if len(queryState) < 16 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OAuth state format"})
		return
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code not found"})
		return
	}

	googleUser, err := h.Auth.GetUserFromOAuthCode(code)
	if err != nil {
		fmt.Printf("Error getting user data from Google: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user data", "details": err.Error()})
		return
	}

	// Check if user exists in database
	user, err := h.Store.GetUser(c.Request.Context(), googleUser.ID)
	if err != nil {
		// Create new user
		user = &models.User{
			ID:             googleUser.ID,
			Email:          googleUser.Email,
			Name:           googleUser.Name,
			Picture:        googleUser.Picture,
			Type:           "regular", // Default to regular user
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			Score:          0,
			Badges:         []string{},
			CasesAttempted: 0,
			CasesSolved:    0,
		}

		if err := h.Store.SetUser(c.Request.Context(), user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user", "details": err.Error()})
			return
		}
	} else {
		// Update existing user info from Google (preserve Type and Name fields)
		userType := user.Type
		userName := user.Name
		user.Email = googleUser.Email
		// Only update name from Google if user hasn't set a custom name
		if userName == "" || userName == googleUser.Name {
			user.Name = googleUser.Name
		} else {
			user.Name = userName // Keep custom name
		}
		user.Picture = googleUser.Picture
		user.UpdatedAt = time.Now()

		// Ensure type is preserved (set default if empty)
		if userType != "" {
			user.Type = userType
		} else if user.Type == "" {
			user.Type = "regular"
		}

		if err := h.Store.SetUser(c.Request.Context(), user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}
	}

	// Generate JWT token
	token, err := h.Auth.GenerateJWT(user.ID, user.Email, user.Name)
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

func (h *Handler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	user, err := h.Store.GetUser(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *Handler) Logout(c *gin.Context) {
	// For JWT tokens, logout is handled client-side by removing the token
	// We could implement a token blacklist here if needed
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
