package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mmd-moradi/goup/internal/domain"
)

type PhotoRepository interface {
	Create(ctx context.Context, photo *domain.Photo) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Photo, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Photo, int, error)
	Update(ctx context.Context, photo *domain.Photo) error
	Delete(ctx context.Context, id uuid.UUID) error

	WithTx(ctx context.Context, txOption pgx.TxOptions, fn func(PhotoRepository) error) error
}
