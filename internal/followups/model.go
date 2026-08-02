package followups

import "time"

type FollowUp struct {
	ID            int64      `json:"id"`
	ApplicationID int64      `json:"application_id"`
	DueDate       time.Time  `json:"due_date"`
	Message       string     `json:"message"`
	CompletedAt   *time.Time `json:"completed_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type CreateFollowUpRequest struct {
	DueDate string `json:"due_date"`
	Message string `json:"message"`
}

type FollowUpResponse struct {
	ID            int64      `json:"id"`
	ApplicationID int64      `json:"application_id"`
	DueDate       string     `json:"due_date"`
	Message       string     `json:"message"`
	CompletedAt   *time.Time `json:"completed_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
