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
