package contacts

import "time"

type Contact struct {
	ID          int64     `json:"id"`
	CompanyID   string    `json:"company_id"`
	Name        string    `json:"name"`
	Email       *string   `json:"email"`
	LinkedInURL *string   `json:"linkedin_url"`
	Role        *string   `json:"role"`
	Notes       *string   `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
