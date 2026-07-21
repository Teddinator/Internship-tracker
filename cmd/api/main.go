package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
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
	Website   *string   `json:"website"`
	Industry  *string   `json:"industry"`
	Location  *string   `json:"location"`
	Notes     *string   `json:"notes"`
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
		r.Post("/", app.createApplicationHandler)
		r.Get("/", app.getApplicationsHandler)
		r.Get("/{id}", app.getApplicationsByIDHandler)
		r.Put("/{id}", app.updateApplicationHandler)
		r.Delete("/{id}", app.deleteApplicationHandler)
	})

	router.Route("/companies", func(r chi.Router) {
		r.Get("/", app.getCompaniesHandler)
		r.Get("/{id}", app.getCompaniesByIDHandler)
		r.Post("/", app.createCompanyHandler)
		r.Put("/{id}", app.updateCompanyHandler)
		r.Delete("/{id}", app.deleteCompanyHandler)
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

func (app *applicationServer) createApplicationHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		CompanyID string  `json:"company_id"`
		Role      string  `json:"role"`
		Status    string  `json:"status"`
		AppliedAt *string `json:"applied_at"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		log.Printf("Failed to decode application request :%v", err)
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	input.CompanyID = strings.TrimSpace(input.CompanyID)
	input.Role = strings.TrimSpace(input.Role)
	input.Status = strings.TrimSpace(input.Status)

	if input.CompanyID == "" {
		http.Error(w, "company_id is required", http.StatusBadRequest)
		return
	}

	companyID, err := uuid.Parse(input.CompanyID)
	if err != nil {
		http.Error(w, "company_id must be a valid UUID", http.StatusBadRequest)
		return
	}

	if input.Role == "" {
		http.Error(w, "role is required", http.StatusBadRequest)
		return
	}

	if input.Status == "" {
		input.Status = "applied"
	}

	allowedStatuses := map[string]bool{
		"applied":   true,
		"interview": true,
		"offer":     true,
		"rejected":  true,
		"withdrawn": true,
	}

	if !allowedStatuses[input.Status] {
		http.Error(w, "Invalid application status", http.StatusBadRequest)
		return
	}

	var appliedAt *time.Time

	if input.AppliedAt != nil {
		value := strings.TrimSpace(*input.AppliedAt)

		if value != "" {
			parsed, err := time.Parse("2006-01-02", value)
			if err != nil {
				http.Error(w, "applied_at must use YYY-MM-DD format", http.StatusBadRequest)
				return
			}

			appliedAt = &parsed
		}
	}

	var a Application

	err = app.db.QueryRowContext(r.Context(),
		`
			INSERT INTO applications (
				company_id,
				role,
				status,
				applied_at
			)
			VALUES ($1, $2, $3, $4)
			RETURNING
				id,
				company_id,
				role,
				status,
				applied_at,
				created_at,
				updated_at
		`,
		companyID,
		input.Role,
		input.Status,
		appliedAt,
	).Scan(
		&a.ID,
		&a.CompanyID,
		&a.Role,
		&a.Status,
		&a.AppliedAt,
		&a.CreatedAt,
		&a.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		// company_id points at a company that doesn't exist
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			http.Error(w, "Company not found", http.StatusNotFound)
			return
		}

		log.Printf("Failed to create application : %v", err)
		http.Error(w, "Failed to create application", http.StatusInternalServerError)
		return
	}

	// Company field exist in JSON-model but not in application table. Get company name seperately.
	err = app.db.QueryRowContext(
		r.Context(),
		`SELECT name FROM companies WHERE id = $1`,
		companyID,
	).Scan(&a.Company)

	if err != nil {
		log.Printf("failed to retrieve company name: %v", err)
		http.Error(w, "Application was not created but response could not be built", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set(
		"Location",
		fmt.Sprintf("/applications/%d", a.ID),
	)
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(a); err != nil {
		log.Printf("failed to encode application response: %v", err)
	}
}

func (app *applicationServer) updateApplicationHandler(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idString, 10, 64)

	if err != nil {
		http.Error(w, "Invalid application ID", http.StatusBadRequest)
		return
	}

	var input struct {
		CompanyID string  `json:"company_id"`
		Role      string  `json:"role"`
		Status    string  `json:"status"`
		AppliedAt *string `json:"applied_at"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	input.CompanyID = strings.TrimSpace(input.CompanyID)
	input.Role = strings.TrimSpace(input.Role)
	input.Status = strings.TrimSpace(input.Status)

	if input.CompanyID == "" {
		http.Error(w, "company_id is required", http.StatusBadRequest)
		return
	}

	CompanyID, err := uuid.Parse(input.CompanyID)
	if err != nil {
		http.Error(w, "company_id must be a valid UUID", http.StatusBadRequest)
		return
	}

	if input.Role == "" {
		http.Error(w, "role is required", http.StatusBadRequest)
		return
	}

	allowedStatuses := map[string]bool{
		"applied":   true,
		"interview": true,
		"offer":     true,
		"rejected":  true,
		"withdrawn": true,
	}

	if !allowedStatuses[input.Status] {
		http.Error(w, "Invalid application status", http.StatusBadRequest)
		return
	}

	var appliedAt *time.Time

	if input.AppliedAt != nil {
		value := strings.TrimSpace(*input.AppliedAt)

		if value != "" {
			parsed, err := time.Parse("2006-01-02", value)
			if err != nil {
				http.Error(w, "applied_at must use YYY-MM-DD format", http.StatusBadRequest)
				return
			}

			appliedAt = &parsed
		}
	}

	var a Application

	err = app.db.QueryRowContext(
		r.Context(),
		`
			UPDATE applications
			SET
				company_id = $1,
				role = $2,
				status = $3,
				applied_at = $4,
				updated_at = NOW()
			WHERE id = $5
			RETURNING
				id,
				company_id,
				role,
				status,
				applied_at,
				updated_at,
				created_at
		`,
		CompanyID,
		input.Role,
		input.Status,
		appliedAt,
		id,
	).Scan(
		&a.ID,
		&a.CompanyID,
		&a.Role,
		&a.Status,
		&a.AppliedAt,
		&a.UpdatedAt,
		&a.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Application not found", http.StatusNotFound)
		return
	}

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			http.Error(w, "Company not found", http.StatusNotFound)
			return
		}

		log.Printf("Failed to create application %d, %v", id, err)
		http.Error(w, "Failed to update application", http.StatusInternalServerError)
		return
	}

	err = app.db.QueryRowContext(
		r.Context(),
		`SELECT name FROM companies WHERE id = $1`,
		CompanyID,
	).Scan(&a.Company)

	if err != nil {
		log.Printf("Failed to retrieve company name: %v", err)
		http.Error(w, "Failed to retrieve company", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(a); err != nil {
		log.Printf("Failed to encode application: %v", err)
		return
	}
}

func (app *applicationServer) deleteApplicationHandler(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idString, 10, 64)

	if err != nil {
		http.Error(w, "Invalid application id", http.StatusBadRequest)
		return
	}

	result, err := app.db.ExecContext(
		r.Context(),
		`	
			DELETE FROM applications
			WHERE id = $1
		`, id,
	)

	if err != nil {
		http.Error(w, "Failed to delete application", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		http.Error(w, "failed to check deleted application", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Application not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *applicationServer) getCompaniesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := app.db.QueryContext(r.Context(),
		`
			SELECT 
			c.id,
			c.name,
			c.website,
			c.industry,
			c.location,
			c.notes,
			c.created_at,
			c.updated_at
			FROM companies AS c
			ORDER BY id
		`,
	)

	if err != nil {
		http.Error(w, "Couldn't reterive companies", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	companies := make([]Company, 0)

	for rows.Next() {
		var c Company

		err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.Website,
			&c.Industry,
			&c.Location,
			&c.Notes,
			&c.CreatedAt,
			&c.UpdatedAt,
		)

		if err != nil {
			log.Printf("Failed to convert sql to json: %v", err)
			http.Error(w, "Couldnt convert database entry to JSON", http.StatusInternalServerError)
			return
		}

		companies = append(companies, c)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "failed to read applications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(companies); err != nil {
		log.Printf("Couldn't encode data: %v", err)
		http.Error(w, "Couldn't encode data", http.StatusInternalServerError)
		return
	}
}

func (app *applicationServer) getCompaniesByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if _, err := uuid.Parse(id); err != nil {
		log.Printf("Invalid company UUID %q: %v", id, err)
		http.Error(w, "Invalid company ID", http.StatusInternalServerError)
		return
	}

	var c Company

	err := app.db.QueryRowContext(r.Context(),
		`
			SELECT
			c.id,
			c.name,
			c.website,
			c.industry,
			c.location,
			c.notes,
			c.created_at,
			c.updated_at
			FROM companies AS c
			WHERE ID = $1
		`, id,
	).Scan(
		&c.ID,
		&c.Name,
		&c.Website,
		&c.Industry,
		&c.Location,
		&c.Notes,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		log.Printf("No valid entry was found")
		http.Error(w, "Could not find entry matching ID", http.StatusInternalServerError)
		return
	}

	if err != nil {
		log.Printf("Failed to query applications: %v", err)
		http.Error(w, "Failed to query applications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(c); err != nil {
		log.Printf("Couldnt convert data to JSON: %v", err)
		http.Error(w, "Couldn't convert data to JSON", http.StatusInternalServerError)
	}
}

func (app *applicationServer) createCompanyHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Website  string `json:"website"`
		Industry string `json:"industry"`
		Location string `json:"location"`
		Notes    string `json:"notes"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		log.Printf("Failed to decode company request: %v", err)
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Website = strings.TrimSpace(input.Website)
	input.Industry = strings.TrimSpace(input.Industry)
	input.Location = strings.TrimSpace(input.Location)
	input.Notes = strings.TrimSpace(input.Notes)

	if input.Name == "" {
		http.Error(w, "Company name is required", http.StatusBadRequest)
		return
	}

	var c Company

	err = app.db.QueryRowContext(
		r.Context(),
		`INSERT INTO companies (
			name,
			website,
			industry,
			location,
			notes
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			name,
			website,
			industry,
			location,
			notes,
			created_at,
			updated_at
		`,
		input.Name,
		input.Website,
		input.Industry,
		input.Location,
		input.Notes,
	).Scan(
		&c.ID,
		&c.Name,
		&c.Website,
		&c.Industry,
		&c.Location,
		&c.Notes,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if err != nil {
		log.Printf("Failed to create company: %v", err)
		http.Error(w, "Failed to create company", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set(
		"Location",
		fmt.Sprintf("/companies/%s", c.ID),
	)
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(c); err != nil {
		log.Printf("Failed to encode company response: %v", err)
		http.Error(w, "Failed to encode company respone", http.StatusInternalServerError)
		return
	}
}

func (app *applicationServer) updateCompanyHandler(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	id, err := uuid.Parse(idString)
	if err != nil {
		log.Printf("Invalid company UUID %q: %v", idString, err)
		http.Error(w, "Invalid company ID", http.StatusBadRequest)
		return
	}

	var input struct {
		Name     string `json:"name"`
		Website  string `json:"website"`
		Industry string `json:"industry"`
		Location string `json:"location"`
		Notes    string `json:"notes"`
	}

	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		log.Printf("Failed to decode company update: %v", err)
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Website = strings.TrimSpace(input.Website)
	input.Industry = strings.TrimSpace(input.Industry)
	input.Location = strings.TrimSpace(input.Location)
	input.Notes = strings.TrimSpace(input.Notes)

	if input.Name == "" {
		http.Error(w, "Company name is required", http.StatusBadRequest)
		return
	}

	var c Company

	err = app.db.QueryRowContext(
		r.Context(),
		`
			UPDATE companies
			SET
				name = $1,
				website = $2,
				industry = $3,
				location = $4,
				notes = $5,
				updated_at = NOW()
			WHERE id = $6
			RETURNING
				id,
				name,
				website,
				industry,
				location,
				notes,
				created_at,
				updated_at
		`,
		input.Name,
		input.Website,
		input.Industry,
		input.Location,
		input.Notes,
		id,
	).Scan(
		&c.ID,
		&c.Name,
		&c.Website,
		&c.Industry,
		&c.Location,
		&c.Notes,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		http.Error(w, "Company not found", http.StatusNotFound)
		return

	case err != nil:
		log.Printf("Failed to update company %s: %v", id, err)
		http.Error(w, "Falied to update company", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(c); err != nil {
		log.Printf("Failed to encode updated company: %v", err)
	}
}

func (app *applicationServer) deleteCompanyHandler(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	id, err := uuid.Parse(idString)

	if err != nil {
		log.Printf("Invalid company UUID: %q: %v", idString, err)
		http.Error(w, "Invalid company ID", http.StatusBadRequest)
		return
	}

	result, err := app.db.ExecContext(r.Context(),
		`
			DELETE FROM companies
			WHERE id = $1
		`,
		id,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			http.Error(
				w,
				"Company cannot be deleted because it has applications",
				http.StatusConflict,
			)
			return
		}

		log.Printf("Failed to delete company %s: %v", id, err)
		http.Error(w, "Failed to delete company", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		log.Printf("Failed to read affected rows: %v", err)
		http.Error(w, "Failed to read affected rows", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Company not found", http.StatusNotFound)
	}

	w.WriteHeader(http.StatusNoContent)
}
