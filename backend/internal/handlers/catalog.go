package handlers

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"brew-detective-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) GetCatalogByCategory(c *gin.Context) {
	category := c.Param("category")

	validCategories := map[string]bool{
		"region":         true,
		"variety":        true,
		"process":        true,
		"brewing_method": true,
	}

	if !validCategories[category] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category. Must be region, variety, process, or brewing_method"})
		return
	}

	items, err := h.Store.ListCatalogByCategory(c.Request.Context(), category, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch catalog items", "details": err.Error()})
		return
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].DisplayOrder < items[j].DisplayOrder
	})

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) GetAllCatalog(c *gin.Context) {
	items, err := h.Store.ListActiveCatalog(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch catalog items", "details": err.Error()})
		return
	}

	catalogMap := make(map[string][]models.CatalogItem)
	for _, item := range items {
		catalogMap[item.Category] = append(catalogMap[item.Category], item)
	}

	for category := range catalogMap {
		sort.Slice(catalogMap[category], func(i, j int) bool {
			return catalogMap[category][i].DisplayOrder < catalogMap[category][j].DisplayOrder
		})
	}

	c.JSON(http.StatusOK, gin.H{"catalog": catalogMap})
}

func (h *Handler) CreateCatalogItem(c *gin.Context) {
	var item models.CatalogItem

	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid catalog item data"})
		return
	}

	if item.Value == "" || item.Label == "" || item.Category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Value, label, and category are required"})
		return
	}

	validCategories := map[string]bool{
		"region":         true,
		"variety":        true,
		"process":        true,
		"brewing_method": true,
	}

	if !validCategories[item.Category] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
		return
	}

	item.ID = uuid.New().String()
	item.CreatedAt = time.Now()
	if !item.IsActive {
		item.IsActive = true
	}

	if err := h.Store.CreateCatalogItem(c.Request.Context(), &item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create catalog item"})
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (h *Handler) UpdateCatalogItem(c *gin.Context) {
	itemID := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid update data"})
		return
	}

	// Only allow specific fields to be updated
	filtered := make(map[string]interface{})
	for key, value := range updates {
		switch key {
		case "label", "value", "is_active", "display_order":
			filtered[key] = value
		}
	}

	if len(filtered) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update"})
		return
	}

	if err := h.Store.UpdateCatalogItem(c.Request.Context(), itemID, filtered); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update catalog item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Catalog item updated successfully"})
}

func (h *Handler) DeleteCatalogItem(c *gin.Context) {
	itemID := c.Param("id")

	if err := h.Store.DeleteCatalogItem(c.Request.Context(), itemID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete catalog item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Catalog item deleted successfully"})
}

func (h *Handler) GetAllCatalogItems(c *gin.Context) {
	category := c.Query("category")
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit := 50
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

	items, err := h.Store.ListAllCatalogItems(c.Request.Context(), category, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch catalog items", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  items,
		"limit":  limit,
		"offset": offset,
		"count":  len(items),
	})
}
