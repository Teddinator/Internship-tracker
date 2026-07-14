package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"

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

	router := chi.NewRouter()

	router.Route("/health", func(r chi.Router) {
		r.Get("/", app.healthHandler)
	})

	router.Route("/applications", func(r chi.Router) {
		r.Get("/", app.getApplicationsHandler)
		r.Get("/{id}", app.getApplicationsByIDHandler)
	})

	log.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
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

func (app *applicationServer) getApplicationsByIDHandler(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		http.Error(w, "Invalid application id", http.StatusBadRequest)
		return
	}

	var a application

	err = app.db.QueryRowContext(
		r.Context(),
		`
			SELECT id, company, role, status, applied_at, created_at, updated_at
			FROM applications
			WHERE id = $1
		`, id,
	).Scan(
		&a.ID,
		&a.Company,
		&a.Role,
		&a.Status,
		&a.AppliedAt,
		&a.CreatedAt,
		&a.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "failed to query application", http.StatusNotFound)
		return
	}

	if err != nil {
		log.Printf("failed to query application: %v", err)
		http.Error(
			w,
			"failed to query application",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(a); err != nil {
		log.Printf("failed to encode application :%v", err)
	}
}
