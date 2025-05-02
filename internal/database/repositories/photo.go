package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/mmd-moradi/goup/internal/database/db"
	"github.com/mmd-moradi/goup/internal/domain"
)

type PhotoRepository interface {
	CreatePhoto(ctx context.Context, params db.CreatePhotoParams) (*domain.Photo, error)
	GetPhotoByID(ctx context.Context, id string) (*domain.Photo, error)
	ListUserPhotos(ctx context.Context, userID string, limit int32, offset int32) ([]domain.Photo, error)
	UpdatePhotoDetails(ctx context.Context, id, title, description string) error
	DeletePhoto(ctx context.Context, id string) error

	WithTx(ctx context.Context, txOption pgx.TxOptions, fn func(PhotoRepository) error) error
}
