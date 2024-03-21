package storage

import (
	"testing"
)

func BenchmarkGenShortURL(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenShortURL()
	}
}
