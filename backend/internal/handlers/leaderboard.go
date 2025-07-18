package handlers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
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

	// First, try to get all users (remove the where clause since fields might not exist yet)
	iter := database.FirestoreClient.Collection(database.UsersCollection).Documents(ctx)

	var entries []models.LeaderboardEntry
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch leaderboard", 
				"details": err.Error(),
			})
			return
		}

		var user models.User
		if err := doc.DataTo(&user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to parse user data", 
				"details": err.Error(),
				"document_id": doc.Ref.ID,
			})
			return
		}

		// Only include users who have attempted at least one case
		if user.CasesAttempted > 0 || user.CasesSolved > 0 || user.Score > 0 {
			entry := models.LeaderboardEntry{
				UserID:        user.ID,
				DetectiveName: user.Name,
				Points:        user.Score, // Use Score instead of Points
				Accuracy:      user.Accuracy,
				CasesCount:    user.CasesAttempted, // Use CasesAttempted
				Badges:        user.Badges,
			}
			entries = append(entries, entry)
		}
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

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": entries,
		"total_users": len(entries),
	})
}

// TestFirestore tests if Firestore connection is working
func TestFirestore(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Show what credentials are being used
	response := gin.H{
		"status": "testing",
		"project_id": os.Getenv("GOOGLE_CLOUD_PROJECT"),
		"database_id": os.Getenv("FIRESTORE_DATABASE_ID"),
		"credentials_file": os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
	}

	// Try to read and show the credentials being used
	if credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); credentialsPath != "" {
		if data, err := ioutil.ReadFile(credentialsPath); err != nil {
			response["credentials_read_error"] = err.Error()
		} else {
			var creds map[string]interface{}
			if err := json.Unmarshal(data, &creds); err != nil {
				response["credentials_parse_error"] = err.Error()
			} else {
				response["authenticated_user"] = gin.H{
					"type": creds["type"],
					"client_email": creds["client_email"],
					"project_id": creds["project_id"],
				}
			}
		}
	} else {
		response["auth_method"] = "Application Default Credentials"
		
		// Try to read the default credentials file
		homeDir, _ := os.UserHomeDir()
		defaultCredsPath := homeDir + "/.config/gcloud/application_default_credentials.json"
		if data, err := ioutil.ReadFile(defaultCredsPath); err == nil {
			var creds map[string]interface{}
			if json.Unmarshal(data, &creds) == nil {
				response["authenticated_user"] = gin.H{
					"type": creds["type"],
					"client_id": creds["client_id"],
					"quota_project_id": creds["quota_project_id"],
				}
			}
		}
	}

	// Try to list collections
	collections := database.FirestoreClient.Collections(ctx)
	var collectionNames []string
	
	for {
		collection, err := collections.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			response["error"] = "Failed to list collections"
			response["details"] = err.Error()
			c.JSON(http.StatusInternalServerError, response)
			return
		}
		collectionNames = append(collectionNames, collection.ID)
	}

	// Try to count users
	userQuery := database.FirestoreClient.Collection(database.UsersCollection).Limit(1)
	userDocs, err := userQuery.Documents(ctx).GetAll()
	
	response["status"] = "connected"
	response["collections"] = collectionNames
	response["user_count_sample"] = len(userDocs)
	
	if err != nil {
		response["user_query_error"] = err.Error()
	}

	c.JSON(http.StatusOK, response)
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