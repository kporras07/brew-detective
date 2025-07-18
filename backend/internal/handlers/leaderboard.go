package handlers

import (
	"context"
	"net/http"
	"sort"
	"time"

	"brew-detective-backend/internal/database"
	"brew-detective-backend/internal/models"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

// GetLeaderboard returns the top users ranked by points
func GetLeaderboard(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	iter := database.FirestoreClient.Collection(database.UsersCollection).
		Where("cases_count", ">", 0).
		Documents(ctx)

	var entries []models.LeaderboardEntry
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leaderboard"})
			return
		}

		var user models.User
		if err := doc.DataTo(&user); err != nil {
			continue
		}

		entry := models.LeaderboardEntry{
			UserID:        user.ID,
			DetectiveName: user.Name,
			Points:        user.Points,
			Accuracy:      user.Accuracy,
			CasesCount:    user.CasesCount,
			Badges:        user.Badges,
		}

		entries = append(entries, entry)
	}

	// Sort by points (descending)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Points > entries[j].Points
	})

	// Assign ranks
	for i := range entries {
		entries[i].Rank = i + 1
	}

	// Limit to top 50
	if len(entries) > 50 {
		entries = entries[:50]
	}

	c.JSON(http.StatusOK, gin.H{"leaderboard": entries})
}

// GetUserProfile returns user profile information
func GetUserProfile(c *gin.Context) {
	userID := c.Param("id")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	doc, err := database.FirestoreClient.Collection(database.UsersCollection).Doc(userID).Get(ctx)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var user models.User
	if err := doc.DataTo(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// UpdateUserProfile updates user profile information
func UpdateUserProfile(c *gin.Context) {
	userID := c.Param("id")
	
	var updates struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Level string `json:"level"`
	}
	
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid update data"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userRef := database.FirestoreClient.Collection(database.UsersCollection).Doc(userID)
	
	// Check if user exists
	doc, err := userRef.Get(ctx)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var user models.User
	if err := doc.DataTo(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user data"})
		return
	}

	// Update fields
	if updates.Name != "" {
		user.Name = updates.Name
	}
	if updates.Email != "" {
		user.Email = updates.Email
	}
	if updates.Level != "" {
		user.Level = updates.Level
	}
	user.UpdatedAt = time.Now()

	// Save updated user
	if _, err := userRef.Set(ctx, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully", "user": user})
}