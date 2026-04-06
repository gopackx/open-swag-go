package examples

import (
	"testing"
)

func TestGenerator(t *testing.T) {
	type User struct {
		ID    string `json:"id" format:"uuid"`
		Name  string `json:"name" example:"John Doe"`
		Email string `json:"email"`
		Age   int    `json:"age"`
	}

	gen := New(Config{})
	result := gen.Generate(User{})

	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("expected map result")
	}

	// Check explicit example
	if m["name"] != "John Doe" {
		t.Errorf("expected name 'John Doe', got '%v'", m["name"])
	}

	// Check format-based example
	if m["id"] != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected uuid format for id, got '%v'", m["id"])
	}

	// Check email heuristic
	if m["email"] != "user@example.com" {
		t.Errorf("expected email heuristic, got '%v'", m["email"])
	}
}

func TestGeneratorFieldNameHeuristics(t *testing.T) {
	type Request struct {
		UserID      string  `json:"user_id"`
		PhoneNumber string  `json:"phone_number"`
		CreatedAt   string  `json:"created_at"`
		TotalCount  int     `json:"total_count"`
		Price       float64 `json:"price"`
	}

	gen := New(Config{})
	result := gen.GenerateJSON(Request{})

	if result["user_id"] != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected uuid for user_id, got '%v'", result["user_id"])
	}

	if result["phone_number"] != "+1-555-123-4567" {
		t.Errorf("expected phone for phone_number, got '%v'", result["phone_number"])
	}
}

func TestGeneratorSlice(t *testing.T) {
	type Item struct {
		Name string `json:"name"`
	}

	type Response struct {
		Items []Item `json:"items"`
	}

	gen := New(Config{})
	result := gen.GenerateJSON(Response{})

	items, ok := result["items"].([]interface{})
	if !ok {
		t.Fatal("expected items to be slice")
	}

	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}
