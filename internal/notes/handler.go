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
		http.Error(w, "id must be a valid number", http.StatusBadRequest)
		return
	}

	var input struct {
		Content string `json:"content"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err = decoder.Decode(&input); err != nil {
		log.Printf("Failed to decode note: %v", err)
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	input.Content = strings.TrimSpace(input.Content)

	if input.Content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	var n Note
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
		&n.ID,
		&n.ApplicationID,
		&n.Content,
		&n.CreatedAt,
		&n.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			http.Error(w, "Application not found", http.StatusNotFound)
			return
		}

		log.Printf("Failed to create note: %v", err)
		http.Error(w, "Failed to create note", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// TODO: Implementera location när get note by id handler fixad
	// w.Header().Set(
	// 	"Location",
	// 	fmt.Sprintf("/applications/%d/notes", applicationID),
	// )
	w.WriteHeader(http.StatusCreated)

	if err = json.NewEncoder(w).Encode(n); err != nil {
		log.Printf("Failed to encode response: %v", err)
		return
	}

}

func (h *Handler) GetNotes(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idString, 10, 64)

	if err != nil {
		http.Error(w, "ID must be a valid number", http.StatusBadRequest)
		return
	}

	rows, err := h.db.QueryContext(
		r.Context(),
		`
			SELECT * FROM notes
			WHERE application_id = $1

		`, id,
	)

	if err != nil {
		http.Error(w, "Couldn't query database", http.StatusInternalServerError)
		return
	}

	notes := make([]Note, 0)

	for rows.Next() {
		var n Note

		err = rows.Scan(
			&n.ID,
			&n.ApplicationID,
			&n.Content,
			&n.CreatedAt,
			&n.UpdatedAt,
		)

		if err != nil {
			http.Error(w, "Couldn't convert database entry to JSON", http.StatusInternalServerError)
			return
		}

		notes = append(notes, n)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Failed to read notes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(notes); err != nil {
		http.Error(w, "Failed to encode data", http.StatusInternalServerError)
		return
	}
}
