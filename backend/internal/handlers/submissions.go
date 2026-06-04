package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"brew-detective-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SubmitCase handles case submission
func (h *Handler) SubmitCase(c *gin.Context) {
	var submission models.Submission

	if err := c.ShouldBindJSON(&submission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid submission data"})
		return
	}

	if submission.OrderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}

	// Get the current active case
	activeCase, err := h.Store.GetActiveCase(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No active case available", "details": err.Error()})
		return
	}
	submission.CaseID = activeCase.ID

	// Validate order ID with detailed error messages
	orderValidation := h.validateOrderIDDetailed(c.Request.Context(), submission.OrderID)
	if !orderValidation.IsValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": orderValidation.ErrorMessage})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User authentication required"})
		return
	}
	submission.UserID = userID.(string)

	submission.ID = uuid.New().String()
	submission.SubmittedAt = time.Now()

	// Calculate score and accuracy
	score, accuracy := calculateScoreWithCase(&submission, activeCase)
	submission.Score = score
	submission.Accuracy = accuracy

	if err := h.Store.CreateSubmission(c.Request.Context(), &submission); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save submission"})
		return
	}

	// Mark order ID as used
	h.markOrderIDAsUsed(c.Request.Context(), submission.OrderID, submission.UserID)

	// Update user stats
	go h.updateUserStats(submission.UserID, score, accuracy)

	c.JSON(http.StatusCreated, gin.H{
		"message":       "Submission successful",
		"submission_id": submission.ID,
		"score":         score,
		"accuracy":      accuracy,
	})
}

// calculateScoreWithCase calculates the score and accuracy for a submission against a given case
func calculateScoreWithCase(submission *models.Submission, activeCase *models.CoffeeCase) (int, float64) {
	if activeCase == nil {
		return calculateScoreDefault(submission)
	}

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
		var correctCoffee *models.CoffeeItem
		for _, coffee := range activeCase.Coffees {
			if coffee.ID == answer.CoffeeID {
				correctCoffee = &coffee
				break
			}
		}

		if correctCoffee == nil {
			continue
		}

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

		var awardedTastingNotes []string

		if activeCase.EnabledQuestions.TasteNote1 && answer.TasteNote1 != "" {
			if matchedNote := getMatchedTastingNote(answer.TasteNote1, correctCoffee.TastingNotes); matchedNote != "" {
				awardedTastingNotes = append(awardedTastingNotes, matchedNote)
				correctAnswers++
			}
		}

		if activeCase.EnabledQuestions.TasteNote2 && answer.TasteNote2 != "" {
			if matchedNote := getMatchedTastingNote(answer.TasteNote2, correctCoffee.TastingNotes); matchedNote != "" {
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

	bonusPoints := 0

	if activeCase.EnabledQuestions.FavoriteCoffee && submission.FavoriteCoffee != "" {
		bonusPoints += 50
	}

	if activeCase.EnabledQuestions.BrewingMethod && submission.BrewingMethod != "" {
		bonusPoints += 50
	}

	accuracy := float64(correctAnswers) / float64(totalQuestions)
	score := int(float64(basePoints)*accuracy*float64(len(submission.CoffeeAnswers))) + bonusPoints

	return score, accuracy
}

// calculateScoreDefault provides fallback scoring when case info is not available
func calculateScoreDefault(submission *models.Submission) (int, float64) {
	totalQuestions := len(submission.CoffeeAnswers) * 3
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
func (h *Handler) updateUserStats(userID string, score int, accuracy float64) {
	if userID == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := h.Store.GetUser(ctx, userID)
	if err != nil {
		return
	}

	user.Points += score
	user.CasesCount++
	user.Accuracy = (user.Accuracy*float64(user.CasesCount-1) + accuracy) / float64(user.CasesCount)
	user.UpdatedAt = time.Now()

	h.Store.SetUser(ctx, user)
}

// updateBadges updates user badges based on achievements
func updateBadges(user *models.User) {
	badges := make(map[string]bool)

	for _, badge := range user.Badges {
		badges[badge] = true
	}

	if user.CasesCount >= 1 && !badges["🔍 Primer Caso"] {
		badges["🔍 Primer Caso"] = true
	}

	if user.Accuracy >= 0.7 && !badges["🎯 Precisión 70%"] {
		badges["🎯 Precisión 70%"] = true
	}
	if user.Accuracy >= 0.8 && !badges["💎 Catador Nivel 2"] {
		badges["💎 Catador Nivel 2"] = true
	}

	if user.Points >= 2000 && !badges["🏆 Detective Maestro"] {
		badges["🏆 Detective Maestro"] = true
	}

	if user.CasesCount >= 5 && !badges["🔥 Experto en Tuestes"] {
		badges["🔥 Experto en Tuestes"] = true
	}

	user.Badges = make([]string, 0, len(badges))
	for badge := range badges {
		user.Badges = append(user.Badges, badge)
	}
}

// OrderValidationResult holds the result of order ID validation
type OrderValidationResult struct {
	IsValid      bool
	ErrorMessage string
}

func (h *Handler) validateOrderIDDetailed(ctx context.Context, orderID string) OrderValidationResult {
	order, err := h.Store.GetOrderByOrderID(ctx, orderID)
	if err != nil {
		return OrderValidationResult{
			IsValid:      false,
			ErrorMessage: "Código de pedido no válido. Verifica que hayas ingresado el código correctamente.",
		}
	}

	if order.IsSubmissionUsed {
		return OrderValidationResult{
			IsValid:      false,
			ErrorMessage: "Este código de pedido ya fue utilizado para enviar respuestas. Cada código solo puede usarse una vez.",
		}
	}

	if order.Status != "delivered" {
		return OrderValidationResult{
			IsValid:      false,
			ErrorMessage: "Tu pedido aún no ha sido entregado. Solo puedes enviar respuestas después de recibir tu café.",
		}
	}

	return OrderValidationResult{
		IsValid:      true,
		ErrorMessage: "",
	}
}

func (h *Handler) markOrderIDAsUsed(ctx context.Context, orderID string, userID string) {
	order, err := h.Store.GetOrderByOrderID(ctx, orderID)
	if err != nil {
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"is_submission_used": true,
		"submission_used_by": userID,
		"submission_used_at": now,
		"updated_at":         now,
	}

	h.Store.UpdateOrderFields(ctx, order.ID, updates)
}

// GetUserSubmissions returns submissions for a specific user
func (h *Handler) GetUserSubmissions(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User authentication required"})
		return
	}

	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit := 10
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

	submissions, err := h.Store.ListSubmissionsByUser(c.Request.Context(), userID.(string), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch submissions", "details": err.Error()})
		return
	}

	var submissionResponses []map[string]interface{}
	for _, submission := range submissions {
		caseName := "Caso Desconocido"
		if coffeeCase, err := h.Store.GetCase(c.Request.Context(), submission.CaseID); err == nil {
			caseName = coffeeCase.Name
		}

		submissionResponses = append(submissionResponses, map[string]interface{}{
			"id":           submission.ID,
			"case_id":      submission.CaseID,
			"case_name":    caseName,
			"score":        submission.Score,
			"accuracy":     submission.Accuracy,
			"submitted_at": submission.SubmittedAt,
			"status":       "completed",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"submissions": submissionResponses,
		"limit":       limit,
		"offset":      offset,
		"count":       len(submissionResponses),
	})
}

func getMatchedTastingNote(userNote, correctNotes string) string {
	if userNote == "" || correctNotes == "" {
		return ""
	}

	userNote = strings.TrimSpace(strings.ToLower(userNote))
	correctNotesSlice := strings.Split(correctNotes, ",")

	for _, note := range correctNotesSlice {
		note = strings.TrimSpace(strings.ToLower(note))

		if note == "" {
			continue
		}

		if userNote == note || strings.Contains(userNote, note) || strings.Contains(note, userNote) {
			return note
		}
	}

	return ""
}

// matchesTastingNotes checks if a user's tasting note matches any of the comma-separated correct notes
func matchesTastingNotes(userNote, correctNotes string) bool {
	return getMatchedTastingNote(userNote, correctNotes) != ""
}
