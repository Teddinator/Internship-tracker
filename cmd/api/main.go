package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

type application struct {
	ID        int64  `json:"id"`
	Company   string `json:"company"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	AppliedAt any    `json:"applied_at"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type applicationServer struct {
	db *sql.DB
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	app := &applicationServer{
		db: db,
	}

	r := chi.NewRouter()

	r.Get("/health", app.healthHandler)
	r.Get("/applications", app.getApplicationsHandler)

	log.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func (app *applicationServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func (app *applicationServer) getApplicationsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := app.db.Query(`
		SELECT id, company, role, status, applied_at, created_at, updated_at
		FROM applications
		ORDER BY id
	`)
	if err != nil {
		http.Error(w, "failed to query applications", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var applications []application

	for rows.Next() {
		var a application

		err := rows.Scan(
			&a.ID,
			&a.Company,
			&a.Role,
			&a.Status,
			&a.AppliedAt,
			&a.CreatedAt,
			&a.UpdatedAt,
		)
		if err != nil {
			http.Error(w, "failed to scan application", http.StatusInternalServerError)
			return
		}

		applications = append(applications, a)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "failed to read applications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(applications)
}
