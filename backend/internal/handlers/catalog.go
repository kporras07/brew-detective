package handlers

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"time"

	"brew-detective-backend/internal/database"
	"brew-detective-backend/internal/models"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

// CreateCatalogItem creates a new catalog item (admin only)
func CreateCatalogItem(c *gin.Context) {
	var item models.CatalogItem

	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid catalog item data"})
		return
	}

	// Validate required fields
	if item.Value == "" || item.Label == "" || item.Category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Value, label, and category are required"})
		return
	}

	// Validate category
	validCategories := map[string]bool{
		"region":        true,
		"variety":       true,
		"process":       true,
		"brewing_method": true,
	}
	
	if !validCategories[item.Category] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
		return
	}

	// Set defaults
	item.ID = uuid.New().String()
	item.CreatedAt = time.Now()
	if !item.IsActive {
		item.IsActive = true // Default to active
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Save to Firestore
	_, err := database.FirestoreClient.Collection(database.CatalogCollection).
		Doc(item.ID).Set(ctx, item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create catalog item"})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// UpdateCatalogItem updates an existing catalog item (admin only)
func UpdateCatalogItem(c *gin.Context) {
	itemID := c.Param("id")
	
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid update data"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build Firestore updates
	var firestoreUpdates []firestore.Update
	for key, value := range updates {
		// Only allow specific fields to be updated
		switch key {
		case "label", "value", "is_active", "display_order":
			firestoreUpdates = append(firestoreUpdates, firestore.Update{
				Path:  key,
				Value: value,
			})
		}
	}

	if len(firestoreUpdates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update"})
		return
	}

	// Update document
	_, err := database.FirestoreClient.Collection(database.CatalogCollection).
		Doc(itemID).Update(ctx, firestoreUpdates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update catalog item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Catalog item updated successfully"})
}

// DeleteCatalogItem deletes a catalog item (admin only)
func DeleteCatalogItem(c *gin.Context) {
	itemID := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Delete from Firestore
	_, err := database.FirestoreClient.Collection(database.CatalogCollection).
		Doc(itemID).Delete(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete catalog item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Catalog item deleted successfully"})
}

// GetAllCatalogItems returns all catalog items with pagination (admin only)
func GetAllCatalogItems(c *gin.Context) {
	// Get query parameters
	category := c.Query("category")
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit := 50 // Default limit
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

	// Build query
	query := database.FirestoreClient.Collection(database.CatalogCollection).
		OrderBy("category", firestore.Asc).
		OrderBy("display_order", firestore.Asc).
		Limit(limit).
		Offset(offset)

	// Add category filter if provided
	if category != "" {
		query = query.Where("category", "==", category)
	}

	iter := query.Documents(ctx)
	
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

	c.JSON(http.StatusOK, gin.H{
		"items":  items,
		"limit":  limit,
		"offset": offset,
		"count":  len(items),
	})
}