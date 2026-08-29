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
	"github.com/teddinator/Internship-tracker/internal/apierror"
)

type handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *handler {
	return &handler{
		db: db,
	}
}

// TODO: Fixa error handling
func (h *handler) CreateFollowUp(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	applicationID, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		apierror.BadRequest(
			w,
			"invalid_application_id",
			"application id must be a valid integer",
		)
		return
	}

	var req CreateFollowUpRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		apierror.BadRequest(
			w,
			"invalid_json_body",
			"request body must contain valid json",
		)
		return
	}

	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		apierror.BadRequest(
			w,
			"invalid_due_date",
			"due_date must use format YYYY-MM-DD",
		)
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

		log.Printf("failed to create followup: %v", err)
		apierror.Internal(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := toFollowUpResponse(followUp)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("failed to encode json: %v", err)
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

	if apierror.HandleContextError(w, err) {
		return
	}

	if err != nil {
		log.Printf("failed to query followups: %v", err)
		apierror.Internal(w)
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
			if apierror.HandleContextError(w, err) {
				return
			}

			log.Printf("failed to scan followups: %v", err)
			apierror.Internal(w)
			return
		}

		response := toFollowUpResponse(followUp)
		dueFollowUps = append(dueFollowUps, response)
	}

	if err := rows.Err(); err != nil {
		if apierror.HandleContextError(w, err) {
			return
		}

		log.Printf("failed while reading followups: %v", err)
		apierror.Internal(w)
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
		apierror.BadRequest(
			w,
			"invalid_followup_id",
			"followup_id must be a valid integer",
		)
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

	if apierror.HandleContextError(w, err) {
		return
	}

	if errors.Is(err, sql.ErrNoRows) {
		apierror.NotFound(
			w,
			"followup_not_found",
			"followup not found",
		)
		return
	}

	if err != nil {
		log.Printf("failed to complete followup: %v", err)
		apierror.Internal(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := toFollowUpResponse(followUp)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("failed to encode to json: %v", err)
		return
	}
}

func (h *handler) DeleteFollowUp(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		apierror.BadRequest(
			w,
			"invalid_followup_id",
			"followup_id must be a valid integer",
		)
		return
	}

	const query = `
		DELETE FROM followups
		WHERE id = $1
	`

	result, err := h.db.ExecContext(
		r.Context(),
		query,
		id,
	)

	if err != nil {
		log.Printf("failed to delete follow-up: %v", err)
		apierror.Internal(w)
		return
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		log.Printf("failed to check affected rows: %v", err)
		apierror.Internal(w)
		return
	}

	if rowsAffected == 0 {
		apierror.NotFound(
			w,
			"followup_not_found",
			"followup not found",
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
