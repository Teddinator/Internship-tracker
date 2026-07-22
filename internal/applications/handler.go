package applications

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type Handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{
		db: db,
	}
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(), `
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

func (h *Handler) GetAppByID(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		http.Error(w, "Invalid application id", http.StatusBadRequest)
		return
	}

	var a Application

	err = h.db.QueryRowContext(
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

func (h *Handler) CreateApp(w http.ResponseWriter, r *http.Request) {

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

	err = h.db.QueryRowContext(r.Context(),
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
	err = h.db.QueryRowContext(
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

func (h *Handler) UpdateApp(w http.ResponseWriter, r *http.Request) {
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

	err = h.db.QueryRowContext(
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

	err = h.db.QueryRowContext(
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

func (h *Handler) DeleteApp(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idString, 10, 64)

	if err != nil {
		http.Error(w, "Invalid application id", http.StatusBadRequest)
		return
	}

	result, err := h.db.ExecContext(
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
