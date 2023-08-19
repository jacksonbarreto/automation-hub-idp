package models

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID                 uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Email              string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	Password           string    `gorm:"type:varchar(255);not null"`
	FirstAccess        bool      `gorm:"default:true"`
	FailedAttempts     int       `gorm:"default:0"`
	LastAttempt        *time.Time
	IsBlocked          bool `gorm:"default:false"`
	BlockedUntil       *time.Time
	CreatedAt          time.Time `gorm:"autoUpdateTime"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"`
	ResetPasswordToken string    `gorm:"type:varchar(255);index"`
	ResetTokenExpires  *time.Time
}

func SimulateUser() User {
	currentTime := time.Now()
	futureTime := currentTime.Add(time.Hour * 2)

	return User{
		ID:                 uuid.New(),
		Email:              "john.doe@example.com",
		Password:           "hashedPasswordHere",
		FirstAccess:        true,
		FailedAttempts:     1,
		LastAttempt:        &currentTime,
		IsBlocked:          false,
		BlockedUntil:       &futureTime,
		CreatedAt:          currentTime,
		UpdatedAt:          currentTime,
		ResetPasswordToken: "randomTokenHere",
		ResetTokenExpires:  &futureTime,
	}
}
