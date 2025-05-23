package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func NewUser(username, email string) *User {
	now := time.Now()
	return &User{
		ID:        uuid.New(),
		Username:  username,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
