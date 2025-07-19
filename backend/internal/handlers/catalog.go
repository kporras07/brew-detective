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

// GetCatalogByCategory returns catalog items for a specific category
func GetCatalogByCategory(c *gin.Context) {
	category := c.Param("category")
	
	// Validate category
	validCategories := map[string]bool{
		"region":        true,
		"variety":       true,
		"process":       true,
		"brewing_method": true,
	}
	
	if !validCategories[category] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category. Must be region, variety, process, or brewing_method"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	iter := database.FirestoreClient.Collection(database.CatalogCollection).
		Where("category", "==", category).
		Where("is_active", "==", true).
		Documents(ctx)

	var items []models.CatalogItem
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch catalog items"})
			return
		}

		var item models.CatalogItem
		if err := doc.DataTo(&item); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse catalog item"})
			return
		}

		items = append(items, item)
	}

	// Sort by display order
	sort.Slice(items, func(i, j int) bool {
		return items[i].DisplayOrder < items[j].DisplayOrder
	})

	c.JSON(http.StatusOK, gin.H{"items": items})
}

// GetAllCatalog returns all catalog items grouped by category
func GetAllCatalog(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	iter := database.FirestoreClient.Collection(database.CatalogCollection).
		Where("is_active", "==", true).
		Documents(ctx)

	catalogMap := make(map[string][]models.CatalogItem)
	
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch catalog items"})
			return
		}

		var item models.CatalogItem
		if err := doc.DataTo(&item); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse catalog item"})
			return
		}

		catalogMap[item.Category] = append(catalogMap[item.Category], item)
	}

	// Sort each category by display order
	for category := range catalogMap {
		sort.Slice(catalogMap[category], func(i, j int) bool {
			return catalogMap[category][i].DisplayOrder < catalogMap[category][j].DisplayOrder
		})
	}

	c.JSON(http.StatusOK, gin.H{"catalog": catalogMap})
}