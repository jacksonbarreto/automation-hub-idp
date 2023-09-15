package dto

import "github.com/google/uuid"

type UserRequest struct {
	ID       uuid.UUID `json:"id,omitempty"`
	Email    string    `json:"email,omitempty"`
	Password string    `json:"password,omitempty"`
}
