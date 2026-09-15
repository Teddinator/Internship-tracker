package applications

import (
	"time"

	"github.com/google/uuid"
)

type Application struct {
	ID        int64      `json:"id"`
	CompanyID string     `json:"company_id"`
	Company   string     `json:"company"`
	Role      string     `json:"role"`
	Status    string     `json:"status"`
	AppliedAt *time.Time `json:"applied_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type applicationInput struct {
	CompanyID string  `json:"company_id"`
	Role      string  `json:"role"`
	Status    string  `json:"status"`
	AppliedAt *string `json:"applied_at"`
}

type updateApplicationInput struct {
	CompanyID uuid.UUID
	Role      string
	Status    string
	AppliedAt *time.Time
}

type applicationResponse struct {
	ID        int64     `json:"id"`
	CompanyID string    `json:"company_id"`
	Company   string    `json:"company"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	AppliedAt *string   `json:"applied_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type applicationIdempotencyInput struct {
	CompanyID string  `json:"company_id"`
	Role      string  `json:"role"`
	Status    string  `json:"status"`
	AppliedAt *string `json:"applied_at"`
}

type applicationFilter struct {
	Status    string
	CompanyID string
	Location  string
	Industry  string
}
