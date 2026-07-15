package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

type Application struct {
	ID        int64      `json:"id"`
	CompanyID string     `json:"company_id"`
	Company   string     `json:"company"`
	Role      string     `json:"role"`
	Status    string     `json:"status"`
	AppliedAt *time.Time `json:"applied_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type Company struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Website   string    `json:"website"`
	Industry  string    `json:"industry"`
	Location  string    `json:"location"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
	rows, err := app.db.QueryContext(r.Context(), `
		SELECT
			a.id,
			a.company_id,
			c.name,
			a.role,
			a.status,
			a.applied_at,
			a.created_at,
			a.updated_at 
		FROM applications AS a
		JOIN companies AS c
			ON c.id = a.company_id
		ORDER BY id
	`)
	if err != nil {
		log.Printf("failed to query applications: %v", err)
		http.Error(w, "failed to query applications", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	applications := make([]Application, 0)

	for rows.Next() {
		var a Application

		err := rows.Scan(
			&a.ID,
			&a.CompanyID,
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
	if err := json.NewEncoder(w).Encode(applications); err != nil {
		log.Printf("failed to encode applications: %v", err)
	}
}

func (app *applicationServer) getApplicationsByIDHandler(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		http.Error(w, "Invalid application id", http.StatusBadRequest)
		return
	}

	var a Application

	err = app.db.QueryRowContext(
		r.Context(),
		`
			SELECT 
				a.id,
				a.company_id,
				c.name,
				a.role,
				a.status,
				a.applied_at,
				a.created_at,
				a.updated_at
			FROM applications AS a
			JOIN companies AS c
				ON c.id = a.company_id
			WHERE a.id = $1
		`, id,
	).Scan(
		&a.ID,
		&a.CompanyID,
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
