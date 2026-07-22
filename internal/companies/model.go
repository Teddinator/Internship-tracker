package companies

import "time"

type Company struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Website   *string   `json:"website"`
	Industry  *string   `json:"industry"`
	Location  *string   `json:"location"`
	Notes     *string   `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
