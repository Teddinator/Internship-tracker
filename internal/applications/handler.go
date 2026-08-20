package applications

import (
	"context"
	"database/sql"
	"encoding/csv"
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

	"github.com/teddinator/Internship-tracker/internal/apierror"
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
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	companyID := strings.TrimSpace(r.URL.Query().Get("company_id"))

	query := `
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
		WHERE 1 = 1
	`

	args := make([]any, 0)

	if status != "" {
		if !isValidStatus(status) {
			apierror.BadRequest(
				w,
				"invalid_application_id",
				"application id must be a valid integer",
			)
			return
		}
		args = append(args, status)
		query += fmt.Sprintf(" AND a.status = $%d", len(args))
	}

	if companyID != "" {
		parsedCompanyID, err := uuid.Parse(companyID)
		if err != nil {
			apierror.Write(
				w,
				http.StatusBadRequest,
				"invalid_company_id",
				"company id must be a valid UUID",
			)
			return
		}

		args = append(args, parsedCompanyID)
		query += fmt.Sprintf("AND a.company_id = $%d", len(args))
	}

	query += " ORDER BY a.id"

	rows, err := h.db.QueryContext(
		r.Context(),
		query,
		args...,
	)

	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			log.Printf("Request canceled while querying applications: %v", err)
			return

		case errors.Is(err, context.DeadlineExceeded):
			log.Printf("database query timed out :%v", err)
			apierror.Write(
				w,
				http.StatusGatewayTimeout,
				"query_timeout",
				"the request timed out",
			)
			return

		default:
			log.Printf("Failed to query application: %v", err)
			apierror.Internal(w)
			return
		}

	}

	defer rows.Close()

	applications := make([]Application, 0)

	for rows.Next() {
		var a Application

		if err := rows.Scan(
			&a.ID,
			&a.CompanyID,
			&a.Company,
			&a.Role,
			&a.Status,
			&a.AppliedAt,
			&a.CreatedAt,
			&a.UpdatedAt,
		); err != nil {
			log.Printf("scan application row: %v", err)
			apierror.Internal(w)
			return
		}

		applications = append(applications, a)
	}

	if err := rows.Err(); err != nil {
		log.Printf("iterate application rows: %v", err)
		apierror.Internal(w)
		return
	}

	w.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(w).Encode(applications); err != nil {
		log.Printf("encode applications response: %v", err)
	}
}

func (h *Handler) GetAppByID(w http.ResponseWriter, r *http.Request) {

	idString := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		apierror.BadRequest(
			w,
			"Invalid application id",
			"Application id must be an integer",
		)
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
		apierror.NotFound(
			w,
			"Application_not_found",
			"Application not found",
		)
		return
	}

	if err != nil {
		log.Printf("query application: %d: %v", id, err)
		apierror.Internal(w)
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
		log.Printf("Decode application request :%v", err)
		apierror.BadRequest(
			w,
			"invalid_json",
			"request body contains invalid json",
		)
		return
	}

	input.CompanyID = strings.TrimSpace(input.CompanyID)
	input.Role = strings.TrimSpace(input.Role)
	input.Status = strings.TrimSpace(input.Status)

	if input.CompanyID == "" {
		apierror.BadRequest(
			w,
			"missing_company_id",
			"company_id is required",
		)
		return
	}

	companyID, err := uuid.Parse(input.CompanyID)
	if err != nil {
		apierror.BadRequest(
			w,
			"invalid_company_id",
			"company_id must be a valid UUID",
		)
		return
	}

	if input.Role == "" {
		apierror.BadRequest(
			w,
			"missing_role",
			"role is required",
		)
		return
	}

	if input.Status == "" {
		input.Status = "applied"
	}

	if !isValidStatus(input.Status) {
		apierror.BadRequest(
			w,
			"invalid_status",
			"status must be one of: applied, interview, offer, rejected, withdrawn",
		)
		return
	}

	var appliedAt *time.Time

	if input.AppliedAt != nil {
		value := strings.TrimSpace(*input.AppliedAt)

		if value != "" {
			parsed, err := time.Parse("2006-01-02", value)
			if err != nil {
				apierror.BadRequest(
					w,
					"invalid_applied_at",
					"applied_at must use YYYY-MM-DD format",
				)
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
			apierror.NotFound(
				w,
				"company_not_found",
				"company_not_found",
			)
			return
		}

		log.Printf("Create application : %v", err)
		apierror.Internal(w)
		return
	}

	// TODO: Avoid partial success here.
	// The application may already be created if this query fails.
	// Consider using a transaction or combining the queries
	err = h.db.QueryRowContext(
		r.Context(),
		`SELECT name FROM companies WHERE id = $1`,
		companyID,
	).Scan(&a.Company)

	if err != nil {
		log.Printf("retrieve company name after creating application: %v", err)
		apierror.Internal(w)
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
		apierror.BadRequest(
			w,
			"invalid_application_id",
			"application id must be a valid integer",
		)
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
		apierror.BadRequest(
			w,
			"invalid_json",
			"request body contains invalid json",
		)
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

	if !isValidStatus(input.Status) {
		http.Error(w, "Invalid application status", http.StatusBadRequest)
		return
	}

	var appliedAt *time.Time

	if input.AppliedAt != nil {
		value := strings.TrimSpace(*input.AppliedAt)

		if value != "" {
			parsed, err := time.Parse("2006-01-02", value)
			if err != nil {
				http.Error(w, "applied_at must use YYYY-MM-DD format", http.StatusBadRequest)
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

func (h *Handler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(
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
			ORDER BY a.id	
		`,
	)
	if err != nil {
		log.Printf("Failed to query application for CSV export: %v", err)
		http.Error(w, "failed to export applications", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	applications := make([]Application, 0)

	for rows.Next() {
		var a Application

		if err := rows.Scan(
			&a.ID,
			&a.CompanyID,
			&a.Company,
			&a.Role,
			&a.Status,
			&a.AppliedAt,
			&a.CreatedAt,
			&a.UpdatedAt,
		); err != nil {
			log.Printf("Failed to scan application for CSV export: %v", err)
			http.Error(w, "Failed to export applications", http.StatusInternalServerError)
			return
		}
		applications = append(applications, a)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Failed while reading applications for CSV export: %v", err)
		http.Error(w, "Failed to export applications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "json/application")
	w.Header().Set(
		"Content-Disposition",
		`attachment; filename="applications.csv"`,
	)

	writer := csv.NewWriter(w)

	if err := writer.Write([]string{
		"id",
		"company_id",
		"company",
		"role",
		"status",
		"applied_at",
		"created_at",
		"updated_at",
	}); err != nil {
		log.Printf("Failed to write CSV header: %v", err)
		return
	}

	for _, application := range applications {
		appliedAt := ""

		if application.AppliedAt != nil {
			appliedAt = application.AppliedAt.Format("2006-01-02")
		}

		record := []string{
			strconv.FormatInt(application.ID, 10),
			application.CompanyID,
			application.Company,
			application.Role,
			application.Status,
			appliedAt,
			application.CreatedAt.Format(time.RFC3339),
			application.UpdatedAt.Format(time.RFC3339),
		}

		if err := writer.Write(record); err != nil {
			log.Printf("Failed to write CSV record: %v", err)
			return
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		log.Printf("Failed to complete CSV export: %v", err)
	}
}
