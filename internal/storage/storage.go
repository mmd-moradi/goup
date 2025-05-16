package storage

import (
	"context"

	"github.com/google/uuid"
	"github.com/mmd-moradi/goup/internal/domain"
)

type StorageService interface {
	UploadPhoto(ctx context.Context, data []byte, userID uuid.UUID, photo *domain.Photo) error
	GetPhoto(ctx context.Context, storagePath string) ([]byte, string, error)
	DeletePhoto(ctx context.Context, storagePath string) error
}
