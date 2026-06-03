package handlers

import (
	"net/http"
	"strconv"
	"time"

	"brew-detective-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"brew-detective-backend/internal/models"
)

func (h *Handler) CreateOrder(c *gin.Context) {
	var order models.Order

	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order data"})
		return
	}

	order.ID = uuid.New().String()
	order.OrderID = utils.GenerateOrderID()
	order.Status = "pending"
	order.IsSubmissionUsed = false
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	if err := h.Store.CreateOrder(c.Request.Context(), &order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":           "Order created successfully",
		"order_id":          order.ID,
		"customer_order_id": order.OrderID,
		"status":            order.Status,
	})
}

func (h *Handler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")

	order, err := h.Store.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"order": order})
}

func (h *Handler) UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")

	var updates struct {
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status data"})
		return
	}

	order, err := h.Store.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	order.Status = updates.Status
	order.UpdatedAt = time.Now()

	if err := h.Store.SetOrder(c.Request.Context(), order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order status updated successfully", "order": order})
}

func (h *Handler) GetAllOrders(c *gin.Context) {
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

	orders, err := h.Store.ListOrders(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders", "details": err.Error()})
		return
	}

	var orderResponses []map[string]interface{}
	for _, order := range orders {
		var userName string
		if order.UserID != "" {
			if user, err := h.Store.GetUser(c.Request.Context(), order.UserID); err == nil {
				userName = user.Name
			}
		}

		var caseName string
		if order.CaseID != "" {
			if coffeeCase, err := h.Store.GetCase(c.Request.Context(), order.CaseID); err == nil {
				caseName = coffeeCase.Name
			}
		}

		orderResponses = append(orderResponses, map[string]interface{}{
			"id":                 order.ID,
			"order_id":           order.OrderID,
			"user_id":            order.UserID,
			"user_name":          userName,
			"case_id":            order.CaseID,
			"case_name":          caseName,
			"contact_info":       order.ContactInfo,
			"status":             order.Status,
			"total_amount":       order.TotalAmount,
			"is_submission_used": order.IsSubmissionUsed,
			"submission_used_by": order.SubmissionUsedBy,
			"submission_used_at": order.SubmissionUsedAt,
			"created_at":         order.CreatedAt,
			"updated_at":         order.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"orders": orderResponses,
		"limit":  limit,
		"offset": offset,
		"count":  len(orderResponses),
	})
}
