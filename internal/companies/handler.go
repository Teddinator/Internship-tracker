package companies

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

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

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if _, err := uuid.Parse(id); err != nil {
		log.Printf("Invalid company UUID %q: %v", id, err)
		http.Error(w, "Invalid company ID", http.StatusInternalServerError)
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

func (h *Handler) CreateComp(w http.ResponseWriter, r *http.Request) {
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

	err = h.db.QueryRowContext(
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

func (h *Handler) UpdateComp(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) DeleteComp(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	id, err := uuid.Parse(idString)

	if err != nil {
		log.Printf("Invalid company UUID: %q: %v", idString, err)
		http.Error(w, "Invalid company ID", http.StatusBadRequest)
		return
	}

	result, err := h.db.ExecContext(r.Context(),
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
