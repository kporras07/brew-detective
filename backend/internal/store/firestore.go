package store

import (
	"context"
	"fmt"

	"brew-detective-backend/internal/models"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

const (
	usersCollection       = "users"
	casesCollection       = "cases"
	submissionsCollection = "submissions"
	ordersCollection      = "orders"
	catalogCollection     = "catalog"
)

// FirestoreStore implements Store using Google Cloud Firestore.
type FirestoreStore struct {
	Client *firestore.Client
}

// NewFirestoreStore creates a new FirestoreStore.
func NewFirestoreStore(client *firestore.Client) *FirestoreStore {
	return &FirestoreStore{Client: client}
}

// --- Users ---

func (s *FirestoreStore) GetUser(ctx context.Context, id string) (*models.User, error) {
	doc, err := s.Client.Collection(usersCollection).Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}
	var user models.User
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *FirestoreStore) SetUser(ctx context.Context, user *models.User) error {
	_, err := s.Client.Collection(usersCollection).Doc(user.ID).Set(ctx, user)
	return err
}

func (s *FirestoreStore) ListUsers(ctx context.Context) ([]models.User, error) {
	iter := s.Client.Collection(usersCollection).OrderBy("name", firestore.Asc).Documents(ctx)
	var users []models.User
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var user models.User
		if err := doc.DataTo(&user); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// --- Cases ---

func (s *FirestoreStore) GetCase(ctx context.Context, id string) (*models.CoffeeCase, error) {
	doc, err := s.Client.Collection(casesCollection).Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}
	var coffeeCase models.CoffeeCase
	if err := doc.DataTo(&coffeeCase); err != nil {
		return nil, err
	}
	return &coffeeCase, nil
}

func (s *FirestoreStore) CreateCase(ctx context.Context, coffeeCase *models.CoffeeCase) error {
	_, err := s.Client.Collection(casesCollection).Doc(coffeeCase.ID).Set(ctx, coffeeCase)
	return err
}

func (s *FirestoreStore) UpdateCase(ctx context.Context, id string, updates map[string]interface{}) error {
	var firestoreUpdates []firestore.Update
	for key, value := range updates {
		firestoreUpdates = append(firestoreUpdates, firestore.Update{Path: key, Value: value})
	}
	_, err := s.Client.Collection(casesCollection).Doc(id).Update(ctx, firestoreUpdates)
	return err
}

func (s *FirestoreStore) DeleteCase(ctx context.Context, id string) error {
	_, err := s.Client.Collection(casesCollection).Doc(id).Delete(ctx)
	return err
}

func (s *FirestoreStore) ListCases(ctx context.Context, limit, offset int) ([]models.CoffeeCase, error) {
	iter := s.Client.Collection(casesCollection).
		OrderBy("created_at", firestore.Desc).
		Limit(limit).Offset(offset).Documents(ctx)
	var cases []models.CoffeeCase
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var c models.CoffeeCase
		if err := doc.DataTo(&c); err != nil {
			return nil, err
		}
		cases = append(cases, c)
	}
	return cases, nil
}

func (s *FirestoreStore) ListActiveCases(ctx context.Context) ([]models.CoffeeCase, error) {
	iter := s.Client.Collection(casesCollection).Where("is_active", "==", true).Documents(ctx)
	var cases []models.CoffeeCase
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var c models.CoffeeCase
		if err := doc.DataTo(&c); err != nil {
			return nil, err
		}
		cases = append(cases, c)
	}
	return cases, nil
}

func (s *FirestoreStore) GetActiveCase(ctx context.Context) (*models.CoffeeCase, error) {
	iter := s.Client.Collection(casesCollection).
		Where("is_active", "==", true).Limit(1).Documents(ctx)
	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("no active case found")
	}
	if err != nil {
		return nil, err
	}
	var c models.CoffeeCase
	if err := doc.DataTo(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

// --- Orders ---

func (s *FirestoreStore) GetOrder(ctx context.Context, id string) (*models.Order, error) {
	doc, err := s.Client.Collection(ordersCollection).Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}
	var order models.Order
	if err := doc.DataTo(&order); err != nil {
		return nil, err
	}
	return &order, nil
}

func (s *FirestoreStore) CreateOrder(ctx context.Context, order *models.Order) error {
	_, err := s.Client.Collection(ordersCollection).Doc(order.ID).Set(ctx, order)
	return err
}

func (s *FirestoreStore) SetOrder(ctx context.Context, order *models.Order) error {
	_, err := s.Client.Collection(ordersCollection).Doc(order.ID).Set(ctx, order)
	return err
}

func (s *FirestoreStore) ListOrders(ctx context.Context, limit, offset int) ([]models.Order, error) {
	iter := s.Client.Collection(ordersCollection).
		OrderBy("created_at", firestore.Desc).
		Limit(limit).Offset(offset).Documents(ctx)
	var orders []models.Order
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var o models.Order
		if err := doc.DataTo(&o); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (s *FirestoreStore) GetOrderByOrderID(ctx context.Context, orderID string) (*models.Order, error) {
	docs, err := s.Client.Collection(ordersCollection).
		Where("order_id", "==", orderID).Limit(1).Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return nil, fmt.Errorf("order not found")
	}
	var order models.Order
	if err := docs[0].DataTo(&order); err != nil {
		return nil, err
	}
	return &order, nil
}

func (s *FirestoreStore) UpdateOrderFields(ctx context.Context, docID string, updates map[string]interface{}) error {
	// We need to find the document by its Firestore document ID
	var firestoreUpdates []firestore.Update
	for key, value := range updates {
		firestoreUpdates = append(firestoreUpdates, firestore.Update{Path: key, Value: value})
	}
	_, err := s.Client.Collection(ordersCollection).Doc(docID).Update(ctx, firestoreUpdates)
	return err
}

// --- Submissions ---

func (s *FirestoreStore) CreateSubmission(ctx context.Context, submission *models.Submission) error {
	_, err := s.Client.Collection(submissionsCollection).Doc(submission.ID).Set(ctx, submission)
	return err
}

func (s *FirestoreStore) ListSubmissionsByUser(ctx context.Context, userID string, limit, offset int) ([]models.Submission, error) {
	iter := s.Client.Collection(submissionsCollection).
		Where("user_id", "==", userID).
		OrderBy("submitted_at", firestore.Desc).
		Limit(limit).Offset(offset).Documents(ctx)
	var submissions []models.Submission
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var sub models.Submission
		if err := doc.DataTo(&sub); err != nil {
			return nil, err
		}
		submissions = append(submissions, sub)
	}
	return submissions, nil
}

func (s *FirestoreStore) ListSubmissionsByCase(ctx context.Context, caseID string) ([]models.Submission, error) {
	iter := s.Client.Collection(submissionsCollection).
		Where("case_id", "==", caseID).Documents(ctx)
	var submissions []models.Submission
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var sub models.Submission
		if err := doc.DataTo(&sub); err != nil {
			return nil, err
		}
		submissions = append(submissions, sub)
	}
	return submissions, nil
}

// --- Catalog ---

func (s *FirestoreStore) ListCatalogByCategory(ctx context.Context, category string, activeOnly bool) ([]models.CatalogItem, error) {
	query := s.Client.Collection(catalogCollection).Where("category", "==", category)
	if activeOnly {
		query = query.Where("is_active", "==", true)
	}
	iter := query.Documents(ctx)
	var items []models.CatalogItem
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var item models.CatalogItem
		if err := doc.DataTo(&item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *FirestoreStore) ListActiveCatalog(ctx context.Context) ([]models.CatalogItem, error) {
	iter := s.Client.Collection(catalogCollection).Where("is_active", "==", true).Documents(ctx)
	var items []models.CatalogItem
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var item models.CatalogItem
		if err := doc.DataTo(&item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *FirestoreStore) CreateCatalogItem(ctx context.Context, item *models.CatalogItem) error {
	_, err := s.Client.Collection(catalogCollection).Doc(item.ID).Set(ctx, item)
	return err
}

func (s *FirestoreStore) UpdateCatalogItem(ctx context.Context, id string, updates map[string]interface{}) error {
	var firestoreUpdates []firestore.Update
	for key, value := range updates {
		firestoreUpdates = append(firestoreUpdates, firestore.Update{Path: key, Value: value})
	}
	_, err := s.Client.Collection(catalogCollection).Doc(id).Update(ctx, firestoreUpdates)
	return err
}

func (s *FirestoreStore) DeleteCatalogItem(ctx context.Context, id string) error {
	_, err := s.Client.Collection(catalogCollection).Doc(id).Delete(ctx)
	return err
}

func (s *FirestoreStore) ListAllCatalogItems(ctx context.Context, category string, limit, offset int) ([]models.CatalogItem, error) {
	query := s.Client.Collection(catalogCollection).
		OrderBy("category", firestore.Asc).
		OrderBy("display_order", firestore.Asc).
		Limit(limit).Offset(offset)
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
			return nil, err
		}
		var item models.CatalogItem
		if err := doc.DataTo(&item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
