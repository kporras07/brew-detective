package store

import (
	"context"

	"brew-detective-backend/internal/models"
)

// Store defines the interface for all database operations.
type Store interface {
	// Users
	GetUser(ctx context.Context, id string) (*models.User, error)
	SetUser(ctx context.Context, user *models.User) error
	ListUsers(ctx context.Context) ([]models.User, error)

	// Cases
	GetCase(ctx context.Context, id string) (*models.CoffeeCase, error)
	CreateCase(ctx context.Context, coffeeCase *models.CoffeeCase) error
	UpdateCase(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteCase(ctx context.Context, id string) error
	ListCases(ctx context.Context, limit, offset int) ([]models.CoffeeCase, error)
	ListActiveCases(ctx context.Context) ([]models.CoffeeCase, error)
	GetActiveCase(ctx context.Context) (*models.CoffeeCase, error)

	// Orders
	GetOrder(ctx context.Context, id string) (*models.Order, error)
	CreateOrder(ctx context.Context, order *models.Order) error
	SetOrder(ctx context.Context, order *models.Order) error
	ListOrders(ctx context.Context, limit, offset int) ([]models.Order, error)
	GetOrderByOrderID(ctx context.Context, orderID string) (*models.Order, error)
	UpdateOrderFields(ctx context.Context, orderID string, updates map[string]interface{}) error

	// Submissions
	CreateSubmission(ctx context.Context, submission *models.Submission) error
	ListSubmissionsByUser(ctx context.Context, userID string, limit, offset int) ([]models.Submission, error)
	ListSubmissionsByCase(ctx context.Context, caseID string) ([]models.Submission, error)

	// Catalog
	ListCatalogByCategory(ctx context.Context, category string, activeOnly bool) ([]models.CatalogItem, error)
	ListActiveCatalog(ctx context.Context) ([]models.CatalogItem, error)
	CreateCatalogItem(ctx context.Context, item *models.CatalogItem) error
	UpdateCatalogItem(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteCatalogItem(ctx context.Context, id string) error
	ListAllCatalogItems(ctx context.Context, category string, limit, offset int) ([]models.CatalogItem, error)
}
