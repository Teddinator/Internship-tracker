package main

import (
	"fmt"
	"time"
)

func main() {
	start := time.Now()

	var result uint64

	for i := uint64(0); i < 1_000_000_000; i++ {
		result += i*i + 1
	}

	fmt.Printf("Result: %d\n", result)
	fmt.Printf("Elapsed: %s\n", time.Since(start))
}
