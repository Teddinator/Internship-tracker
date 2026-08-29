package notes

import "time"

type Note struct {
	ID            int64     `json:"id"`
	ApplicationID int64     `json:"application_id"`
	Content       string    `json:"content"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type noteInput struct {
	Content string `json:"content"`
}
