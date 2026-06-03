package utils

import (
	"testing"
)

func TestGenerateOrderID(t *testing.T) {
	t.Run("returns 6 character string", func(t *testing.T) {
		id := GenerateOrderID()
		if len(id) != 6 {
			t.Errorf("expected length 6, got %d: %q", len(id), id)
		}
	})

	t.Run("returns only uppercase alphanumeric characters", func(t *testing.T) {
		const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		for i := 0; i < 100; i++ {
			id := GenerateOrderID()
			for _, c := range id {
				found := false
				for _, valid := range charset {
					if c == valid {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("invalid character %c in order ID %q", c, id)
				}
			}
		}
	})

	t.Run("generates unique IDs", func(t *testing.T) {
		seen := make(map[string]bool)
		for i := 0; i < 1000; i++ {
			id := GenerateOrderID()
			if seen[id] {
				t.Errorf("duplicate order ID generated: %q", id)
			}
			seen[id] = true
		}
	})
}
