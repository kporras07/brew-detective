package handlers

import (
	"context"
	"fmt"
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
	fmt.Printf("🔍 [SCORING DEBUG] Starting score calculation for submission ID: %s\n", submission.ID)

	if activeCase == nil {
		fmt.Printf("❌ [SCORING DEBUG] No case provided, using default scoring\n")
		return calculateScoreDefault(submission)
	}

	fmt.Printf("✅ [SCORING DEBUG] Active case found: %s (ID: %s)\n", activeCase.Name, activeCase.ID)
	fmt.Printf("📋 [SCORING DEBUG] Enabled questions: Region=%t, Variety=%t, Process=%t, TasteNote1=%t, TasteNote2=%t, FavoriteCoffee=%t, BrewingMethod=%t\n",
		activeCase.EnabledQuestions.Region, activeCase.EnabledQuestions.Variety, activeCase.EnabledQuestions.Process,
		activeCase.EnabledQuestions.TasteNote1, activeCase.EnabledQuestions.TasteNote2,
		activeCase.EnabledQuestions.FavoriteCoffee, activeCase.EnabledQuestions.BrewingMethod)
	fmt.Printf("☕ [SCORING DEBUG] Case has %d coffees\n", len(activeCase.Coffees))

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
	fmt.Printf("📊 [SCORING DEBUG] Enabled questions per coffee: %d\n", enabledQuestionsPerCoffee)
	fmt.Printf("📊 [SCORING DEBUG] Total questions: %d coffees × %d questions = %d\n", len(submission.CoffeeAnswers), enabledQuestionsPerCoffee, totalQuestions)

	if totalQuestions == 0 {
		fmt.Printf("⚠️ [SCORING DEBUG] No questions enabled, returning 0 score\n")
		return 0, 0.0
	}

	correctAnswers := 0
	basePoints := 100

	fmt.Printf("👤 [SCORING DEBUG] User submitted %d coffee answers\n", len(submission.CoffeeAnswers))

	for i, answer := range submission.CoffeeAnswers {
		fmt.Printf("\n☕ [SCORING DEBUG] === Processing Coffee #%d (ID: %s) ===\n", i+1, answer.CoffeeID)

		// Find the correct coffee data for this answer
		var correctCoffee *models.CoffeeItem
		for _, coffee := range activeCase.Coffees {
			if coffee.ID == answer.CoffeeID {
				correctCoffee = &coffee
				break
			}
		}

		if correctCoffee == nil {
			fmt.Printf("❌ [SCORING DEBUG] Coffee ID %s not found in case, skipping\n", answer.CoffeeID)
			continue
		}

		fmt.Printf("✅ [SCORING DEBUG] Found correct coffee: %s\n", correctCoffee.Name)
		fmt.Printf("📋 [SCORING DEBUG] Correct data - Region: '%s', Variety: '%s', Process: '%s', TastingNotes: '%s'\n",
			correctCoffee.Region, correctCoffee.Variety, correctCoffee.Process, correctCoffee.TastingNotes)
		fmt.Printf("👤 [SCORING DEBUG] User answers - Region: '%s', Variety: '%s', Process: '%s', Note1: '%s', Note2: '%s'\n",
			answer.Region, answer.Variety, answer.Process, answer.TasteNote1, answer.TasteNote2)

		coffeeCorrectAnswers := 0

		if activeCase.EnabledQuestions.Region && answer.Region != "" {
			userRegion := strings.TrimSpace(answer.Region)
			correctRegion := strings.TrimSpace(correctCoffee.Region)
			isCorrect := strings.EqualFold(userRegion, correctRegion)
			fmt.Printf("🌍 [SCORING DEBUG] Region check: '%s' vs '%s' = %t\n", userRegion, correctRegion, isCorrect)
			if isCorrect {
				correctAnswers++
				coffeeCorrectAnswers++
			}
		} else if activeCase.EnabledQuestions.Region {
			fmt.Printf("🌍 [SCORING DEBUG] Region question enabled but user answer is empty\n")
		}

		if activeCase.EnabledQuestions.Variety && answer.Variety != "" {
			userVariety := strings.TrimSpace(answer.Variety)
			correctVariety := strings.TrimSpace(correctCoffee.Variety)
			isCorrect := strings.EqualFold(userVariety, correctVariety)
			fmt.Printf("🌱 [SCORING DEBUG] Variety check: '%s' vs '%s' = %t\n", userVariety, correctVariety, isCorrect)
			if isCorrect {
				correctAnswers++
				coffeeCorrectAnswers++
			}
		} else if activeCase.EnabledQuestions.Variety {
			fmt.Printf("🌱 [SCORING DEBUG] Variety question enabled but user answer is empty\n")
		}

		if activeCase.EnabledQuestions.Process && answer.Process != "" {
			userProcess := strings.TrimSpace(answer.Process)
			correctProcess := strings.TrimSpace(correctCoffee.Process)
			isCorrect := strings.EqualFold(userProcess, correctProcess)
			fmt.Printf("⚙️ [SCORING DEBUG] Process check: '%s' vs '%s' = %t\n", userProcess, correctProcess, isCorrect)
			if isCorrect {
				correctAnswers++
				coffeeCorrectAnswers++
			}
		} else if activeCase.EnabledQuestions.Process {
			fmt.Printf("⚙️ [SCORING DEBUG] Process question enabled but user answer is empty\n")
		}

		var awardedTastingNotes []string
		fmt.Printf("🍫 [SCORING DEBUG] Starting tasting notes evaluation...\n")

		if activeCase.EnabledQuestions.TasteNote1 && answer.TasteNote1 != "" {
			fmt.Printf("🍫 [SCORING DEBUG] TasteNote1 check: user='%s' vs correct='%s'\n", answer.TasteNote1, correctCoffee.TastingNotes)
			if matchedNote := getMatchedTastingNote(answer.TasteNote1, correctCoffee.TastingNotes); matchedNote != "" {
				fmt.Printf("✅ [SCORING DEBUG] TasteNote1 MATCHED: '%s'\n", matchedNote)
				awardedTastingNotes = append(awardedTastingNotes, matchedNote)
				correctAnswers++
				coffeeCorrectAnswers++
			} else {
				fmt.Printf("❌ [SCORING DEBUG] TasteNote1 NO MATCH\n")
			}
		} else if activeCase.EnabledQuestions.TasteNote1 {
			fmt.Printf("🍫 [SCORING DEBUG] TasteNote1 question enabled but user answer is empty\n")
		}

		if activeCase.EnabledQuestions.TasteNote2 && answer.TasteNote2 != "" {
			fmt.Printf("🍫 [SCORING DEBUG] TasteNote2 check: user='%s' vs correct='%s'\n", answer.TasteNote2, correctCoffee.TastingNotes)
			if matchedNote := getMatchedTastingNote(answer.TasteNote2, correctCoffee.TastingNotes); matchedNote != "" {
				alreadyAwarded := false
				for _, awarded := range awardedTastingNotes {
					if strings.EqualFold(awarded, matchedNote) {
						alreadyAwarded = true
						break
					}
				}
				if !alreadyAwarded {
					fmt.Printf("✅ [SCORING DEBUG] TasteNote2 MATCHED (new): '%s'\n", matchedNote)
					correctAnswers++
					coffeeCorrectAnswers++
				} else {
					fmt.Printf("⚠️ [SCORING DEBUG] TasteNote2 matched '%s' but already awarded for this note\n", matchedNote)
				}
			} else {
				fmt.Printf("❌ [SCORING DEBUG] TasteNote2 NO MATCH\n")
			}
		} else if activeCase.EnabledQuestions.TasteNote2 {
			fmt.Printf("🍫 [SCORING DEBUG] TasteNote2 question enabled but user answer is empty\n")
		}

		fmt.Printf("📊 [SCORING DEBUG] Coffee #%d results: %d/%d correct answers\n", i+1, coffeeCorrectAnswers, enabledQuestionsPerCoffee)
	}

	bonusPoints := 0
	fmt.Printf("\n🎁 [SCORING DEBUG] === Bonus Questions Evaluation ===\n")

	if activeCase.EnabledQuestions.FavoriteCoffee && submission.FavoriteCoffee != "" {
		bonusPoints += 50
		fmt.Printf("✅ [SCORING DEBUG] FavoriteCoffee bonus: +50 points (answer: '%s')\n", submission.FavoriteCoffee)
	} else if activeCase.EnabledQuestions.FavoriteCoffee {
		fmt.Printf("❌ [SCORING DEBUG] FavoriteCoffee enabled but no answer provided\n")
	} else {
		fmt.Printf("⚪ [SCORING DEBUG] FavoriteCoffee question disabled\n")
	}

	if activeCase.EnabledQuestions.BrewingMethod && submission.BrewingMethod != "" {
		bonusPoints += 50
		fmt.Printf("✅ [SCORING DEBUG] BrewingMethod bonus: +50 points (answer: '%s')\n", submission.BrewingMethod)
	} else if activeCase.EnabledQuestions.BrewingMethod {
		fmt.Printf("❌ [SCORING DEBUG] BrewingMethod enabled but no answer provided\n")
	} else {
		fmt.Printf("⚪ [SCORING DEBUG] BrewingMethod question disabled\n")
	}

	accuracy := float64(correctAnswers) / float64(totalQuestions)
	score := int(float64(basePoints)*accuracy*float64(len(submission.CoffeeAnswers))) + bonusPoints

	fmt.Printf("\n🏆 [SCORING DEBUG] === FINAL CALCULATION ===\n")
	fmt.Printf("📊 [SCORING DEBUG] Correct answers: %d/%d\n", correctAnswers, totalQuestions)
	fmt.Printf("📊 [SCORING DEBUG] Accuracy: %.4f (%.1f%%)\n", accuracy, accuracy*100)
	fmt.Printf("📊 [SCORING DEBUG] Base calculation: %d points × %.4f accuracy × %d coffees = %.2f\n",
		basePoints, accuracy, len(submission.CoffeeAnswers), float64(basePoints)*accuracy*float64(len(submission.CoffeeAnswers)))
	fmt.Printf("📊 [SCORING DEBUG] Bonus points: %d\n", bonusPoints)
	fmt.Printf("🏆 [SCORING DEBUG] FINAL SCORE: %d points\n", score)

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

// getMatchedTastingNote returns the matched note from correct notes, or empty string if no match
func getMatchedTastingNote(userNote, correctNotes string) string {
	fmt.Printf("🔍 [TASTING DEBUG] Checking note: user='%s' vs correct='%s'\n", userNote, correctNotes)

	if userNote == "" || correctNotes == "" {
		fmt.Printf("⚠️ [TASTING DEBUG] Empty input - userNote='%s', correctNotes='%s'\n", userNote, correctNotes)
		return ""
	}

	userNote = strings.TrimSpace(strings.ToLower(userNote))
	fmt.Printf("🧹 [TASTING DEBUG] Normalized user note: '%s'\n", userNote)

	correctNotesSlice := strings.Split(correctNotes, ",")
	fmt.Printf("📝 [TASTING DEBUG] Split correct notes into %d parts: %v\n", len(correctNotesSlice), correctNotesSlice)

	for i, note := range correctNotesSlice {
		originalNote := note
		note = strings.TrimSpace(strings.ToLower(note))
		fmt.Printf("🔍 [TASTING DEBUG] Checking note #%d: original='%s', normalized='%s'\n", i+1, originalNote, note)

		if note == "" {
			fmt.Printf("⚠️ [TASTING DEBUG] Note #%d is empty after normalization, skipping\n", i+1)
			continue
		}

		if userNote == note {
			fmt.Printf("✅ [TASTING DEBUG] EXACT MATCH found: '%s' == '%s'\n", userNote, note)
			return note
		}

		if strings.Contains(userNote, note) {
			fmt.Printf("✅ [TASTING DEBUG] PARTIAL MATCH found: user '%s' contains correct '%s'\n", userNote, note)
			return note
		}
		if strings.Contains(note, userNote) {
			fmt.Printf("✅ [TASTING DEBUG] PARTIAL MATCH found: correct '%s' contains user '%s'\n", note, userNote)
			return note
		}

		fmt.Printf("❌ [TASTING DEBUG] No match for note #%d\n", i+1)
	}

	fmt.Printf("❌ [TASTING DEBUG] No matches found for user note '%s'\n", userNote)
	return ""
}

// matchesTastingNotes checks if a user's tasting note matches any of the comma-separated correct notes
func matchesTastingNotes(userNote, correctNotes string) bool {
	return getMatchedTastingNote(userNote, correctNotes) != ""
}
