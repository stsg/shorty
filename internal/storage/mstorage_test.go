package storage

import "testing"

func TestGetAllURLs_EmptyList(t *testing.T) {
	// Initialize MapStorage object
	mStorage, err := NewMapStorage()

	// Check if there is no error
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Invoke GetAllURLs method
	result, err := mStorage.GetAllURLs(123, "http://example.com")

	// Check if the result is an empty list
	if len(result) != 0 {
		t.Errorf("Expected an empty list, but got %v", result)
	}

	// Check if there is no error
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
