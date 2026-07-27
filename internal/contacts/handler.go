package contacts

import (
	"database/sql"
	"encoding/json"
	"errors"
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

func (h *Handler) CreateContact(w http.ResponseWriter, r *http.Request) {
	log.Println("CreateContact handler reached")
	idString := chi.URLParam(r, "id")

	companyID, err := uuid.Parse(idString)

	if err != nil {
		http.Error(w, "ID must be a valid UUID", http.StatusBadRequest)
		return
	}

	var input struct {
		Name        string `json:"name"`
		Email       string `json:"email"`
		LinkedInURL string `json:"linkedin_url"`
		Role        string `json:"role"`
		Notes       string `json:"notes"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err = decoder.Decode(&input); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	Email := stringPtrOrNil(input.Email)
	LinkedInURL := stringPtrOrNil(input.LinkedInURL)
	Role := stringPtrOrNil(input.Role)
	Notes := stringPtrOrNil(input.Notes)

	if input.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	var c Contact

	err = h.db.QueryRowContext(
		r.Context(),
		`
			INSERT INTO contacts (
				company_id,
				name,
				email,
				linkedin_url,
				role,
				notes
			)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING
				id,
				company_id,
				name,
				email,
				linkedin_url,
				role,
				notes,
				created_at,
				updated_at

		`,
		companyID,
		input.Name,
		Email,
		LinkedInURL,
		Role,
		Notes,
	).Scan(
		&c.ID,
		&c.CompanyID,
		&c.Name,
		&c.Email,
		&c.LinkedInURL,
		&c.Role,
		&c.Notes,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			http.Error(w, "Company not found", http.StatusNotFound)
			return
		}

		log.Printf("Failed to create contact: %v", err)
		http.Error(w, "Failed to create contact", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err = json.NewEncoder(w).Encode(c); err != nil {
		http.Error(w, "Failed to encode contact", http.StatusInternalServerError)
		return
	}
}
