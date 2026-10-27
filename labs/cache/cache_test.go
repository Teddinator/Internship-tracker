package cache

import (
	"math/rand"
	"testing"
)

var result int64

func BenchmarkSequential(b *testing.B) {
	data := make([]int64, 10_000_000)

	for i := range data {
		data[i] = int64(i)
	}

	b.ResetTimer()

	var sum int64

	for i := 0; i < b.N; i++ {
		sum = sequentialSum(data)
	}

	result = sum

}

func BenchmarkRandom(b *testing.B) {
	data := make([]int64, 10_000_000)

	for i := range data {
		data[i] = int64(i)
	}

	indexes := rand.New(
		rand.NewSource(42),
	).Perm(len(data))

	b.ResetTimer()

	var sum int64

	for i := 0; i < b.N; i++ {
		sum = randomSum(data, indexes)
	}

	result = sum
}

func TestSum(t *testing.T) {
	data := []int64{10, 20, 30, 40, 50}

	indexes := []int{3, 0, 4, 1, 2}

	sequential := sequentialSum(data)
	random := randomSum(data, indexes)

	expected := int64(150)

	if sequential != expected {
		t.Errorf("sequentialSum: got %d, want %d", sequential, expected)
	}

	if random != expected {
		t.Errorf("randomSum: got %d, want %d", random, expected)
	}
}
