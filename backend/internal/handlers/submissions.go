package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"brew-detective-backend/internal/database"
	"brew-detective-backend/internal/models"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

// SubmitCase handles case submission
func SubmitCase(c *gin.Context) {
	var submission models.Submission

	if err := c.ShouldBindJSON(&submission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid submission data"})
		return
	}

	// Validate required fields
	if submission.OrderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}

	// Get the current active case
	activeCase, err := getActiveCase()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No active case available", "details": err.Error()})
		return
	}
	submission.CaseID = activeCase.ID

	// Validate order ID
	if !validateOrderID(submission.OrderID) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or already used order ID"})
		return
	}

	// Get user ID from auth context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User authentication required"})
		return
	}
	submission.UserID = userID.(string)

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
	_, err = database.FirestoreClient.Collection(database.SubmissionsCollection).
		Doc(submission.ID).Set(ctx, submission)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save submission"})
		return
	}

	// Mark order ID as used
	markOrderIDAsUsed(submission.OrderID, submission.UserID)

	// Update user stats if user exists
	go updateUserStats(submission.UserID, score, accuracy)

	c.JSON(http.StatusCreated, gin.H{
		"message":       "Submission successful",
		"submission_id": submission.ID,
		"score":         score,
		"accuracy":      accuracy,
	})
}

// calculateScore calculates the score and accuracy for a submission
func calculateScore(submission *models.Submission) (int, float64) {
	// Get the active case to determine enabled questions
	activeCase, err := getActiveCase()
	if err != nil {
		// Fallback to default scoring if case not found
		return calculateScoreDefault(submission)
	}

	// Count enabled questions per coffee
	enabledQuestionsPerCoffee := 0
	if activeCase.EnabledQuestions.Region {
		enabledQuestionsPerCoffee++
	}
	if activeCase.EnabledQuestions.Variety {
		enabledQuestionsPerCoffee++
	}
	if activeCase.EnabledQuestions.Process {
		enabledQuestionsPerCoffee++
	}
	if activeCase.EnabledQuestions.TasteNote1 {
		enabledQuestionsPerCoffee++
	}
	if activeCase.EnabledQuestions.TasteNote2 {
		enabledQuestionsPerCoffee++
	}

	totalQuestions := len(submission.CoffeeAnswers) * enabledQuestionsPerCoffee
	if totalQuestions == 0 {
		return 0, 0.0
	}

	correctAnswers := 0
	basePoints := 100

	for _, answer := range submission.CoffeeAnswers {
		// Find the correct coffee data for this answer
		var correctCoffee *models.CoffeeItem
		for _, coffee := range activeCase.Coffees {
			if coffee.ID == answer.CoffeeID {
				correctCoffee = &coffee
				break
			}
		}
		
		if correctCoffee == nil {
			continue // Skip if coffee not found
		}
		
		// Compare user answers against correct coffee data
		if activeCase.EnabledQuestions.Region && answer.Region != "" {
			if strings.EqualFold(strings.TrimSpace(answer.Region), strings.TrimSpace(correctCoffee.Region)) {
				correctAnswers++
			}
		}
		if activeCase.EnabledQuestions.Variety && answer.Variety != "" {
			if strings.EqualFold(strings.TrimSpace(answer.Variety), strings.TrimSpace(correctCoffee.Variety)) {
				correctAnswers++
			}
		}
		if activeCase.EnabledQuestions.Process && answer.Process != "" {
			if strings.EqualFold(strings.TrimSpace(answer.Process), strings.TrimSpace(correctCoffee.Process)) {
				correctAnswers++
			}
		}
		
		// Handle comma-separated tasting notes (avoid double points for same note)
		var awardedTastingNotes []string
		
		if activeCase.EnabledQuestions.TasteNote1 && answer.TasteNote1 != "" {
			if matchedNote := getMatchedTastingNote(answer.TasteNote1, correctCoffee.TastingNotes); matchedNote != "" {
				awardedTastingNotes = append(awardedTastingNotes, matchedNote)
				correctAnswers++
			}
		}
		
		if activeCase.EnabledQuestions.TasteNote2 && answer.TasteNote2 != "" {
			if matchedNote := getMatchedTastingNote(answer.TasteNote2, correctCoffee.TastingNotes); matchedNote != "" {
				// Check if we already awarded points for this exact note
				alreadyAwarded := false
				for _, awarded := range awardedTastingNotes {
					if strings.EqualFold(awarded, matchedNote) {
						alreadyAwarded = true
						break
					}
				}
				if !alreadyAwarded {
					correctAnswers++
				}
			}
		}
	}

	// Add bonus points for non-coffee questions
	bonusPoints := 0
	if activeCase.EnabledQuestions.FavoriteCoffee && submission.FavoriteCoffee != "" {
		bonusPoints += 50
	}
	if activeCase.EnabledQuestions.BrewingMethod && submission.BrewingMethod != "" {
		bonusPoints += 50
	}

	accuracy := float64(correctAnswers) / float64(totalQuestions)
	score := int(float64(basePoints) * accuracy * float64(len(submission.CoffeeAnswers))) + bonusPoints

	return score, accuracy
}

// calculateScoreDefault provides fallback scoring when case info is not available
func calculateScoreDefault(submission *models.Submission) (int, float64) {
	totalQuestions := len(submission.CoffeeAnswers) * 3 // region, variety, process
	if totalQuestions == 0 {
		return 0, 0.0
	}

	correctAnswers := 0
	basePoints := 100

	for _, answer := range submission.CoffeeAnswers {
		if answer.Region != "" {
			correctAnswers++
		}
		if answer.Variety != "" {
			correctAnswers++
		}
		if answer.Process != "" {
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
		// User doesn't exist - submissions should only work for existing users
		return
	}

	if !doc.Exists() {
		// User doesn't exist - submissions should only work for existing users
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

// validateOrderID validates if an order ID is valid and unused
func validateOrderID(orderID string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Query orders collection to find order with this order ID
	query := database.FirestoreClient.Collection(database.OrdersCollection).
		Where("order_id", "==", orderID).
		Limit(1)

	docs, err := query.Documents(ctx).GetAll()
	if err != nil || len(docs) == 0 {
		return false // Order ID doesn't exist
	}

	var order models.Order
	if err := docs[0].DataTo(&order); err != nil {
		return false
	}

	// Check if already used for submission
	if order.IsSubmissionUsed {
		return false
	}

	// Order must be in delivered status to allow submission
	if order.Status != "delivered" {
		return false
	}

	return true
}

// markOrderIDAsUsed marks an order ID as used for submission
func markOrderIDAsUsed(orderID string, userID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Query orders collection to find order with this order ID
	query := database.FirestoreClient.Collection(database.OrdersCollection).
		Where("order_id", "==", orderID).
		Limit(1)

	docs, err := query.Documents(ctx).GetAll()
	if err != nil || len(docs) == 0 {
		return // Order not found
	}

	now := time.Now()
	updates := []firestore.Update{
		{Path: "is_submission_used", Value: true},
		{Path: "submission_used_by", Value: userID},
		{Path: "submission_used_at", Value: now},
		{Path: "updated_at", Value: now},
	}

	docs[0].Ref.Update(ctx, updates)
}

// getActiveCase gets the current active case
func getActiveCase() (*models.CoffeeCase, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	iter := database.FirestoreClient.Collection(database.CasesCollection).
		Where("is_active", "==", true).
		Limit(1).
		Documents(ctx)

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	var coffeeCase models.CoffeeCase
	if err := doc.DataTo(&coffeeCase); err != nil {
		return nil, err
	}

	return &coffeeCase, nil
}

// GetUserSubmissions returns submissions for a specific user
func GetUserSubmissions(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User authentication required"})
		return
	}

	// Get query parameters for pagination
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit := 10 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Query submissions for this user, ordered by most recent first
	query := database.FirestoreClient.Collection(database.SubmissionsCollection).
		Where("user_id", "==", userID.(string)).
		OrderBy("submitted_at", firestore.Desc).
		Limit(limit).
		Offset(offset)

	iter := query.Documents(ctx)

	var submissions []map[string]interface{}
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch submissions", "details": err.Error()})
			return
		}

		var submission models.Submission
		if err := doc.DataTo(&submission); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse submission"})
			return
		}

		// Get case information for this submission
		caseRef := database.FirestoreClient.Collection(database.CasesCollection).Doc(submission.CaseID)
		caseDoc, err := caseRef.Get(ctx)
		var caseName string
		if err == nil && caseDoc.Exists() {
			var coffeeCase models.CoffeeCase
			if err := caseDoc.DataTo(&coffeeCase); err == nil {
				caseName = coffeeCase.Name
			}
		}
		if caseName == "" {
			caseName = "Caso Desconocido"
		}

		// Create response object with submission and case info
		submissionResponse := map[string]interface{}{
			"id":           submission.ID,
			"case_id":      submission.CaseID,
			"case_name":    caseName,
			"score":        submission.Score,
			"accuracy":     submission.Accuracy,
			"submitted_at": submission.SubmittedAt,
			"status":       "completed",
		}

		submissions = append(submissions, submissionResponse)
	}

	c.JSON(http.StatusOK, gin.H{
		"submissions": submissions,
		"limit":       limit,
		"offset":      offset,
		"count":       len(submissions),
	})
}

// getMatchedTastingNote returns the matched note from correct notes, or empty string if no match
func getMatchedTastingNote(userNote, correctNotes string) string {
	if userNote == "" || correctNotes == "" {
		return ""
	}
	
	// Clean and normalize user input
	userNote = strings.TrimSpace(strings.ToLower(userNote))
	
	// Split correct notes by comma and check each one
	correctNotesSlice := strings.Split(correctNotes, ",")
	for _, note := range correctNotesSlice {
		note = strings.TrimSpace(strings.ToLower(note))
		if note == "" {
			continue
		}
		
		// Check for exact match
		if userNote == note {
			return note
		}
		
		// Check for partial match (user note contains correct note or vice versa)
		if strings.Contains(userNote, note) || strings.Contains(note, userNote) {
			return note
		}
	}
	
	return ""
}

// matchesTastingNotes checks if a user's tasting note matches any of the comma-separated correct notes (legacy function)
func matchesTastingNotes(userNote, correctNotes string) bool {
	return getMatchedTastingNote(userNote, correctNotes) != ""
}
