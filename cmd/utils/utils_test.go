package utils

import (
	"fmt"
	"math"
	"testing"
)

func BenchmarkGenerateID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateID()
	}
}

func BenchmarkGenerateRandom(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateRandom(int(math.Min(float64(b.N), 128)))
	}
}

func BenchmarkEncodeUserID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		EncodeUserID(fmt.Sprintf("User_%d", b.N))
	}
}
