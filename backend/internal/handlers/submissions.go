package handlers

import (
	"context"
	"net/http"
	"time"

	"brew-detective-backend/internal/database"
	"brew-detective-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SubmitCase handles case submission
func SubmitCase(c *gin.Context) {
	var submission models.Submission
	
	if err := c.ShouldBindJSON(&submission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid submission data"})
		return
	}

	// Generate submission ID and set timestamps
	submission.ID = uuid.New().String()
	submission.SubmittedAt = time.Now()
	
	// Calculate score and accuracy
	score, accuracy := calculateScore(&submission)
	submission.Score = score
	submission.Accuracy = accuracy

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Save submission to Firestore
	_, err := database.FirestoreClient.Collection(database.SubmissionsCollection).
		Doc(submission.ID).Set(ctx, submission)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save submission"})
		return
	}

	// Update user stats if user exists
	go updateUserStats(submission.UserID, score, accuracy)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Submission successful",
		"submission_id": submission.ID,
		"score": score,
		"accuracy": accuracy,
	})
}

// calculateScore calculates the score and accuracy for a submission
func calculateScore(submission *models.Submission) (int, float64) {
	// This is a simplified scoring system
	// In a real implementation, you'd fetch the correct answers from the case
	
	totalQuestions := len(submission.CoffeeAnswers) * 4 // 4 questions per coffee
	if totalQuestions == 0 {
		return 0, 0.0
	}
	
	correctAnswers := 0
	basePoints := 100
	
	for _, answer := range submission.CoffeeAnswers {
		// Simulate scoring - in reality, you'd compare with correct answers
		// Each coffee has 4 attributes: region, variety, process, roast_level
		if answer.Region != "" {
			correctAnswers++
		}
		if answer.Variety != "" {
			correctAnswers++
		}
		if answer.Process != "" {
			correctAnswers++
		}
		if answer.RoastLevel != "" {
			correctAnswers++
		}
	}
	
	accuracy := float64(correctAnswers) / float64(totalQuestions)
	score := int(float64(basePoints) * accuracy * float64(len(submission.CoffeeAnswers)))
	
	return score, accuracy
}

// updateUserStats updates user statistics
func updateUserStats(userID string, score int, accuracy float64) {
	if userID == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userRef := database.FirestoreClient.Collection(database.UsersCollection).Doc(userID)
	
	// Get current user data
	doc, err := userRef.Get(ctx)
	if err != nil {
		// User doesn't exist, create new user
		newUser := models.User{
			ID:         userID,
			Points:     score,
			CasesCount: 1,
			Accuracy:   accuracy,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		userRef.Set(ctx, newUser)
		return
	}

	var user models.User
	if err := doc.DataTo(&user); err != nil {
		return
	}

	// Update user stats
	user.Points += score
	user.CasesCount++
	user.Accuracy = (user.Accuracy*float64(user.CasesCount-1) + accuracy) / float64(user.CasesCount)
	user.UpdatedAt = time.Now()

	// Update badges based on achievements
	updateBadges(&user)

	userRef.Set(ctx, user)
}

// updateBadges updates user badges based on achievements
func updateBadges(user *models.User) {
	badges := make(map[string]bool)
	
	// Convert existing badges to map for easy lookup
	for _, badge := range user.Badges {
		badges[badge] = true
	}

	// First case badge
	if user.CasesCount >= 1 && !badges["🔍 Primer Caso"] {
		badges["🔍 Primer Caso"] = true
	}

	// Accuracy badges
	if user.Accuracy >= 0.7 && !badges["🎯 Precisión 70%"] {
		badges["🎯 Precisión 70%"] = true
	}
	if user.Accuracy >= 0.8 && !badges["💎 Catador Nivel 2"] {
		badges["💎 Catador Nivel 2"] = true
	}

	// Points badges
	if user.Points >= 2000 && !badges["🏆 Detective Maestro"] {
		badges["🏆 Detective Maestro"] = true
	}

	// Cases count badges
	if user.CasesCount >= 5 && !badges["🔥 Experto en Tuestes"] {
		badges["🔥 Experto en Tuestes"] = true
	}

	// Convert back to slice
	user.Badges = make([]string, 0, len(badges))
	for badge := range badges {
		user.Badges = append(user.Badges, badge)
	}
}