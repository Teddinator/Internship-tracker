package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
)

func waitForEnter() {
	fmt.Print("Tryck enter för att fortsätta...")
	bufio.NewReader(os.Stdin).ReadString('\n')
}

func saveHeapProfile() error {
	runtime.GC()

	f, err := os.Create("heap.pprof")
	if err != nil {
		return err
	}
	defer f.Close()

	return pprof.WriteHeapProfile(f)
}

func main() {
	fmt.Println("PID:", os.Getpid())

	waitForEnter()

	//Allokera 200 MiB
	size := 200 * 1024 * 1024
	data := make([]byte, size)

	fmt.Println("200 MiB allokerat")
	waitForEnter()

	// Skriv till varje minnessida (antar 4 KiB)
	pageSize := 4096

	for i := 0; i < len(data); i += pageSize {
		data[i] = 1
	}

	fmt.Println("Minnet har skrivits")
	waitForEnter()

	if err := saveHeapProfile(); err != nil {
		fmt.Println("Profiliering misslyckades")
		return
	}

	runtime.KeepAlive(data)
	fmt.Println("Heap profile sparad")
}
