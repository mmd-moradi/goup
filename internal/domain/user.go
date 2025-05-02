package domain

import (
	"github.com/mmd-moradi/goup/internal/database/db"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
}

func NewUserFromDB(user db.User) *User {
	return &User{
		ID:           user.ID.String(),
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
	}
}
