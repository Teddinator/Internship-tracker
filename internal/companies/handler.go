package companies

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

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

// TODO: Fixa error handling
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(),
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

	if apierror.HandleContextError(w, err) {
		return
	}

	if err != nil {
		log.Printf("failed to fetch companies: %v", err)
		apierror.Internal(w)
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
			if apierror.HandleContextError(w, err) {
				return
			}

			log.Printf("failed to scan company row: %v", err)
			apierror.Internal(w)
			return
		}

		companies = append(companies, c)
	}

	if err := rows.Err(); err != nil {
		if apierror.HandleContextError(w, err) {
			return
		}

		log.Printf("failed while iterating companies: %v", err)
		apierror.Internal(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(companies); err != nil {
		log.Printf("couldn't encode data: %v", err)
		return
	}
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if _, err := uuid.Parse(id); err != nil {
		log.Printf("invalid company UUID %q: %v", id, err)
		apierror.BadRequest(
			w,
			"invalid_company_id",
			"company id must be a valid UUID",
		)
		return
	}

	var c Company

	err := h.db.QueryRowContext(r.Context(),
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

	if apierror.HandleContextError(w, err) {
		return
	}

	if errors.Is(err, sql.ErrNoRows) {
		log.Printf("company not found: id=%q", id)
		apierror.NotFound(
			w,
			"company_not_found",
			"company not found",
		)
		return
	}

	if err != nil {
		log.Printf("failed to fetch company %q: %v", id, err)
		apierror.Internal(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(c); err != nil {
		log.Printf("couldn't convert data to JSON: %v", err)
	}
}

func (h *Handler) CreateComp(w http.ResponseWriter, r *http.Request) {
	var input companyInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Printf("Failed to decode company request: %v", err)
		apierror.BadRequest(
			w,
			"invalid_json_body",
			"request body must contain valid JSON body",
		)
		return
	}

	input.trim()

	if input.Name == "" {
		apierror.BadRequest(
			w,
			"company_name_required",
			"name is required",
		)
		return
	}

	var c Company

	err := h.db.QueryRowContext(
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

	if apierror.HandleContextError(w, err) {
		return
	}

	if err != nil {
		log.Printf("failed to create company: %v", err)
		apierror.Internal(w)
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
		return
	}
}

func (h *Handler) UpdateComp(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	id, err := uuid.Parse(idString)
	if err != nil {
		log.Printf("Invalid company UUID %q: %v", idString, err)
		apierror.BadRequest(
			w,
			"invalid_company_id",
			"company_id must be a valid UUID",
		)
		return
	}

	var input companyInput

	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		log.Printf("failed to decode company update: %v", err)
		apierror.BadRequest(
			w,
			"invalid_json_body",
			"request body must contain valid JSON body",
		)
		return
	}

	input.trim()

	if input.Name == "" {
		apierror.BadRequest(
			w,
			"company_name_required",
			"name is required",
		)
		return
	}

	var c Company

	err = h.db.QueryRowContext(
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

	if apierror.HandleContextError(w, err) {
		return
	}

	if errors.Is(err, sql.ErrNoRows) {
		apierror.NotFound(
			w,
			"company_not_found",
			"company not found",
		)
		return
	}

	if err != nil {
		log.Printf("failed to update company %s: %v", id, err)
		apierror.Internal(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(c); err != nil {
		log.Printf("failed to encode updated company: %v", err)
	}
}

func (h *Handler) DeleteComp(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	id, err := uuid.Parse(idString)

	if err != nil {
		log.Printf("invalid company UUID: %q: %v", idString, err)
		apierror.BadRequest(
			w,
			"invalid_company_id",
			"company_id must be a valid UUID",
		)
		return
	}

	result, err := h.db.ExecContext(r.Context(),
		`
			DELETE FROM companies
			WHERE id = $1
		`,
		id,
	)

	if apierror.HandleContextError(w, err) {
		return
	}

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			apierror.Conflict(
				w,
				"company_has_applications",
				"company cannot be deleted because it has applications",
			)
			return
		}

		log.Printf("failed to delete company %s: %v", id, err)
		apierror.Internal(w)
		return
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		log.Printf("failed to read affected rows: %v", err)
		apierror.Internal(w)
		return
	}

	if rowsAffected == 0 {
		apierror.NotFound(
			w,
			"company_not_found",
			"company not found",
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
