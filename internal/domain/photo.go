package domain

import (
	"time"

	"github.com/google/uuid"
)

type Photo struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	FileName    string    `json:"filename"`
	FileSize    int64     `json:"file_size"`
	ContentType string    `json:"content_type"`
	StoragePath string    `json:"storage_path"`
	PublicURL   string    `json:"public_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewPhoto(userID uuid.UUID, fileSize int64, title, description, fileName, contentType, storagePath, publicURL string) *Photo {
	now := time.Now()
	return &Photo{
		ID:          uuid.New(),
		UserID:      userID,
		Title:       title,
		Description: description,
		FileName:    fileName,
		FileSize:    fileSize,
		ContentType: contentType,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
