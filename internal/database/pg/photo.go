package pg

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmd-moradi/goup/internal/database/db"
	"github.com/mmd-moradi/goup/internal/database/repositories"
	"github.com/mmd-moradi/goup/internal/domain"
)

type PgPhotoRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewPgPhotoRepository(pool *pgxpool.Pool) *PgPhotoRepository {
	return &PgPhotoRepository{
		queries: db.New(pool),
		pool:    pool,
	}
}

func (r *PgPhotoRepository) CreatePhoto(ctx context.Context, params db.CreatePhotoParams) (*domain.Photo, error) {
	createdPhoto, err := r.queries.CreatePhoto(ctx, params)

	if err != nil {
		return nil, err
	}

	return domain.NewPhotoFromDB(createdPhoto), nil

}

func (r *PgPhotoRepository) GetPhotoByID(ctx context.Context, id string) (*domain.Photo, error) {
	photoID, er := uuid.Parse(id)
	if er != nil {
		return nil, er
	}

	photo, err := r.queries.GetPhotoByID(ctx, photoID)
	if err != nil {
		return nil, err
	}
	return domain.NewPhotoFromDB(photo), nil
}

func (r *PgPhotoRepository) ListUserPhotos(ctx context.Context, id string, limit int32, offset int32) ([]domain.Photo, error) {
	photoID, er := uuid.Parse(id)
	if er != nil {
		return nil, er
	}
	photos, err := r.queries.ListUserPhotos(ctx, db.ListUserPhotosParams{
		UserID: pgtype.UUID{Bytes: photoID, Valid: true},
		Limit:  limit,
		Offset: offset,
	})

	if err != nil {
		return nil, err
	}

	domainPhotos := make([]domain.Photo, len(photos))
	for i, photo := range photos {
		domainPhotos[i] = domain.Photo{
			ID:          photoID.String(),
			UserID:      photo.UserID.String(),
			Title:       photo.Title,
			Description: photo.Description.String,
			S3Key:       photo.S3Key,
			CreatedAt:   photo.CreatedAt.Time,
			UpdatedAt:   photo.UpdatedAt.Time,
		}
	}

	return domainPhotos, nil
}

func (r *PgPhotoRepository) UpdatePhotoDetails(ctx context.Context, id, title, description string) error {
	photoID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	_, err = r.queries.UpdatePhotoDetails(ctx, db.UpdatePhotoDetailsParams{
		ID:          photoID,
		Title:       title,
		Description: pgtype.Text{String: description, Valid: true},
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *PgPhotoRepository) DeletePhoto(ctx context.Context, id string) error {
	photoID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	err = r.queries.DeletePhoto(ctx, photoID)
	if err != nil {
		return err
	}
	return nil
}

func (r *PgPhotoRepository) WithTx(ctx context.Context, txOptions pgx.TxOptions, fn func(repositories.PhotoRepository) error) error {
	tx, err := r.pool.BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txRepo := &PgPhotoRepository{
		queries: r.queries.WithTx(tx),
		pool:    r.pool,
	}

	if err := fn(txRepo); err != nil {
		return err
	}

	return tx.Commit(ctx)

}
