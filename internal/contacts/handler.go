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

func (h *Handler) GetContacts(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	id, err := uuid.Parse(idString)

	if err != nil {
		apierror.BadRequest(
			w,
			"invalid_company_id",
			"company_id must be a valid UUID",
		)
		return
	}

	rows, err := h.db.QueryContext(
		r.Context(),
		`
			SELECT
				c.id,
				c.company_id,
				c.name,
				c.email,
				c.linkedin_url,
				c.role,
				c.notes,
				c.created_at,
				c.updated_at
			FROM contacts as c
			WHERE c.company_id = $1
			ORDER BY created_at DESC
		`, id,
	)

	if apierror.HandleContextError(w, err) {
		return
	}

	if err != nil {
		log.Printf("Failed to retreive contacts: %v", err)
		apierror.Internal(w)
		return
	}

	defer rows.Close()

	contacts := make([]Contact, 0)

	for rows.Next() {
		var contact Contact

		err = rows.Scan(
			&contact.ID,
			&contact.CompanyID,
			&contact.Name,
			&contact.Email,
			&contact.LinkedInURL,
			&contact.Role,
			&contact.Notes,
			&contact.CreatedAt,
			&contact.UpdatedAt,
		)

		if err != nil {
			if apierror.HandleContextError(w, err) {
				return
			}

			log.Printf("failed to scan contacts: %v", err)
			apierror.Internal(w)
			return
		}

		contacts = append(contacts, contact)
	}

	if err := rows.Err(); err != nil {
		if apierror.HandleContextError(w, err) {
			return
		}

		log.Printf("failed to read contacts: %v", err)
		apierror.Internal(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(contacts); err != nil {
		log.Printf("failed to encode json: %v", err)
		return
	}
}

func (h *Handler) CreateContact(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	companyID, err := uuid.Parse(idString)

	if err != nil {
		apierror.BadRequest(
			w,
			"invalid_company_id",
			"company_id must be a valid UUID",
		)
		return
	}

	var input contactInput

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err = decoder.Decode(&input); err != nil {
		apierror.BadRequest(
			w,
			"invalid_json_body",
			"request body must contain valid JSON body",
		)
		return
	}

	input.Name = strings.TrimSpace(input.Name)

	Email := stringPtrOrNil(input.Email)
	LinkedInURL := stringPtrOrNil(input.LinkedInURL)
	Role := stringPtrOrNil(input.Role)
	Notes := stringPtrOrNil(input.Notes)

	if input.Name == "" {
		apierror.BadRequest(
			w,
			"contact_name_required",
			"name is required",
		)
		return
	}

	var contact Contact

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
		&contact.ID,
		&contact.CompanyID,
		&contact.Name,
		&contact.Email,
		&contact.LinkedInURL,
		&contact.Role,
		&contact.Notes,
		&contact.CreatedAt,
		&contact.UpdatedAt,
	)

	if apierror.HandleContextError(w, err) {
		return
	}

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			apierror.NotFound(
				w,
				"company_not_found",
				"company not found",
			)
			return
		}

		log.Printf("Failed to create contact: %v", err)
		apierror.Internal(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	//TODO: Add location when adding handler GetContactByID
	// w.Header().Set(
	// 	"Location",
	// 	fmt.Sprintf("companies/%s/contacts/%d", companyID, contact.ID),
	// )
	w.WriteHeader(http.StatusCreated)

	if err = json.NewEncoder(w).Encode(contact); err != nil {
		log.Printf("failed to encode contact: %v", err)
		return
	}
}
