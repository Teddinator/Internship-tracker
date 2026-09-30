package main

import (
	"sync"
)

func work(iterations int) {
	sum := 0

	for i := 0; i < iterations; i++ {
		sum = sum + i
	}
	_ = sum
}

func main() {
	var wg sync.WaitGroup

	totalWork := 100_000_000_000
	workers := 100
	workPerWorker := totalWork / workers

	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			work(workPerWorker)
		}()
	}

	wg.Wait()
}
