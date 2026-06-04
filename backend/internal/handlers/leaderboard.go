package handlers

import (
	"net/http"
	"sort"
	"time"

	"brew-detective-backend/internal/models"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetLeaderboard(c *gin.Context) {
	users, err := h.Store.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leaderboard", "details": err.Error()})
		return
	}

	var entries []models.LeaderboardEntryWithUser
	for _, user := range users {
		if user.CasesAttempted > 0 || user.CasesSolved > 0 || user.Points > 0 {
			entry := models.LeaderboardEntryWithUser{
				UserID:        user.ID,
				DetectiveName: user.Name,
				Points:        user.Points,
				Accuracy:      user.Accuracy,
				CasesCount:    user.CasesCount,
				Badges:        user.Badges,
			}
			entries = append(entries, entry)
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Points > entries[j].Points
	})

	for i := range entries {
		entries[i].Rank = i + 1
	}

	if len(entries) > 50 {
		entries = entries[:50]
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": entries,
		"total_users": len(entries),
	})
}

func (h *Handler) GetUserProfile(c *gin.Context) {
	userID := c.Param("id")

	user, err := h.Store.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *Handler) UpdateUserProfile(c *gin.Context) {
	userID := c.Param("id")

	authenticatedUserID, exists := c.Get("userID")
	if !exists || authenticatedUserID.(string) != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own profile"})
		return
	}

	var updates struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid update data"})
		return
	}

	user, err := h.Store.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if updates.Name != "" {
		user.Name = updates.Name
	}
	if updates.Email != "" {
		user.Email = updates.Email
	}
	user.UpdatedAt = time.Now()

	if err := h.Store.SetUser(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully", "user": user})
}

func (h *Handler) GetAllUsers(c *gin.Context) {
	users, err := h.Store.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users", "details": err.Error()})
		return
	}

	var userResponses []map[string]interface{}
	for _, user := range users {
		userResponses = append(userResponses, map[string]interface{}{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"users": userResponses,
		"count": len(userResponses),
	})
}

func (h *Handler) GetCurrentCaseLeaderboard(c *gin.Context) {
	activeCase, err := h.Store.GetActiveCase(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active case found"})
		return
	}

	submissions, err := h.Store.ListSubmissionsByCase(c.Request.Context(), activeCase.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch submissions", "details": err.Error()})
		return
	}

	userScores := make(map[string]models.LeaderboardEntryWithUser)

	for _, submission := range submissions {
		entry, exists := userScores[submission.UserID]
		if !exists {
			user, err := h.Store.GetUser(c.Request.Context(), submission.UserID)
			if err != nil {
				continue
			}
			entry = models.LeaderboardEntryWithUser{
				UserID:        submission.UserID,
				DetectiveName: user.Name,
				Points:        0,
				Accuracy:      0,
				CasesCount:    0,
				Badges:        user.Badges,
			}
		}

		if submission.Score > entry.Points {
			entry.Points = submission.Score
			entry.Accuracy = submission.Accuracy
		}
		entry.CasesCount = 1

		userScores[submission.UserID] = entry
	}

	var entries []models.LeaderboardEntryWithUser
	for _, entry := range userScores {
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Points == entries[j].Points {
			return entries[i].Accuracy > entries[j].Accuracy
		}
		return entries[i].Points > entries[j].Points
	})

	for i := range entries {
		entries[i].Rank = i + 1
	}

	if len(entries) > 50 {
		entries = entries[:50]
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": entries,
		"total_users": len(entries),
		"case_id":     activeCase.ID,
		"case_name":   activeCase.Name,
	})
}
