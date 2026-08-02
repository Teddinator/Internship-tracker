package followups

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *handler {
	return &handler{
		db: db,
	}
}

func (h *handler) CreateFollowUp(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	applicationID, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		http.Error(w, "ID must be a valid integer", http.StatusBadRequest)
		return
	}

	var req CreateFollowUpRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		http.Error(w, "due_date must use YYY-MM-DD format", http.StatusBadRequest)
		return
	}

	const query = `
		INSERT INTO followups (
			application_id,
			due_date,
			message
		)
		VALUES ($1, $2, $3)
		RETURNING
			id,
			application_id,
			due_date,
			message,
			completed_at,
			created_at,
			updated_at
	`

	var followUp FollowUp

	err = h.db.QueryRowContext(
		r.Context(),
		query,
		applicationID, dueDate, req.Message,
	).Scan(
		&followUp.ID,
		&followUp.ApplicationID,
		&followUp.DueDate,
		&followUp.Message,
		&followUp.CompletedAt,
		&followUp.CreatedAt,
		&followUp.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			http.Error(w, "Application not found", http.StatusNotFound)
			return
		}

		http.Error(w, "could not create follow-up", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(followUp); err != nil {
		http.Error(w, "Failed to encode data to JSON", http.StatusInternalServerError)
		return
	}
}
