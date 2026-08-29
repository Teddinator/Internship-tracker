package notes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/teddinator/Internship-tracker/internal/apierror"
)

type Handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

// TODO: Fixa error handling
func (h *Handler) CreateNote(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	applicationID, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		apierror.BadRequest(
			w,
			"invalid_application_id",
			"application_id must be a valid integer",
		)
		return
	}

	var input noteInput

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err = decoder.Decode(&input); err != nil {
		log.Printf("failed to decode note: %v", err)
		apierror.BadRequest(
			w,
			"invalid_json_body",
			"request body must contain valid json",
		)
		return
	}

	input.Content = strings.TrimSpace(input.Content)

	if input.Content == "" {
		apierror.BadRequest(
			w,
			"note_content_required",
			"note content is required",
		)
		return
	}

	var note Note
	err = h.db.QueryRowContext(
		r.Context(),
		`INSERT INTO notes(
			application_id,
			content
		)
		VALUES ($1, $2)
		RETURNING 
			id,
			application_id,
			content,
			created_at,
			updated_at
		`, applicationID, input.Content,
	).Scan(
		&note.ID,
		&note.ApplicationID,
		&note.Content,
		&note.CreatedAt,
		&note.UpdatedAt,
	)

	if apierror.HandleContextError(w, err) {
		return
	}

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			apierror.NotFound(
				w,
				"application_not_found",
				"application not found",
			)
			return
		}

		log.Printf("failed to create note: %v", err)
		apierror.Internal(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// TODO: Implementera location när get note by id handler fixad
	// w.Header().Set(
	// 	"Location",
	// 	fmt.Sprintf("/applications/%d/notes", applicationID),
	// )
	w.WriteHeader(http.StatusCreated)

	if err = json.NewEncoder(w).Encode(note); err != nil {
		log.Printf("Failed to encode response: %v", err)
		return
	}

}

func (h *Handler) GetNotes(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idString, 10, 64)

	if err != nil {
		apierror.BadRequest(
			w,
			"invalid_application_id",
			"application_id must be a valid integer",
		)
		return
	}

	var exists bool

	err = h.db.QueryRowContext(
		r.Context(),
		`SELECT EXISTS(
			SELECT 1
			FROM applications
			WHERE id = $1	
		)`,
		id,
	).Scan(&exists)

	if apierror.HandleContextError(w, err) {
		return
	}

	if err != nil {
		log.Printf("failed to check applications: %v", err)
		apierror.Internal(w)
		return
	}

	if !exists {
		apierror.NotFound(
			w,
			"application_not_found",
			"application not found",
		)
		return
	}

	rows, err := h.db.QueryContext(
		r.Context(),
		`
			SELECT 
				id,
				application_id,
				content,
				created_at,
				updated_at
			FROM notes
			WHERE application_id = $1

		`, id,
	)

	if apierror.HandleContextError(w, err) {
		return
	}

	if err != nil {
		log.Printf("failed to retrieve notes: %v", err)
		apierror.Internal(w)
		return
	}

	defer rows.Close()

	notes := make([]Note, 0)

	for rows.Next() {
		var note Note

		err = rows.Scan(
			&note.ID,
			&note.ApplicationID,
			&note.Content,
			&note.CreatedAt,
			&note.UpdatedAt,
		)

		if err != nil {
			if apierror.HandleContextError(w, err) {
				return
			}

			log.Printf("failed to scan notes: %v", err)
			apierror.Internal(w)
			return
		}

		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		if apierror.HandleContextError(w, err) {
			return
		}

		log.Printf("failed while reading notes: %v", err)
		apierror.Internal(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(notes); err != nil {
		log.Printf("failed to encode notes: %v", err)
		return
	}
}
