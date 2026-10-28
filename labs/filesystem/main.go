package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"time"
)

func writeBuffered(f *os.File, b []byte) error {
	for i := 0; i < 1000; i++ {
		_, err := f.Write(b)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeSynced(f *os.File, b []byte) error {
	for i := 0; i < 1000; i++ {
		_, err := f.Write(b)

		if err != nil {
			return err
		}

		err = f.Sync()

		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	data := make([]byte, 4096)

	if _, err := rand.Read(data); err != nil {
		log.Fatal(err)
	}

	//Experiment A: Buffered
	f1, err := os.Create("buffered.dat")
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()

	err = writeBuffered(f1, data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Buffered: ", time.Since(start))

	if err := f1.Close(); err != nil {
		log.Fatal(err)
	}

	//Experminet B: Synced
	f2, err := os.Create("synced.dat")
	if err != nil {
		log.Fatal(err)
	}

	start = time.Now()

	err = writeSynced(f2, data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Synced: ", time.Since(start))

	if err := f2.Close(); err != nil {
		log.Fatal(err)
	}

}
