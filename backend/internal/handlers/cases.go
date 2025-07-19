package handlers

import (
	"context"
	"net/http"
	"time"

	"brew-detective-backend/internal/database"
	"brew-detective-backend/internal/models"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

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