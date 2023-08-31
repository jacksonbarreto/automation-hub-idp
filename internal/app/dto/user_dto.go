package dto

import (
	"github.com/google/uuid"
)

type UserDTO struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
}
