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

// Admin case management functions

// CreateCase creates a new coffee case (admin only)
func CreateCase(c *gin.Context) {
	var newCase models.CoffeeCase

	if err := c.ShouldBindJSON(&newCase); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case data"})
		return
	}

	// Validate required fields
	if newCase.Name == "" || newCase.Description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name and description are required"})
		return
	}

	// Generate case ID and set timestamps
	newCase.ID = uuid.New().String()
	newCase.CreatedAt = time.Now()
	newCase.UpdatedAt = time.Now()

	// Generate UUIDs for each coffee if they don't have proper ones
	for i := range newCase.Coffees {
		if newCase.Coffees[i].ID == "" || strings.HasPrefix(newCase.Coffees[i].ID, "coffee_") {
			newCase.Coffees[i].ID = uuid.New().String()
		}
	}

	// Default to inactive when created
	if !newCase.IsActive {
		newCase.IsActive = false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Save case to Firestore
	_, err := database.FirestoreClient.Collection(database.CasesCollection).
		Doc(newCase.ID).Set(ctx, newCase)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create case"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Case created successfully",
		"case":    newCase,
	})
}

// UpdateCase updates an existing coffee case (admin only)
func UpdateCase(c *gin.Context) {
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

	// Add updated timestamp
	updates["updated_at"] = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if case exists
	doc, err := database.FirestoreClient.Collection(database.CasesCollection).Doc(caseID).Get(ctx)
	if err != nil || !doc.Exists() {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	// Convert map to firestore updates
	var firestoreUpdates []firestore.Update
	for key, value := range updates {
		firestoreUpdates = append(firestoreUpdates, firestore.Update{
			Path:  key,
			Value: value,
		})
	}

	// Update the case
	_, err = database.FirestoreClient.Collection(database.CasesCollection).
		Doc(caseID).Update(ctx, firestoreUpdates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update case"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Case updated successfully"})
}

// DeleteCase deletes a coffee case (admin only)
func DeleteCase(c *gin.Context) {
	caseID := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if case exists
	doc, err := database.FirestoreClient.Collection(database.CasesCollection).Doc(caseID).Get(ctx)
	if err != nil || !doc.Exists() {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	// Delete the case
	_, err = database.FirestoreClient.Collection(database.CasesCollection).Doc(caseID).Delete(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete case"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Case deleted successfully"})
}

// GetAllCases returns all coffee cases with pagination (admin only)
func GetAllCases(c *gin.Context) {
	// Get query parameters for pagination
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit := 20 // Default limit
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Query all cases ordered by creation date (newest first)
	query := database.FirestoreClient.Collection(database.CasesCollection).
		OrderBy("created_at", firestore.Desc).
		Limit(limit).
		Offset(offset)

	iter := query.Documents(ctx)

	var cases []models.CoffeeCase
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cases", "details": err.Error()})
			return
		}

		var coffeeCase models.CoffeeCase
		if err := doc.DataTo(&coffeeCase); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse case data"})
			return
		}

		cases = append(cases, coffeeCase)
	}

	c.JSON(http.StatusOK, gin.H{
		"cases":  cases,
		"limit":  limit,
		"offset": offset,
		"count":  len(cases),
	})
}

// GetCases returns all active coffee cases
func GetCases(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	iter := database.FirestoreClient.Collection(database.CasesCollection).
		Where("is_active", "==", true).
		Documents(ctx)

	var cases []models.CoffeeCase
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cases"})
			return
		}

		var coffeeCase models.CoffeeCase
		if err := doc.DataTo(&coffeeCase); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse case data"})
			return
		}

		cases = append(cases, coffeeCase)
	}

	c.JSON(http.StatusOK, gin.H{"cases": cases})
}

// GetCaseByID returns a specific coffee case
func GetCaseByID(c *gin.Context) {
	caseID := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	doc, err := database.FirestoreClient.Collection(database.CasesCollection).Doc(caseID).Get(ctx)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	var coffeeCase models.CoffeeCase
	if err := doc.DataTo(&coffeeCase); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse case data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"case": coffeeCase})
}

// GetActiveCase returns the current active coffee case
func GetActiveCase(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	iter := database.FirestoreClient.Collection(database.CasesCollection).
		Where("is_active", "==", true).
		Limit(1).
		Documents(ctx)

	doc, err := iter.Next()
	if err == iterator.Done {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active case found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch active case"})
		return
	}

	var coffeeCase models.CoffeeCase
	if err := doc.DataTo(&coffeeCase); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse case data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"case": coffeeCase})
}
