package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func individualInserts(ctx context.Context, db *sql.DB, n int) error {

	for i := 0; i < n; i++ {
		_, err := db.ExecContext(ctx,
			`INSERT INTO io_benchmark (company, status)
			VALUES ($1, $2)`,
			fmt.Sprintf("Company-%d", i),
			"applied",
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func multiRowInserts(ctx context.Context, db *sql.DB, n, batchSize int) error {
	if batchSize < 0 {
		return fmt.Errorf("batchSize must be greater than 0")
	}

	for start := 0; start < n; start += batchSize {
		var query strings.Builder
		args := make([]any, 0, 2*batchSize)

		end := min(start+batchSize, n)
		rows := end - start

		query.WriteString("INSERT INTO io_benchmark (company, status) VALUES ")
		for i := 0; i < rows; i++ {
			if i > 0 {
				query.WriteString(", ")
			}

			fmt.Fprintf(&query, "($%d, $%d)", 2*i+1, 2*i+2)

			args = append(args,
				fmt.Sprintf("company%d", i+1),
				"applied",
			)
		}
		_, err := db.ExecContext(
			ctx,
			query.String(),
			args...,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func transactionBatches(ctx context.Context, db *sql.DB, n, batchSize int) error {
	if batchSize < 0 {
		return fmt.Errorf("batchSize must be greater than 0")
	}

	for start := 0; start < n; start += batchSize {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		end := min(start+batchSize, n)

		for i := start; i < end; i++ {
			_, err = tx.ExecContext(ctx,
				`INSERT INTO io_benchmark (company, status) VALUES 
				($1, $2)`,
				fmt.Sprintf("company%d", i+1),
				"applied",
			)
			if err != nil {
				_ = tx.Rollback()
				return err
			}
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to PostgreSQL!")

	const n = 1000
	const batchSize = 100

	_, err = db.ExecContext(ctx, "TRUNCATE TABLE io_benchmark RESTART IDENTITY")
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()

	if err := individualInserts(ctx, db, n); err != nil {
		log.Fatal(err)
	}

	elapsed := time.Since(start)

	fmt.Printf("Individual INSERTs: %v\n", elapsed)
	fmt.Printf("Throughput: %.0f rows/sec\n", float64(n)/elapsed.Seconds())

	_, err = db.ExecContext(ctx, "TRUNCATE TABLE io_benchmark RESTART IDENTITY")
	if err != nil {
		log.Fatal(err)
	}

	start = time.Now()

	if err := multiRowInserts(ctx, db, n, batchSize); err != nil {
		log.Fatal(err)
	}

	elapsed = time.Since(start)

	fmt.Printf("Multi-row INSERTs: %v\n", elapsed)
	fmt.Printf("Throughput: %.0f rows/sec\n", float64(n)/elapsed.Seconds())

	_, err = db.ExecContext(ctx, "TRUNCATE TABLE io_benchmark RESTART IDENTITY")
	if err != nil {
		log.Fatal(err)
	}

	start = time.Now()

	if err := transactionBatches(ctx, db, n, batchSize); err != nil {
		log.Fatal(err)
	}

	elapsed = time.Since(start)

	fmt.Printf("Transaction batches: %v\n", elapsed)
	fmt.Printf("Throughput: %.0f rows/sec\n", float64(n)/elapsed.Seconds())
}
