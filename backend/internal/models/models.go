package models

import (
	"time"
)

// User represents a coffee detective user
type User struct {
	ID             string    `firestore:"id" json:"id"`
	Name           string    `firestore:"name" json:"name"`
	Email          string    `firestore:"email" json:"email"`
	Picture        string    `firestore:"picture" json:"picture"`
	Type           string    `firestore:"type" json:"type"`     // regular, admin
	Level          string    `firestore:"level" json:"level"`   // beginner, intermediate, expert
	Score          int       `firestore:"score" json:"score"`
	Points         int       `firestore:"points" json:"points"`
	CasesAttempted int       `firestore:"cases_attempted" json:"cases_attempted"`
	CasesSolved    int       `firestore:"cases_solved" json:"cases_solved"`
	CasesCount     int       `firestore:"cases_count" json:"cases_count"`
	Accuracy       float64   `firestore:"accuracy" json:"accuracy"`
	Badges         []string  `firestore:"badges" json:"badges"`
	CreatedAt      time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt      time.Time `firestore:"updated_at" json:"updated_at"`
}

// CoffeeCase represents a coffee mystery case
type CoffeeCase struct {
	ID          string       `firestore:"id" json:"id"`
	Name        string       `firestore:"name" json:"name"`
	Description string       `firestore:"description" json:"description"`
	Price       int          `firestore:"price" json:"price"`
	Coffees     []CoffeeItem `firestore:"coffees" json:"coffees"`
	CreatedAt   time.Time    `firestore:"created_at" json:"created_at"`
	IsActive    bool         `firestore:"is_active" json:"is_active"`
}

// CoffeeItem represents a single coffee in a case
type CoffeeItem struct {
	ID          string `firestore:"id" json:"id"`
	Name        string `firestore:"name" json:"name"`
	Region      string `firestore:"region" json:"region"`
	Variety     string `firestore:"variety" json:"variety"`
	Process     string `firestore:"process" json:"process"`
	TastingNotes string `firestore:"tasting_notes" json:"tasting_notes"`
	Farm        string `firestore:"farm" json:"farm"`
	Altitude    int    `firestore:"altitude" json:"altitude"`
}

// Submission represents a user's case submission
type Submission struct {
	ID              string           `firestore:"id" json:"id"`
	UserID          string           `firestore:"user_id" json:"user_id"`
	CaseID          string           `firestore:"case_id" json:"case_id"`
	OrderID         string           `firestore:"order_id" json:"order_id"` // 6-character order ID provided to customers
	CoffeeAnswers   []CoffeeAnswer   `firestore:"coffee_answers" json:"coffee_answers"`
	FavoriteCoffee  string           `firestore:"favorite_coffee" json:"favorite_coffee"`
	BrewingMethod   string           `firestore:"brewing_method" json:"brewing_method"`
	Score           int              `firestore:"score" json:"score"`
	Accuracy        float64          `firestore:"accuracy" json:"accuracy"`
	SubmittedAt     time.Time        `firestore:"submitted_at" json:"submitted_at"`
	ProcessedAt     *time.Time       `firestore:"processed_at" json:"processed_at"`
}

// CoffeeAnswer represents a user's answer for a single coffee
type CoffeeAnswer struct {
	CoffeeID     string `firestore:"coffee_id" json:"coffee_id"`
	Region       string `firestore:"region" json:"region"`
	Variety      string `firestore:"variety" json:"variety"`
	Process      string `firestore:"process" json:"process"`
	TasteNote1   string `firestore:"taste_note_1" json:"taste_note_1"`
	TasteNote2   string `firestore:"taste_note_2" json:"taste_note_2"`
	Points       int    `firestore:"points" json:"points"`
}

// Order represents a coffee case order
type Order struct {
	ID              string     `firestore:"id" json:"id"`
	OrderID         string     `firestore:"order_id" json:"order_id"`         // 6-character unique order ID
	UserID          string     `firestore:"user_id" json:"user_id"`
	CaseID          string     `firestore:"case_id" json:"case_id"`
	ContactInfo     string     `firestore:"contact_info" json:"contact_info"`
	Status          string     `firestore:"status" json:"status"` // pending, confirmed, shipped, delivered
	TotalAmount     int        `firestore:"total_amount" json:"total_amount"`
	IsSubmissionUsed bool      `firestore:"is_submission_used" json:"is_submission_used"` // Whether order ID was used for submission
	SubmissionUsedBy string    `firestore:"submission_used_by" json:"submission_used_by"` // User ID who used the order for submission
	SubmissionUsedAt *time.Time `firestore:"submission_used_at" json:"submission_used_at"` // When the order ID was used
	CreatedAt       time.Time  `firestore:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `firestore:"updated_at" json:"updated_at"`
}

// LeaderboardEntry represents a leaderboard entry
type LeaderboardEntry struct {
	UserID        string  `firestore:"user_id" json:"user_id"`
	Points        int     `firestore:"points" json:"points"`
	Accuracy      float64 `firestore:"accuracy" json:"accuracy"`
	CasesCount    int     `firestore:"cases_count" json:"cases_count"`
	Badges        []string `firestore:"badges" json:"badges"`
	Rank          int     `json:"rank"`
}

// LeaderboardEntryWithUser represents a leaderboard entry with user information
type LeaderboardEntryWithUser struct {
	UserID        string  `json:"user_id"`
	DetectiveName string  `json:"detective_name"` // Populated from User.Name
	Points        int     `json:"points"`
	Accuracy      float64 `json:"accuracy"`
	CasesCount    int     `json:"cases_count"`
	Badges        []string `json:"badges"`
	Rank          int     `json:"rank"`
}

// CatalogItem represents a catalog entry for dropdowns
type CatalogItem struct {
	ID          string `firestore:"id" json:"id"`
	Value       string `firestore:"value" json:"value"`         // The option value (e.g., "central_valley")
	Label       string `firestore:"label" json:"label"`         // The display name (e.g., "Valle Central")
	Category    string `firestore:"category" json:"category"`   // "region", "variety", or "process"
	IsActive    bool   `firestore:"is_active" json:"is_active"`
	DisplayOrder int   `firestore:"display_order" json:"display_order"`
	CreatedAt   time.Time `firestore:"created_at" json:"created_at"`
}