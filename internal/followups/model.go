package followups

import "time"

type Followup struct {
	ID            int64      `json:"id"`
	ApplicationID int64      `json:"application_id"`
	DueDate       time.Time  `json:"due_date"`
	Message       string     `json:"message"`
	CompletedAt   *time.Time `json:"completed_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
