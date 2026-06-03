package store

import (
	"context"
	"fmt"

	"brew-detective-backend/internal/models"
)

// MockStore is an in-memory implementation of Store for testing.
type MockStore struct {
	Users       map[string]*models.User
	Cases       map[string]*models.CoffeeCase
	Orders      map[string]*models.Order
	Submissions map[string]*models.Submission
	CatalogItems map[string]*models.CatalogItem

	// Error injection for testing error paths.
	// Global error — affects all operations.
	Err error

	// Per-method error overrides (take precedence over Err when non-nil).
	// Useful when a handler calls multiple store methods sequentially and
	// only the second one should fail.
	DeleteCaseErr            error
	UpdateCaseErr            error
	SetUserErr               error
	UpdateOrderFieldsErr     error
	ListSubmissionsByCaseErr error
	CreateCaseErr            error
	CreateCatalogItemErr     error
	DeleteCatalogItemErr     error
	UpdateCatalogItemErr     error
	CreateOrderErr           error
	CreateSubmissionErr       error
	SetOrderErr              error
}

// NewMockStore creates a new MockStore with empty maps.
func NewMockStore() *MockStore {
	return &MockStore{
		Users:        make(map[string]*models.User),
		Cases:        make(map[string]*models.CoffeeCase),
		Orders:       make(map[string]*models.Order),
		Submissions:  make(map[string]*models.Submission),
		CatalogItems: make(map[string]*models.CatalogItem),
	}
}

// --- Users ---

func (m *MockStore) GetUser(_ context.Context, id string) (*models.User, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	user, ok := m.Users[id]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (m *MockStore) SetUser(_ context.Context, user *models.User) error {
	if m.SetUserErr != nil {
		return m.SetUserErr
	}
	if m.Err != nil {
		return m.Err
	}
	m.Users[user.ID] = user
	return nil
}

func (m *MockStore) ListUsers(_ context.Context) ([]models.User, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	var users []models.User
	for _, u := range m.Users {
		users = append(users, *u)
	}
	return users, nil
}

// --- Cases ---

func (m *MockStore) GetCase(_ context.Context, id string) (*models.CoffeeCase, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	c, ok := m.Cases[id]
	if !ok {
		return nil, fmt.Errorf("case not found")
	}
	return c, nil
}

func (m *MockStore) CreateCase(_ context.Context, c *models.CoffeeCase) error {
	if m.CreateCaseErr != nil {
		return m.CreateCaseErr
	}
	if m.Err != nil {
		return m.Err
	}
	m.Cases[c.ID] = c
	return nil
}

func (m *MockStore) UpdateCase(_ context.Context, id string, updates map[string]interface{}) error {
	if m.UpdateCaseErr != nil {
		return m.UpdateCaseErr
	}
	if m.Err != nil {
		return m.Err
	}
	if _, ok := m.Cases[id]; !ok {
		return fmt.Errorf("case not found")
	}
	return nil
}

func (m *MockStore) DeleteCase(_ context.Context, id string) error {
	if m.DeleteCaseErr != nil {
		return m.DeleteCaseErr
	}
	if m.Err != nil {
		return m.Err
	}
	delete(m.Cases, id)
	return nil
}

func (m *MockStore) ListCases(_ context.Context, limit, offset int) ([]models.CoffeeCase, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	var cases []models.CoffeeCase
	for _, c := range m.Cases {
		cases = append(cases, *c)
	}
	// Simple pagination
	if offset >= len(cases) {
		return nil, nil
	}
	end := offset + limit
	if end > len(cases) {
		end = len(cases)
	}
	return cases[offset:end], nil
}

func (m *MockStore) ListActiveCases(_ context.Context) ([]models.CoffeeCase, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	var cases []models.CoffeeCase
	for _, c := range m.Cases {
		if c.IsActive {
			cases = append(cases, *c)
		}
	}
	return cases, nil
}

func (m *MockStore) GetActiveCase(_ context.Context) (*models.CoffeeCase, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	for _, c := range m.Cases {
		if c.IsActive {
			return c, nil
		}
	}
	return nil, fmt.Errorf("no active case found")
}

// --- Orders ---

func (m *MockStore) GetOrder(_ context.Context, id string) (*models.Order, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	o, ok := m.Orders[id]
	if !ok {
		return nil, fmt.Errorf("order not found")
	}
	return o, nil
}

func (m *MockStore) CreateOrder(_ context.Context, order *models.Order) error {
	if m.CreateOrderErr != nil {
		return m.CreateOrderErr
	}
	if m.Err != nil {
		return m.Err
	}
	m.Orders[order.ID] = order
	return nil
}

func (m *MockStore) SetOrder(_ context.Context, order *models.Order) error {
	if m.SetOrderErr != nil {
		return m.SetOrderErr
	}
	if m.Err != nil {
		return m.Err
	}
	m.Orders[order.ID] = order
	return nil
}

func (m *MockStore) ListOrders(_ context.Context, limit, offset int) ([]models.Order, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	var orders []models.Order
	for _, o := range m.Orders {
		orders = append(orders, *o)
	}
	if offset >= len(orders) {
		return nil, nil
	}
	end := offset + limit
	if end > len(orders) {
		end = len(orders)
	}
	return orders[offset:end], nil
}

func (m *MockStore) GetOrderByOrderID(_ context.Context, orderID string) (*models.Order, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	for _, o := range m.Orders {
		if o.OrderID == orderID {
			return o, nil
		}
	}
	return nil, fmt.Errorf("order not found")
}

func (m *MockStore) UpdateOrderFields(_ context.Context, docID string, updates map[string]interface{}) error {
	if m.UpdateOrderFieldsErr != nil {
		return m.UpdateOrderFieldsErr
	}
	if m.Err != nil {
		return m.Err
	}
	return nil
}

// --- Submissions ---

func (m *MockStore) CreateSubmission(_ context.Context, submission *models.Submission) error {
	if m.CreateSubmissionErr != nil {
		return m.CreateSubmissionErr
	}
	if m.Err != nil {
		return m.Err
	}
	m.Submissions[submission.ID] = submission
	return nil
}

func (m *MockStore) ListSubmissionsByUser(_ context.Context, userID string, limit, offset int) ([]models.Submission, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	var subs []models.Submission
	for _, s := range m.Submissions {
		if s.UserID == userID {
			subs = append(subs, *s)
		}
	}
	if offset >= len(subs) {
		return nil, nil
	}
	end := offset + limit
	if end > len(subs) {
		end = len(subs)
	}
	return subs[offset:end], nil
}

func (m *MockStore) ListSubmissionsByCase(_ context.Context, caseID string) ([]models.Submission, error) {
	if m.ListSubmissionsByCaseErr != nil {
		return nil, m.ListSubmissionsByCaseErr
	}
	if m.Err != nil {
		return nil, m.Err
	}
	var subs []models.Submission
	for _, s := range m.Submissions {
		if s.CaseID == caseID {
			subs = append(subs, *s)
		}
	}
	return subs, nil
}

// --- Catalog ---

func (m *MockStore) ListCatalogByCategory(_ context.Context, category string, activeOnly bool) ([]models.CatalogItem, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	var items []models.CatalogItem
	for _, item := range m.CatalogItems {
		if item.Category == category {
			if activeOnly && !item.IsActive {
				continue
			}
			items = append(items, *item)
		}
	}
	return items, nil
}

func (m *MockStore) ListActiveCatalog(_ context.Context) ([]models.CatalogItem, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	var items []models.CatalogItem
	for _, item := range m.CatalogItems {
		if item.IsActive {
			items = append(items, *item)
		}
	}
	return items, nil
}

func (m *MockStore) CreateCatalogItem(_ context.Context, item *models.CatalogItem) error {
	if m.CreateCatalogItemErr != nil {
		return m.CreateCatalogItemErr
	}
	if m.Err != nil {
		return m.Err
	}
	m.CatalogItems[item.ID] = item
	return nil
}

func (m *MockStore) UpdateCatalogItem(_ context.Context, id string, updates map[string]interface{}) error {
	if m.UpdateCatalogItemErr != nil {
		return m.UpdateCatalogItemErr
	}
	if m.Err != nil {
		return m.Err
	}
	return nil
}

func (m *MockStore) DeleteCatalogItem(_ context.Context, id string) error {
	if m.DeleteCatalogItemErr != nil {
		return m.DeleteCatalogItemErr
	}
	if m.Err != nil {
		return m.Err
	}
	delete(m.CatalogItems, id)
	return nil
}

func (m *MockStore) ListAllCatalogItems(_ context.Context, category string, limit, offset int) ([]models.CatalogItem, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	var items []models.CatalogItem
	for _, item := range m.CatalogItems {
		if category != "" && item.Category != category {
			continue
		}
		items = append(items, *item)
	}
	if offset >= len(items) {
		return nil, nil
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end], nil
}
