package followups

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
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

	response := toFollowUpResponse(followUp)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode data to JSON", http.StatusInternalServerError)
		return
	}
}

func (h *handler) GetDueFollowUps(w http.ResponseWriter, r *http.Request) {
	const query = `
		SELECT
			id,
			application_id,
			due_date,
			message,
			completed_at,
			created_at,
			updated_at
		FROM followups
		WHERE due_date <= CURRENT_DATE
		AND completed_at IS NULL
		ORDER BY due_date, created_at ASC
	`
	rows, err := h.db.QueryContext(
		r.Context(),
		query,
	)

	if err != nil {
		http.Error(w, "Unable to query follow-ups", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	dueFollowUps := make([]FollowUpResponse, 0)

	for rows.Next() {
		var followUp FollowUp

		err = rows.Scan(
			&followUp.ID,
			&followUp.ApplicationID,
			&followUp.DueDate,
			&followUp.Message,
			&followUp.CompletedAt,
			&followUp.CreatedAt,
			&followUp.UpdatedAt,
		)

		if err != nil {
			http.Error(w, "Failed to scan followups", http.StatusInternalServerError)
			return
		}

		response := toFollowUpResponse(followUp)
		dueFollowUps = append(dueFollowUps, response)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Failed while reading follow-ups", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(dueFollowUps); err != nil {
		log.Printf("failed to encode due follow-ups: %v", err)
		return
	}
}

func (h *handler) CompleteFollowUp(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		http.Error(w, "id must be a valid integer", http.StatusBadRequest)
		return
	}

	const query = `
		UPDATE followups
		SET completed_at = COALESCE(completed_at, CURRENT_TIMESTAMP),
			updated_at = CASE
				WHEN completed_at IS NULL THEN CURRENT_TIMESTAMP
				ELSE updated_at
			END
		WHERE id = $1
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
		id,
	).Scan(
		&followUp.ID,
		&followUp.ApplicationID,
		&followUp.DueDate,
		&followUp.Message,
		&followUp.CompletedAt,
		&followUp.CreatedAt,
		&followUp.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "follow-up not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "failed to complete follow-up", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "json/application")
	w.WriteHeader(http.StatusOK)

	response := toFollowUpResponse(followUp)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed top encode to JSON : %v", err)
		return
	}
}
