package handlers

import (
	"testing"

	"brew-detective-backend/internal/models"
)

func TestGetMatchedTastingNote(t *testing.T) {
	t.Run("exact match single note", func(t *testing.T) {
		result := getMatchedTastingNote("chocolate", "chocolate")
		if result != "chocolate" {
			t.Errorf("expected %q, got %q", "chocolate", result)
		}
	})

	t.Run("case insensitive match", func(t *testing.T) {
		result := getMatchedTastingNote("Chocolate", "chocolate")
		if result != "chocolate" {
			t.Errorf("expected %q, got %q", "chocolate", result)
		}
	})

	t.Run("match with leading/trailing spaces", func(t *testing.T) {
		result := getMatchedTastingNote("  chocolate  ", "chocolate")
		if result != "chocolate" {
			t.Errorf("expected %q, got %q", "chocolate", result)
		}
	})

	t.Run("match against comma-separated notes", func(t *testing.T) {
		result := getMatchedTastingNote("caramel", "chocolate, caramel, berry")
		if result != "caramel" {
			t.Errorf("expected %q, got %q", "caramel", result)
		}
	})

	t.Run("partial match - user note contains correct note", func(t *testing.T) {
		result := getMatchedTastingNote("dark chocolate", "chocolate")
		if result != "chocolate" {
			t.Errorf("expected %q, got %q", "chocolate", result)
		}
	})

	t.Run("partial match - correct note contains user note", func(t *testing.T) {
		result := getMatchedTastingNote("berry", "blueberry")
		if result != "blueberry" {
			t.Errorf("expected %q, got %q", "blueberry", result)
		}
	})

	t.Run("no match returns empty string", func(t *testing.T) {
		result := getMatchedTastingNote("vanilla", "chocolate, caramel")
		if result != "" {
			t.Errorf("expected empty string, got %q", result)
		}
	})

	t.Run("empty user note returns empty string", func(t *testing.T) {
		result := getMatchedTastingNote("", "chocolate")
		if result != "" {
			t.Errorf("expected empty string, got %q", result)
		}
	})

	t.Run("empty correct notes returns empty string", func(t *testing.T) {
		result := getMatchedTastingNote("chocolate", "")
		if result != "" {
			t.Errorf("expected empty string, got %q", result)
		}
	})

	t.Run("both empty returns empty string", func(t *testing.T) {
		result := getMatchedTastingNote("", "")
		if result != "" {
			t.Errorf("expected empty string, got %q", result)
		}
	})
}

func TestMatchesTastingNotes(t *testing.T) {
	t.Run("returns true for matching note", func(t *testing.T) {
		if !matchesTastingNotes("chocolate", "chocolate, caramel") {
			t.Error("expected true for matching note")
		}
	})

	t.Run("returns false for non-matching note", func(t *testing.T) {
		if matchesTastingNotes("vanilla", "chocolate, caramel") {
			t.Error("expected false for non-matching note")
		}
	})
}

func TestCalculateScoreDefault(t *testing.T) {
	t.Run("empty submission returns zero", func(t *testing.T) {
		submission := &models.Submission{
			CoffeeAnswers: []models.CoffeeAnswer{},
		}
		score, accuracy := calculateScoreDefault(submission)
		if score != 0 || accuracy != 0.0 {
			t.Errorf("expected score=0, accuracy=0.0, got score=%d, accuracy=%f", score, accuracy)
		}
	})

	t.Run("all fields filled scores full marks", func(t *testing.T) {
		submission := &models.Submission{
			CoffeeAnswers: []models.CoffeeAnswer{
				{Region: "central_valley", Variety: "caturra", Process: "washed"},
			},
		}
		score, accuracy := calculateScoreDefault(submission)
		if accuracy != 1.0 {
			t.Errorf("expected accuracy=1.0, got %f", accuracy)
		}
		if score != 100 {
			t.Errorf("expected score=100, got %d", score)
		}
	})

	t.Run("partial answers score proportionally", func(t *testing.T) {
		submission := &models.Submission{
			CoffeeAnswers: []models.CoffeeAnswer{
				{Region: "central_valley", Variety: "", Process: "washed"},
			},
		}
		score, accuracy := calculateScoreDefault(submission)
		// 2 out of 3 questions answered
		expectedAccuracy := 2.0 / 3.0
		if accuracy < expectedAccuracy-0.01 || accuracy > expectedAccuracy+0.01 {
			t.Errorf("expected accuracy≈%f, got %f", expectedAccuracy, accuracy)
		}
		// score = 100 * (2/3) * 1 = 66
		if score != 66 {
			t.Errorf("expected score=66, got %d", score)
		}
	})

	t.Run("multiple coffees multiply score", func(t *testing.T) {
		submission := &models.Submission{
			CoffeeAnswers: []models.CoffeeAnswer{
				{Region: "central_valley", Variety: "caturra", Process: "washed"},
				{Region: "west_valley", Variety: "geisha", Process: "natural"},
			},
		}
		score, accuracy := calculateScoreDefault(submission)
		if accuracy != 1.0 {
			t.Errorf("expected accuracy=1.0, got %f", accuracy)
		}
		// score = 100 * 1.0 * 2 = 200
		if score != 200 {
			t.Errorf("expected score=200, got %d", score)
		}
	})

	t.Run("no fields filled scores zero", func(t *testing.T) {
		submission := &models.Submission{
			CoffeeAnswers: []models.CoffeeAnswer{
				{Region: "", Variety: "", Process: ""},
			},
		}
		score, accuracy := calculateScoreDefault(submission)
		if score != 0 || accuracy != 0.0 {
			t.Errorf("expected score=0, accuracy=0.0, got score=%d, accuracy=%f", score, accuracy)
		}
	})
}

// Helper to build a test case with all questions enabled and the given coffees.
func makeTestCase(coffees []models.CoffeeItem) *models.CoffeeCase {
	return &models.CoffeeCase{
		ID:   "test-case",
		Name: "Test Case",
		Coffees: coffees,
		EnabledQuestions: models.EnabledQuestions{
			Region:         true,
			Variety:        true,
			Process:        true,
			TasteNote1:     true,
			TasteNote2:     true,
			FavoriteCoffee: true,
			BrewingMethod:  true,
		},
		IsActive: true,
	}
}

func TestCalculateScoreWithCase(t *testing.T) {
	t.Run("nil case falls back to default scoring", func(t *testing.T) {
		submission := &models.Submission{
			CoffeeAnswers: []models.CoffeeAnswer{
				{CoffeeID: "c1", Region: "r", Variety: "v", Process: "p"},
			},
		}
		score, accuracy := calculateScoreWithCase(submission, nil)
		if accuracy != 1.0 {
			t.Errorf("expected accuracy=1.0, got %f", accuracy)
		}
		if score != 100 {
			t.Errorf("expected score=100, got %d", score)
		}
	})

	t.Run("all correct answers with all questions enabled", func(t *testing.T) {
		coffees := []models.CoffeeItem{
			{ID: "c1", Region: "Central Valley", Variety: "Caturra", Process: "Washed", TastingNotes: "chocolate, caramel"},
		}
		testCase := makeTestCase(coffees)

		submission := &models.Submission{
			CoffeeAnswers: []models.CoffeeAnswer{
				{CoffeeID: "c1", Region: "Central Valley", Variety: "Caturra", Process: "Washed", TasteNote1: "chocolate", TasteNote2: "caramel"},
			},
			FavoriteCoffee: "c1",
			BrewingMethod:  "pour_over",
		}

		score, accuracy := calculateScoreWithCase(submission, testCase)
		if accuracy != 1.0 {
			t.Errorf("expected accuracy=1.0, got %f", accuracy)
		}
		// 5 questions per coffee * 1 coffee = 5 total, all correct
		// score = 100 * 1.0 * 1 + 50 (fav) + 50 (brew) = 200
		if score != 200 {
			t.Errorf("expected score=200, got %d", score)
		}
	})

	t.Run("case insensitive matching", func(t *testing.T) {
		coffees := []models.CoffeeItem{
			{ID: "c1", Region: "Central Valley", Variety: "Caturra", Process: "Washed", TastingNotes: "chocolate"},
		}
		testCase := makeTestCase(coffees)
		testCase.EnabledQuestions.TasteNote2 = false
		testCase.EnabledQuestions.FavoriteCoffee = false
		testCase.EnabledQuestions.BrewingMethod = false

		submission := &models.Submission{
			CoffeeAnswers: []models.CoffeeAnswer{
				{CoffeeID: "c1", Region: "central valley", Variety: "caturra", Process: "washed", TasteNote1: "Chocolate"},
			},
		}

		_, accuracy := calculateScoreWithCase(submission, testCase)
		if accuracy != 1.0 {
			t.Errorf("expected accuracy=1.0 with case-insensitive match, got %f", accuracy)
		}
	})

	t.Run("no correct answers scores zero", func(t *testing.T) {
		coffees := []models.CoffeeItem{
			{ID: "c1", Region: "Central Valley", Variety: "Caturra", Process: "Washed", TastingNotes: "chocolate"},
		}
		testCase := makeTestCase(coffees)
		testCase.EnabledQuestions.TasteNote2 = false
		testCase.EnabledQuestions.FavoriteCoffee = false
		testCase.EnabledQuestions.BrewingMethod = false

		submission := &models.Submission{
			CoffeeAnswers: []models.CoffeeAnswer{
				{CoffeeID: "c1", Region: "West Valley", Variety: "Geisha", Process: "Natural", TasteNote1: "berry"},
			},
		}

		score, accuracy := calculateScoreWithCase(submission, testCase)
		if accuracy != 0.0 {
			t.Errorf("expected accuracy=0.0, got %f", accuracy)
		}
		if score != 0 {
			t.Errorf("expected score=0, got %d", score)
		}
	})

	t.Run("disabled questions are not scored", func(t *testing.T) {
		coffees := []models.CoffeeItem{
			{ID: "c1", Region: "Central Valley", Variety: "Caturra", Process: "Washed", TastingNotes: "chocolate"},
		}
		testCase := makeTestCase(coffees)
		// Only region enabled
		testCase.EnabledQuestions.Variety = false
		testCase.EnabledQuestions.Process = false
		testCase.EnabledQuestions.TasteNote1 = false
		testCase.EnabledQuestions.TasteNote2 = false
		testCase.EnabledQuestions.FavoriteCoffee = false
		testCase.EnabledQuestions.BrewingMethod = false

		submission := &models.Submission{
			CoffeeAnswers: []models.CoffeeAnswer{
				{CoffeeID: "c1", Region: "Central Valley", Variety: "Wrong", Process: "Wrong"},
			},
		}

		_, accuracy := calculateScoreWithCase(submission, testCase)
		// Only region counts, and it's correct
		if accuracy != 1.0 {
			t.Errorf("expected accuracy=1.0 (only region enabled & correct), got %f", accuracy)
		}
	})

	t.Run("no questions enabled returns zero", func(t *testing.T) {
		coffees := []models.CoffeeItem{
			{ID: "c1", Region: "Central Valley"},
		}
		testCase := makeTestCase(coffees)
		testCase.EnabledQuestions = models.EnabledQuestions{}

		submission := &models.Submission{
			CoffeeAnswers: []models.CoffeeAnswer{
				{CoffeeID: "c1", Region: "Central Valley"},
			},
		}

		score, accuracy := calculateScoreWithCase(submission, testCase)
		if score != 0 || accuracy != 0.0 {
			t.Errorf("expected score=0, accuracy=0.0, got score=%d, accuracy=%f", score, accuracy)
		}
	})

	t.Run("duplicate tasting note not double-counted", func(t *testing.T) {
		coffees := []models.CoffeeItem{
			{ID: "c1", Region: "R", Variety: "V", Process: "P", TastingNotes: "chocolate, caramel"},
		}
		testCase := makeTestCase(coffees)
		// Disable non-tasting questions to isolate
		testCase.EnabledQuestions.Region = false
		testCase.EnabledQuestions.Variety = false
		testCase.EnabledQuestions.Process = false
		testCase.EnabledQuestions.FavoriteCoffee = false
		testCase.EnabledQuestions.BrewingMethod = false

		submission := &models.Submission{
			CoffeeAnswers: []models.CoffeeAnswer{
				{CoffeeID: "c1", TasteNote1: "chocolate", TasteNote2: "chocolate"},
			},
		}

		// 2 tasting note questions, only 1 should be awarded (second is duplicate)
		_, accuracy := calculateScoreWithCase(submission, testCase)
		if accuracy != 0.5 {
			t.Errorf("expected accuracy=0.5 (duplicate note), got %f", accuracy)
		}
	})

	t.Run("two distinct tasting notes both score", func(t *testing.T) {
		coffees := []models.CoffeeItem{
			{ID: "c1", TastingNotes: "chocolate, caramel"},
		}
		testCase := makeTestCase(coffees)
		testCase.EnabledQuestions.Region = false
		testCase.EnabledQuestions.Variety = false
		testCase.EnabledQuestions.Process = false
		testCase.EnabledQuestions.FavoriteCoffee = false
		testCase.EnabledQuestions.BrewingMethod = false

		submission := &models.Submission{
			CoffeeAnswers: []models.CoffeeAnswer{
				{CoffeeID: "c1", TasteNote1: "chocolate", TasteNote2: "caramel"},
			},
		}

		_, accuracy := calculateScoreWithCase(submission, testCase)
		if accuracy != 1.0 {
			t.Errorf("expected accuracy=1.0, got %f", accuracy)
		}
	})

	t.Run("bonus points only when questions enabled and answered", func(t *testing.T) {
		coffees := []models.CoffeeItem{
			{ID: "c1", Region: "R"},
		}
		testCase := makeTestCase(coffees)
		// Only region + favorite + brewing
		testCase.EnabledQuestions.Variety = false
		testCase.EnabledQuestions.Process = false
		testCase.EnabledQuestions.TasteNote1 = false
		testCase.EnabledQuestions.TasteNote2 = false

		submission := &models.Submission{
			CoffeeAnswers: []models.CoffeeAnswer{
				{CoffeeID: "c1", Region: "R"},
			},
			FavoriteCoffee: "c1",
			BrewingMethod:  "pour_over",
		}

		score, _ := calculateScoreWithCase(submission, testCase)
		// base: 100 * 1.0 * 1 = 100, + 50 + 50 = 200
		if score != 200 {
			t.Errorf("expected score=200, got %d", score)
		}
	})

	t.Run("no bonus when favorite/brewing disabled", func(t *testing.T) {
		coffees := []models.CoffeeItem{
			{ID: "c1", Region: "R"},
		}
		testCase := makeTestCase(coffees)
		testCase.EnabledQuestions.Variety = false
		testCase.EnabledQuestions.Process = false
		testCase.EnabledQuestions.TasteNote1 = false
		testCase.EnabledQuestions.TasteNote2 = false
		testCase.EnabledQuestions.FavoriteCoffee = false
		testCase.EnabledQuestions.BrewingMethod = false

		submission := &models.Submission{
			CoffeeAnswers: []models.CoffeeAnswer{
				{CoffeeID: "c1", Region: "R"},
			},
			FavoriteCoffee: "c1",
			BrewingMethod:  "pour_over",
		}

		score, _ := calculateScoreWithCase(submission, testCase)
		// base: 100 * 1.0 * 1 = 100, no bonus
		if score != 100 {
			t.Errorf("expected score=100, got %d", score)
		}
	})

	t.Run("unknown coffee ID is skipped", func(t *testing.T) {
		coffees := []models.CoffeeItem{
			{ID: "c1", Region: "R"},
		}
		testCase := makeTestCase(coffees)
		testCase.EnabledQuestions.Variety = false
		testCase.EnabledQuestions.Process = false
		testCase.EnabledQuestions.TasteNote1 = false
		testCase.EnabledQuestions.TasteNote2 = false
		testCase.EnabledQuestions.FavoriteCoffee = false
		testCase.EnabledQuestions.BrewingMethod = false

		submission := &models.Submission{
			CoffeeAnswers: []models.CoffeeAnswer{
				{CoffeeID: "unknown", Region: "R"},
			},
		}

		_, accuracy := calculateScoreWithCase(submission, testCase)
		// The coffee isn't found so no correct answers, but totalQuestions is still 1
		if accuracy != 0.0 {
			t.Errorf("expected accuracy=0.0 for unknown coffee, got %f", accuracy)
		}
	})

	t.Run("multiple coffees scored independently", func(t *testing.T) {
		coffees := []models.CoffeeItem{
			{ID: "c1", Region: "Region A", Variety: "V1", Process: "P1", TastingNotes: "note1"},
			{ID: "c2", Region: "Region B", Variety: "V2", Process: "P2", TastingNotes: "note2"},
		}
		testCase := makeTestCase(coffees)
		testCase.EnabledQuestions.TasteNote2 = false
		testCase.EnabledQuestions.FavoriteCoffee = false
		testCase.EnabledQuestions.BrewingMethod = false

		submission := &models.Submission{
			CoffeeAnswers: []models.CoffeeAnswer{
				{CoffeeID: "c1", Region: "Region A", Variety: "V1", Process: "P1", TasteNote1: "note1"},
				{CoffeeID: "c2", Region: "Wrong", Variety: "Wrong", Process: "Wrong", TasteNote1: "wrong"},
			},
		}

		_, accuracy := calculateScoreWithCase(submission, testCase)
		// 4 correct out of 8 total (4 per coffee * 2 coffees)
		expectedAccuracy := 4.0 / 8.0
		if accuracy < expectedAccuracy-0.01 || accuracy > expectedAccuracy+0.01 {
			t.Errorf("expected accuracy≈%f, got %f", expectedAccuracy, accuracy)
		}
	})

	t.Run("empty answer fields count as wrong", func(t *testing.T) {
		coffees := []models.CoffeeItem{
			{ID: "c1", Region: "R", Variety: "V", Process: "P"},
		}
		testCase := makeTestCase(coffees)
		testCase.EnabledQuestions.TasteNote1 = false
		testCase.EnabledQuestions.TasteNote2 = false
		testCase.EnabledQuestions.FavoriteCoffee = false
		testCase.EnabledQuestions.BrewingMethod = false

		submission := &models.Submission{
			CoffeeAnswers: []models.CoffeeAnswer{
				{CoffeeID: "c1", Region: "", Variety: "", Process: ""},
			},
		}

		_, accuracy := calculateScoreWithCase(submission, testCase)
		if accuracy != 0.0 {
			t.Errorf("expected accuracy=0.0 for empty answers, got %f", accuracy)
		}
	})
}

func TestCalculateScoreWithCasePerCoffeeOverride(t *testing.T) {
	t.Run("coffee with override uses override questions", func(t *testing.T) {
		regionOnly := models.EnabledQuestions{Region: true}
		coffees := []models.CoffeeItem{
			{
				ID: "c1", Region: "Central Valley", Variety: "Caturra", Process: "Washed",
				EnabledQuestions: &regionOnly,
			},
		}
		testCase := makeTestCase(coffees)

		submission := &models.Submission{
			CoffeeAnswers: []models.CoffeeAnswer{
				{CoffeeID: "c1", Region: "Central Valley", Variety: "Wrong", Process: "Wrong"},
			},
		}

		_, accuracy := calculateScoreWithCase(submission, testCase)
		if accuracy != 1.0 {
			t.Errorf("expected accuracy=1.0 (only region from override), got %f", accuracy)
		}
	})

	t.Run("coffee without override uses case-level questions", func(t *testing.T) {
		regionOnly := models.EnabledQuestions{Region: true}
		coffees := []models.CoffeeItem{
			{
				ID: "c1", Region: "R1", Variety: "V1",
				EnabledQuestions: &regionOnly,
			},
			{
				ID: "c2", Region: "R2", Variety: "V2",
			},
		}
		testCase := &models.CoffeeCase{
			ID:      "test",
			Coffees: coffees,
			EnabledQuestions: models.EnabledQuestions{
				Region:  true,
				Variety: true,
			},
		}

		submission := &models.Submission{
			CoffeeAnswers: []models.CoffeeAnswer{
				{CoffeeID: "c1", Region: "R1", Variety: "V1"},
				{CoffeeID: "c2", Region: "R2", Variety: "V2"},
			},
		}

		_, accuracy := calculateScoreWithCase(submission, testCase)
		// c1: 1 question (region), 1 correct. c2: 2 questions (region+variety), 2 correct. Total: 3/3
		if accuracy != 1.0 {
			t.Errorf("expected accuracy=1.0, got %f", accuracy)
		}
	})

	t.Run("mixed override and fallback with partial answers", func(t *testing.T) {
		regionOnly := models.EnabledQuestions{Region: true}
		coffees := []models.CoffeeItem{
			{ID: "c1", Region: "R1", Variety: "V1", EnabledQuestions: &regionOnly},
			{ID: "c2", Region: "R2", Variety: "V2"},
		}
		testCase := &models.CoffeeCase{
			ID:      "test",
			Coffees: coffees,
			EnabledQuestions: models.EnabledQuestions{
				Region:  true,
				Variety: true,
			},
		}

		submission := &models.Submission{
			CoffeeAnswers: []models.CoffeeAnswer{
				{CoffeeID: "c1", Region: "Wrong"},
				{CoffeeID: "c2", Region: "R2", Variety: "Wrong"},
			},
		}

		_, accuracy := calculateScoreWithCase(submission, testCase)
		// c1: 1 question, 0 correct. c2: 2 questions, 1 correct. Total: 1/3
		expected := 1.0 / 3.0
		if accuracy < expected-0.01 || accuracy > expected+0.01 {
			t.Errorf("expected accuracy≈%f, got %f", expected, accuracy)
		}
	})
}

func TestCalculateScoreWithCaseAdditionalQuestions(t *testing.T) {
	t.Run("correct additional answer awards points", func(t *testing.T) {
		coffees := []models.CoffeeItem{
			{
				ID: "c1", Region: "R",
				AdditionalQuestions: []models.AdditionalQuestion{
					{ID: "aq1", Question: "What altitude?", Options: []string{"High", "Low"}, CorrectOption: "High", Points: 25},
				},
			},
		}
		testCase := &models.CoffeeCase{
			ID:      "test",
			Coffees: coffees,
			EnabledQuestions: models.EnabledQuestions{
				Region: true,
			},
		}

		submission := &models.Submission{
			CoffeeAnswers: []models.CoffeeAnswer{
				{
					CoffeeID: "c1", Region: "R",
					AdditionalAnswers: []models.AdditionalAnswer{
						{QuestionID: "aq1", Answer: "High"},
					},
				},
			},
		}

		score, accuracy := calculateScoreWithCase(submission, testCase)
		// standard: 1/1 correct, base=100*1.0*1=100. additional: 1/1 correct, +25. Total score=125
		// accuracy = 2/2 = 1.0
		if accuracy != 1.0 {
			t.Errorf("expected accuracy=1.0, got %f", accuracy)
		}
		if score != 125 {
			t.Errorf("expected score=125, got %d", score)
		}
	})

	t.Run("wrong additional answer gives no points", func(t *testing.T) {
		coffees := []models.CoffeeItem{
			{
				ID: "c1", Region: "R",
				AdditionalQuestions: []models.AdditionalQuestion{
					{ID: "aq1", Question: "What altitude?", Options: []string{"High", "Low"}, CorrectOption: "High", Points: 25},
				},
			},
		}
		testCase := &models.CoffeeCase{
			ID:      "test",
			Coffees: coffees,
			EnabledQuestions: models.EnabledQuestions{
				Region: true,
			},
		}

		submission := &models.Submission{
			CoffeeAnswers: []models.CoffeeAnswer{
				{
					CoffeeID: "c1", Region: "R",
					AdditionalAnswers: []models.AdditionalAnswer{
						{QuestionID: "aq1", Answer: "Low"},
					},
				},
			},
		}

		score, accuracy := calculateScoreWithCase(submission, testCase)
		// standard: 1/1 correct, base=100. additional: 0/1 correct, +0. Total score=100
		// accuracy = 1/2 = 0.5
		if accuracy != 0.5 {
			t.Errorf("expected accuracy=0.5, got %f", accuracy)
		}
		if score != 100 {
			t.Errorf("expected score=100, got %d", score)
		}
	})

	t.Run("additional answer matching is case-insensitive", func(t *testing.T) {
		coffees := []models.CoffeeItem{
			{
				ID: "c1", Region: "R",
				AdditionalQuestions: []models.AdditionalQuestion{
					{ID: "aq1", Question: "Q?", Options: []string{"A", "B"}, CorrectOption: "High", Points: 10},
				},
			},
		}
		testCase := &models.CoffeeCase{
			ID: "test", Coffees: coffees,
			EnabledQuestions: models.EnabledQuestions{Region: true},
		}

		submission := &models.Submission{
			CoffeeAnswers: []models.CoffeeAnswer{
				{
					CoffeeID: "c1", Region: "R",
					AdditionalAnswers: []models.AdditionalAnswer{
						{QuestionID: "aq1", Answer: "high"},
					},
				},
			},
		}

		score, _ := calculateScoreWithCase(submission, testCase)
		// standard=100 + additional=10 = 110
		if score != 110 {
			t.Errorf("expected score=110, got %d", score)
		}
	})

	t.Run("mixed override + additional questions + fallback", func(t *testing.T) {
		regionOnly := models.EnabledQuestions{Region: true}
		coffees := []models.CoffeeItem{
			{
				ID: "c1", Region: "R1",
				EnabledQuestions: &regionOnly,
				AdditionalQuestions: []models.AdditionalQuestion{
					{ID: "aq1", Question: "Q?", Options: []string{"A", "B"}, CorrectOption: "A", Points: 10},
				},
			},
			{
				ID: "c2", Region: "R2", Variety: "V2",
			},
		}
		testCase := &models.CoffeeCase{
			ID: "test", Coffees: coffees,
			EnabledQuestions: models.EnabledQuestions{Region: true, Variety: true},
		}

		submission := &models.Submission{
			CoffeeAnswers: []models.CoffeeAnswer{
				{
					CoffeeID: "c1", Region: "R1",
					AdditionalAnswers: []models.AdditionalAnswer{
						{QuestionID: "aq1", Answer: "A"},
					},
				},
				{CoffeeID: "c2", Region: "R2", Variety: "V2"},
			},
		}

		score, accuracy := calculateScoreWithCase(submission, testCase)
		// c1: 1 standard question (region), 1 correct + 1 additional correct (+10)
		// c2: 2 standard questions (region+variety), 2 correct
		// totalStandard=3, correctStandard=3, totalAdditional=1, correctAdditional=1
		// standardScore = 100 * (3/3) * 2 = 200
		// additionalScore = 10
		// accuracy = 4/4 = 1.0
		// score = 200 + 10 = 210
		if accuracy != 1.0 {
			t.Errorf("expected accuracy=1.0, got %f", accuracy)
		}
		if score != 210 {
			t.Errorf("expected score=210, got %d", score)
		}
	})

	t.Run("no additional questions backward compat", func(t *testing.T) {
		coffees := []models.CoffeeItem{
			{ID: "c1", Region: "R", Variety: "V", Process: "P", TastingNotes: "chocolate"},
		}
		testCase := makeTestCase(coffees)
		testCase.EnabledQuestions.TasteNote2 = false
		testCase.EnabledQuestions.FavoriteCoffee = false
		testCase.EnabledQuestions.BrewingMethod = false

		submission := &models.Submission{
			CoffeeAnswers: []models.CoffeeAnswer{
				{CoffeeID: "c1", Region: "R", Variety: "V", Process: "P", TasteNote1: "chocolate"},
			},
		}

		score, accuracy := calculateScoreWithCase(submission, testCase)
		if accuracy != 1.0 {
			t.Errorf("expected accuracy=1.0, got %f", accuracy)
		}
		if score != 100 {
			t.Errorf("expected score=100, got %d", score)
		}
	})
}

func TestCountEnabledQuestions(t *testing.T) {
	eq := models.EnabledQuestions{Region: true, Variety: true}
	if c := countEnabledQuestions(eq); c != 2 {
		t.Errorf("expected 2, got %d", c)
	}

	eq = models.EnabledQuestions{}
	if c := countEnabledQuestions(eq); c != 0 {
		t.Errorf("expected 0, got %d", c)
	}
}

func TestGetEffectiveEnabledQuestions(t *testing.T) {
	caseLevel := models.EnabledQuestions{Region: true, Variety: true, Process: true}

	t.Run("nil override returns case level", func(t *testing.T) {
		coffee := &models.CoffeeItem{ID: "c1"}
		eq := getEffectiveEnabledQuestions(coffee, caseLevel)
		if !eq.Region || !eq.Variety || !eq.Process {
			t.Error("expected case-level questions")
		}
	})

	t.Run("non-nil override returns coffee level", func(t *testing.T) {
		override := models.EnabledQuestions{Region: true}
		coffee := &models.CoffeeItem{ID: "c1", EnabledQuestions: &override}
		eq := getEffectiveEnabledQuestions(coffee, caseLevel)
		if !eq.Region {
			t.Error("expected region enabled")
		}
		if eq.Variety || eq.Process {
			t.Error("expected variety and process disabled")
		}
	})
}

func TestUpdateBadges(t *testing.T) {
	t.Run("first case badge", func(t *testing.T) {
		user := &models.User{CasesCount: 1}
		updateBadges(user)
		if !containsBadge(user.Badges, "🔍 Primer Caso") {
			t.Error("expected 'Primer Caso' badge")
		}
	})

	t.Run("no badge for zero cases", func(t *testing.T) {
		user := &models.User{CasesCount: 0}
		updateBadges(user)
		if containsBadge(user.Badges, "🔍 Primer Caso") {
			t.Error("did not expect 'Primer Caso' badge with 0 cases")
		}
	})

	t.Run("accuracy 70% badge", func(t *testing.T) {
		user := &models.User{Accuracy: 0.7}
		updateBadges(user)
		if !containsBadge(user.Badges, "🎯 Precisión 70%") {
			t.Error("expected 'Precisión 70%' badge")
		}
	})

	t.Run("accuracy 80% badge", func(t *testing.T) {
		user := &models.User{Accuracy: 0.8}
		updateBadges(user)
		if !containsBadge(user.Badges, "💎 Catador Nivel 2") {
			t.Error("expected 'Catador Nivel 2' badge")
		}
	})

	t.Run("master detective badge at 2000 points", func(t *testing.T) {
		user := &models.User{Points: 2000}
		updateBadges(user)
		if !containsBadge(user.Badges, "🏆 Detective Maestro") {
			t.Error("expected 'Detective Maestro' badge")
		}
	})

	t.Run("no master badge below 2000 points", func(t *testing.T) {
		user := &models.User{Points: 1999}
		updateBadges(user)
		if containsBadge(user.Badges, "🏆 Detective Maestro") {
			t.Error("did not expect 'Detective Maestro' badge")
		}
	})

	t.Run("expert badge at 5 cases", func(t *testing.T) {
		user := &models.User{CasesCount: 5}
		updateBadges(user)
		if !containsBadge(user.Badges, "🔥 Experto en Tuestes") {
			t.Error("expected 'Experto en Tuestes' badge")
		}
	})

	t.Run("preserves existing badges", func(t *testing.T) {
		user := &models.User{
			CasesCount: 1,
			Badges:     []string{"🔍 Primer Caso"},
		}
		updateBadges(user)
		if !containsBadge(user.Badges, "🔍 Primer Caso") {
			t.Error("expected existing badge to be preserved")
		}
	})

	t.Run("multiple badges at once", func(t *testing.T) {
		user := &models.User{
			CasesCount: 5,
			Points:     2000,
			Accuracy:   0.85,
		}
		updateBadges(user)
		expected := []string{
			"🔍 Primer Caso",
			"🎯 Precisión 70%",
			"💎 Catador Nivel 2",
			"🏆 Detective Maestro",
			"🔥 Experto en Tuestes",
		}
		for _, badge := range expected {
			if !containsBadge(user.Badges, badge) {
				t.Errorf("expected badge %q", badge)
			}
		}
	})
}

func containsBadge(badges []string, target string) bool {
	for _, b := range badges {
		if b == target {
			return true
		}
	}
	return false
}
