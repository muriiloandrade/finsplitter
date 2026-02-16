package entity

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
)

func TestCardBrandJSONSerialization(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	id := uuid.Must(uuid.NewV4())

	brand := CardBrand{
		ID:               id,
		Name:             "Visa",
		CreatedDate:      now,
		LastModifiedDate: now,
	}

	// Test MarshalJSON
	data, err := json.Marshal(brand)
	if err != nil {
		t.Fatalf("failed to marshal CardBrand: %v", err)
	}

	// Test UnmarshalJSON
	var decoded CardBrand
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("failed to unmarshal CardBrand: %v", err)
	}

	if decoded.ID != brand.ID {
		t.Errorf("ID mismatch: got %v, want %v", decoded.ID, brand.ID)
	}
	if decoded.Name != brand.Name {
		t.Errorf("Name mismatch: got %v, want %v", decoded.Name, brand.Name)
	}
	if !decoded.CreatedDate.Equal(brand.CreatedDate) {
		t.Errorf("CreatedDate mismatch: got %v, want %v", decoded.CreatedDate, brand.CreatedDate)
	}
	if !decoded.LastModifiedDate.Equal(brand.LastModifiedDate) {
		t.Errorf("LastModifiedDate mismatch: got %v, want %v", decoded.LastModifiedDate, brand.LastModifiedDate)
	}
}

func TestCardBrandJSONTags(t *testing.T) {
	brand := CardBrand{
		ID:               uuid.Must(uuid.NewV4()),
		Name:             "Mastercard",
		CreatedDate:      time.Now(),
		LastModifiedDate: time.Now(),
	}

	data, err := json.Marshal(brand)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Verify JSON contains expected keys (lowercase due to json tags)
	str := string(data)
	if len(str) == 0 {
		t.Error("expected non-empty JSON")
	}

	// Check that fields are serialized with lowercase keys
	if !containsSubstring(str, `"id"`) || !containsSubstring(str, `"name"`) ||
		!containsSubstring(str, `"createdDate"`) || !containsSubstring(str, `"lastModifiedDate"`) {
		t.Error("JSON should contain all fields with correct json tags")
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && len(s) >= len(substr) &&
		(s == substr || len(s) > 0 && (s[:len(substr)] == substr || containsAny(s, substr)))
}

func containsAny(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestCardBrandEmptyStruct(t *testing.T) {
	brand := CardBrand{}

	if brand.ID != uuid.Nil {
		t.Error("expected empty ID to be Nil")
	}
	if brand.Name != "" {
		t.Error("expected empty Name")
	}
	if !brand.CreatedDate.IsZero() {
		t.Error("expected zero time for CreatedDate")
	}
	if !brand.LastModifiedDate.IsZero() {
		t.Error("expected zero time for LastModifiedDate")
	}
}
