package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

var (
	FirestoreClient *firestore.Client
	ctx             = context.Background()
)

// InitFirestore initializes the Firestore client
func InitFirestore() error {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		return fmt.Errorf("GOOGLE_CLOUD_PROJECT environment variable not set")
	}

	var client *firestore.Client
	var err error

	// Check if we're running in a local environment
	if credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); credentialsPath != "" {
		client, err = firestore.NewClient(ctx, projectID, option.WithCredentialsFile(credentialsPath))
	} else {
		// Use default credentials in Cloud Run
		client, err = firestore.NewClient(ctx, projectID)
	}

	if err != nil {
		return fmt.Errorf("failed to create Firestore client: %v", err)
	}

	FirestoreClient = client
	log.Printf("Firestore client initialized for project: %s", projectID)
	return nil
}

// CloseFirestore closes the Firestore client
func CloseFirestore() {
	if FirestoreClient != nil {
		FirestoreClient.Close()
		log.Println("Firestore client closed")
	}
}

// Collections
const (
	UsersCollection       = "users"
	CasesCollection       = "cases"
	SubmissionsCollection = "submissions"
	OrdersCollection      = "orders"
)