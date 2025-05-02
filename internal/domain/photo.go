package domain

import (
	"time"

	"github.com/mmd-moradi/goup/internal/database/db"
)

type Photo struct {
	ID          string
	UserID      string
	Title       string
	Description string
	S3Key       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewPhotoFromDB(photo db.Photo) *Photo {
	return &Photo{
		ID:          photo.ID.String(),
		UserID:      photo.UserID.String(),
		Title:       photo.Title,
		Description: photo.Description.String,
		S3Key:       photo.S3Key,
		CreatedAt:   photo.CreatedAt.Time,
		UpdatedAt:   photo.UpdatedAt.Time,
	}
}
