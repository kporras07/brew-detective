package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"brew-detective-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) CreateCase(c *gin.Context) {
	var newCase models.CoffeeCase

	if err := c.ShouldBindJSON(&newCase); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case data"})
		return
	}

	if newCase.Name == "" || newCase.Description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name and description are required"})
		return
	}

	newCase.ID = uuid.New().String()
	newCase.CreatedAt = time.Now()
	newCase.UpdatedAt = time.Now()

	for i := range newCase.Coffees {
		if newCase.Coffees[i].ID == "" || strings.HasPrefix(newCase.Coffees[i].ID, "coffee_") {
			newCase.Coffees[i].ID = uuid.New().String()
		}
	}

	if !newCase.IsActive {
		newCase.IsActive = false
	}

	if err := h.Store.CreateCase(c.Request.Context(), &newCase); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create case"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Case created successfully",
		"case":    newCase,
	})
}

func (h *Handler) UpdateCase(c *gin.Context) {
	caseID := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid update data"})
		return
	}

	// Handle coffee UUIDs if coffees are being updated
	if coffeesData, exists := updates["coffees"]; exists {
		if coffeesSlice, ok := coffeesData.([]interface{}); ok {
			for i, coffeeData := range coffeesSlice {
				if coffeeMap, ok := coffeeData.(map[string]interface{}); ok {
					if id, hasID := coffeeMap["id"]; hasID {
						if idStr, ok := id.(string); ok && (idStr == "" || strings.HasPrefix(idStr, "coffee_")) {
							coffeeMap["id"] = uuid.New().String()
							coffeesSlice[i] = coffeeMap
						}
					}
				}
			}
			updates["coffees"] = coffeesSlice
		}
	}

	updates["updated_at"] = time.Now()

	// Check if case exists
	_, err := h.Store.GetCase(c.Request.Context(), caseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	if err := h.Store.UpdateCase(c.Request.Context(), caseID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update case"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Case updated successfully"})
}

func (h *Handler) DeleteCase(c *gin.Context) {
	caseID := c.Param("id")

	// Check if case exists
	_, err := h.Store.GetCase(c.Request.Context(), caseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	if err := h.Store.DeleteCase(c.Request.Context(), caseID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete case"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Case deleted successfully"})
}

func (h *Handler) GetAllCases(c *gin.Context) {
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	cases, err := h.Store.ListCases(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cases", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cases":  cases,
		"limit":  limit,
		"offset": offset,
		"count":  len(cases),
	})
}

func (h *Handler) GetCases(c *gin.Context) {
	cases, err := h.Store.ListActiveCases(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cases"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cases": cases})
}

func (h *Handler) GetCaseByID(c *gin.Context) {
	caseID := c.Param("id")

	coffeeCase, err := h.Store.GetCase(c.Request.Context(), caseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"case": coffeeCase})
}

func (h *Handler) GetActiveCase(c *gin.Context) {
	coffeeCase, err := h.Store.GetActiveCase(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active case found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"case": coffeeCase})
}

func (h *Handler) GetActiveCasePublic(c *gin.Context) {
	coffeeCase, err := h.Store.GetActiveCase(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active case found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"case": toPublicCase(coffeeCase)})
}

func (h *Handler) GetCasesPublic(c *gin.Context) {
	cases, err := h.Store.ListActiveCases(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cases"})
		return
	}

	var publicCases []models.PublicCoffeeCase
	for _, coffeeCase := range cases {
		publicCases = append(publicCases, toPublicCase(&coffeeCase))
	}

	c.JSON(http.StatusOK, gin.H{"cases": publicCases})
}

func (h *Handler) GetCaseByIDPublic(c *gin.Context) {
	caseID := c.Param("id")

	coffeeCase, err := h.Store.GetCase(c.Request.Context(), caseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"case": toPublicCase(coffeeCase)})
}

func toPublicCase(coffeeCase *models.CoffeeCase) models.PublicCoffeeCase {
	coffeeIDs := make([]string, len(coffeeCase.Coffees))
	publicCoffees := make([]models.PublicCoffeeItem, len(coffeeCase.Coffees))
	for i, coffee := range coffeeCase.Coffees {
		coffeeIDs[i] = coffee.ID
		pc := models.PublicCoffeeItem{
			ID:               coffee.ID,
			EnabledQuestions: coffee.EnabledQuestions,
		}
		for _, aq := range coffee.AdditionalQuestions {
			pc.AdditionalQuestions = append(pc.AdditionalQuestions, models.PublicAdditionalQuestion{
				ID:       aq.ID,
				Question: aq.Question,
				Options:  aq.Options,
				Points:   aq.Points,
			})
		}
		publicCoffees[i] = pc
	}
	return models.PublicCoffeeCase{
		ID:               coffeeCase.ID,
		Name:             coffeeCase.Name,
		Description:      coffeeCase.Description,
		EnabledQuestions: coffeeCase.EnabledQuestions,
		CoffeeIDs:        coffeeIDs,
		Coffees:          publicCoffees,
		CoffeeCount:      len(coffeeCase.Coffees),
		IsActive:         coffeeCase.IsActive,
	}
}
