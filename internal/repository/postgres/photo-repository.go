package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmd-moradi/goup/internal/domain"
	repositories "github.com/mmd-moradi/goup/internal/repository"
	"github.com/mmd-moradi/goup/internal/repository/postgres/db"
	"github.com/mmd-moradi/goup/pkg/apperrors"
)

type PhotoRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewPhotoRepository(pool *pgxpool.Pool) *PhotoRepository {
	return &PhotoRepository{
		queries: db.New(pool),
		pool:    pool,
	}
}

func (r *PhotoRepository) Create(ctx context.Context, photo *domain.Photo) error {
	_, err := r.queries.CreatePhoto(ctx, db.CreatePhotoParams{
		ID:          photo.ID,
		UserID:      photo.UserID,
		Title:       photo.Title,
		Description: pgtype.Text{String: photo.Description, Valid: photo.Description != ""},
		FileName:    photo.FileName,
		FileSize:    photo.FileSize,
		ContentType: photo.ContentType,
		StoragePath: photo.StoragePath,
		PublicUrl:   pgtype.Text{String: photo.PublicURL, Valid: photo.PublicURL != ""},
		CreatedAt:   TimeToTimestamptz(photo.CreatedAt),
		UpdatedAt:   TimeToTimestamptz(photo.UpdatedAt),
	})

	if err != nil {
		return apperrors.NewWithFormat(apperrors.InternalServer, "failed to create photo: %v", err)
	}

	return nil
}

func (r *PhotoRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Photo, error) {
	photo, err := r.queries.GetPhotoByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NewWithFormat(apperrors.NotFound, "photo with id %s not found", id)
		}
		return nil, apperrors.NewWithFormat(apperrors.InternalServer, "failed to get photo: %v", err)
	}

	return &domain.Photo{
		ID:          photo.ID,
		UserID:      photo.UserID,
		Title:       photo.Title,
		Description: photo.Description.String,
		FileName:    photo.FileName,
		FileSize:    photo.FileSize,
		ContentType: photo.ContentType,
		StoragePath: photo.StoragePath,
		PublicURL:   photo.PublicUrl.String,
		CreatedAt:   TimestamptzToTime(photo.CreatedAt),
		UpdatedAt:   TimestamptzToTime(photo.UpdatedAt),
	}, nil
}

func (r *PhotoRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Photo, int, error) {
	photos, err := r.queries.ListPhotosByUserID(ctx, db.ListPhotosByUserIDParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, apperrors.NewWithFormat(apperrors.InternalServer, "failed to list photos: %v", err)
	}

	count, err := r.queries.CountPhotosByUserID(ctx, userID)
	if err != nil {
		return nil, 0, apperrors.NewWithFormat(apperrors.InternalServer, "failed to count photos: %v", err)
	}

	result := make([]*domain.Photo, len(photos))
	for i, photo := range photos {
		result[i] = &domain.Photo{
			ID:          photo.ID,
			UserID:      photo.UserID,
			Title:       photo.Title,
			Description: photo.Description.String,
			FileName:    photo.FileName,
			FileSize:    photo.FileSize,
			ContentType: photo.ContentType,
			StoragePath: photo.StoragePath,
			PublicURL:   photo.PublicUrl.String,
			CreatedAt:   TimestamptzToTime(photo.CreatedAt),
			UpdatedAt:   TimestamptzToTime(photo.UpdatedAt),
		}
	}

	return result, int(count), nil
}

func (r *PhotoRepository) Update(ctx context.Context, photo *domain.Photo) error {
	_, err := r.queries.UpdatePhoto(ctx, db.UpdatePhotoParams{
		ID:          photo.ID,
		Title:       photo.Title,
		Description: pgtype.Text{String: photo.Description, Valid: photo.Description != ""},
		UpdatedAt:   TimeToTimestamptz(photo.UpdatedAt),
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.NewWithFormat(apperrors.NotFound, "photo with id %s not found", photo.ID)
		}
		return apperrors.NewWithFormat(apperrors.InternalServer, "failed to update photo: %v", err)
	}

	return nil
}

func (r *PhotoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeletePhoto(ctx, id)
	if err != nil {
		return apperrors.NewWithFormat(apperrors.InternalServer, "failed to delete photo: %v", err)
	}

	return nil
}

func (r *PhotoRepository) WithTx(ctx context.Context, txOptions pgx.TxOptions, fn func(repositories.PhotoRepository) error) error {
	tx, err := r.pool.BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txRepo := &PhotoRepository{
		queries: r.queries.WithTx(tx),
		pool:    r.pool,
	}

	if err := fn(txRepo); err != nil {
		return err
	}

	return tx.Commit(ctx)

}
