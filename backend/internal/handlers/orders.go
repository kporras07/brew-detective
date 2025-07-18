package handlers

import (
	"context"
	"net/http"
	"time"

	"brew-detective-backend/internal/database"
	"brew-detective-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	order.Status = "pending"
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