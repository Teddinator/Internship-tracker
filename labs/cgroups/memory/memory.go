package main

import (
	"fmt"
	"runtime"
	"time"
)

func main() {
	const chunkSize = 10 * 1024 * 1024

	chunks := make([][]byte, 0, 10)

	for i := 1; i <= 10; i++ {
		chunk := make([]byte, chunkSize)

		// Skriv till varje minnessida så att RAM tas i anspråk.
		for j := 0; j < len(chunk); j += 4096 {
			chunk[j] = 1
		}

		chunks = append(chunks, chunk)

		fmt.Printf("Allocated: %d MiB\n", i*10)
		time.Sleep(1 * time.Second)
	}

	// Behåll allokerade objekt vid liv.
	runtime.KeepAlive(chunks)
	time.Sleep(5 * time.Second)
}
