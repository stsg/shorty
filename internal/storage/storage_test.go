package storage

import (
	"testing"
)

// BenchmarkGenShortURL benchmarks the GenShortURL function.
//
// Parameter b is a testing.B type.
// No return value.
func BenchmarkGenShortURL(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenShortURL()
	}
}

func TestGenShortURLWithValidLength(t *testing.T) {
	// Set the ShortURLLength to a valid value
	ShortURLLength = 6

	// Call the GenShortURL function
	result := GenShortURL()

	// Assert that the length of the generated short URL is equal to the specified ShortURLLength
	if len(result) != ShortURLLength {
		t.Errorf("Expected length of generated short URL to be %d, but got %d", ShortURLLength, len(result))
	}
}

// ShortURLLength is 0.
func TestGenShortURLWithZeroLength(t *testing.T) {
	// Set the ShortURLLength to 0
	ShortURLLength = 0

	// Call the GenShortURL function
	result := GenShortURL()

	// Assert that the length of the generated short URL is 0
	if len(result) != ShortURLLength {
		t.Errorf("Expected length of generated short URL to be %d, but got %d", ShortURLLength, len(result))
	}
}
