package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"brew-detective-backend/internal/database"
	"brew-detective-backend/internal/models"
	"brew-detective-backend/internal/utils"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

// CreateOrder creates a new coffee case order
func CreateOrder(c *gin.Context) {
	var order models.Order
	
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order data"})
		return
	}

	// Generate order ID and set timestamps
	order.ID = uuid.New().String()
	order.OrderID = utils.GenerateOrderID() // Generate 6-character order ID
	order.Status = "pending"
	order.IsSubmissionUsed = false
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Save order to Firestore
	_, err := database.FirestoreClient.Collection(database.OrdersCollection).
		Doc(order.ID).Set(ctx, order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Order created successfully",
		"order_id": order.ID,
		"customer_order_id": order.OrderID, // This is the 6-character ID for customers
		"status": order.Status,
	})
}

// GetOrder returns a specific order
func GetOrder(c *gin.Context) {
	orderID := c.Param("id")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	doc, err := database.FirestoreClient.Collection(database.OrdersCollection).Doc(orderID).Get(ctx)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	var order models.Order
	if err := doc.DataTo(&order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse order data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"order": order})
}

// UpdateOrderStatus updates the status of an order
func UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")
	
	var updates struct {
		Status string `json:"status"`
	}
	
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status data"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	orderRef := database.FirestoreClient.Collection(database.OrdersCollection).Doc(orderID)
	
	// Check if order exists
	doc, err := orderRef.Get(ctx)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	var order models.Order
	if err := doc.DataTo(&order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse order data"})
		return
	}

	// Update status
	order.Status = updates.Status
	order.UpdatedAt = time.Now()

	// Save updated order
	if _, err := orderRef.Set(ctx, order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order status updated successfully", "order": order})
}

// GetAllOrders returns all orders with pagination (admin only)
func GetAllOrders(c *gin.Context) {
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

	// Query all orders ordered by creation date (newest first)
	query := database.FirestoreClient.Collection(database.OrdersCollection).
		OrderBy("created_at", firestore.Desc).
		Limit(limit).
		Offset(offset)

	iter := query.Documents(ctx)

	var orders []map[string]interface{}
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders", "details": err.Error()})
			return
		}

		var order models.Order
		if err := doc.DataTo(&order); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse order data"})
			return
		}

		// Get user name for the order
		var userName string
		if order.UserID != "" {
			userDoc, err := database.FirestoreClient.Collection(database.UsersCollection).Doc(order.UserID).Get(ctx)
			if err == nil && userDoc.Exists() {
				var user models.User
				if err := userDoc.DataTo(&user); err == nil {
					userName = user.Name
				}
			}
		}

		// Get case name for the order
		var caseName string
		if order.CaseID != "" {
			caseDoc, err := database.FirestoreClient.Collection(database.CasesCollection).Doc(order.CaseID).Get(ctx)
			if err == nil && caseDoc.Exists() {
				var coffeeCase models.CoffeeCase
				if err := caseDoc.DataTo(&coffeeCase); err == nil {
					caseName = coffeeCase.Name
				}
			}
		}

		// Create response object with order and related info
		orderResponse := map[string]interface{}{
			"id":                  order.ID,
			"order_id":            order.OrderID,
			"user_id":             order.UserID,
			"user_name":           userName,
			"case_id":             order.CaseID,
			"case_name":           caseName,
			"contact_info":        order.ContactInfo,
			"status":              order.Status,
			"total_amount":        order.TotalAmount,
			"is_submission_used":  order.IsSubmissionUsed,
			"submission_used_by":  order.SubmissionUsedBy,
			"submission_used_at":  order.SubmissionUsedAt,
			"created_at":          order.CreatedAt,
			"updated_at":          order.UpdatedAt,
		}

		orders = append(orders, orderResponse)
	}

	c.JSON(http.StatusOK, gin.H{
		"orders": orders,
		"limit":  limit,
		"offset": offset,
		"count":  len(orders),
	})
}