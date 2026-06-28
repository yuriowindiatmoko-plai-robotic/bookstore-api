package models

import (
	"encoding/json"
	"testing"
)

func TestBookJSONSerialization(t *testing.T) {
	desc := "A great book about Go programming"
	book := Book{
		ID:          1,
		Title:       "Go Programming",
		Author:      "John Doe",
		ISBN:        "978-3-16-148410-0",
		Price:       29.99,
		Stock:       10,
		Description: &desc,
	}

	data, err := json.Marshal(book)
	if err != nil {
		t.Fatalf("Failed to marshal book: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal book: %v", err)
	}

	if result["title"] != "Go Programming" {
		t.Errorf("Expected title 'Go Programming', got %v", result["title"])
	}

	if result["author"] != "John Doe" {
		t.Errorf("Expected author 'John Doe', got %v", result["author"])
	}

	if result["isbn"] != "978-3-16-148410-0" {
		t.Errorf("Expected isbn '978-3-16-148410-0', got %v", result["isbn"])
	}

	if result["price"] != 29.99 {
		t.Errorf("Expected price 29.99, got %v", result["price"])
	}

	if result["stock"] != float64(10) {
		t.Errorf("Expected stock 10, got %v", result["stock"])
	}

	if result["description"] != "A great book about Go programming" {
		t.Errorf("Expected description 'A great book...', got %v", result["description"])
	}
}

func TestBookJSONWithNilDescription(t *testing.T) {
	book := Book{
		ID:          1,
		Title:       "Go Programming",
		Author:      "John Doe",
		ISBN:        "978-3-16-148410-0",
		Price:       29.99,
		Stock:       10,
		Description: nil,
	}

	data, err := json.Marshal(book)
	if err != nil {
		t.Fatalf("Failed to marshal book: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal book: %v", err)
	}

	// *string without omitempty marshals nil as null
	if result["description"] != nil {
		t.Errorf("Description should be null when nil pointer, got: %v", result["description"])
	}
}
