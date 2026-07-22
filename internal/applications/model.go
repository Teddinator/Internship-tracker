package applications

import "time"

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
