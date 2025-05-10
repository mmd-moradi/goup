package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmd-moradi/goup/internal/domain"
	repositories "github.com/mmd-moradi/goup/internal/repository"
	"github.com/mmd-moradi/goup/internal/repository/postgres/db"
	"github.com/mmd-moradi/goup/pkg/apperrors"
)

type UserRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		queries: db.New(pool),
		pool:    pool,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	_, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    TimeToTimestamptz(user.CreatedAt),
		UpdatedAt:    TimeToTimestamptz(user.UpdatedAt),
	})

	if err != nil {
		return apperrors.NewWithFormat(apperrors.InternalServer, "failed to create user %v", err)
	}

	return nil

}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NewWithFormat(apperrors.NotFound, "user with id %s not found", id)
		}
		return nil, apperrors.NewWithFormat(apperrors.InternalServer, "failed to get user: %v", err)
	}

	return &domain.User{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    TimestamptzToTime(user.CreatedAt),
		UpdatedAt:    TimestamptzToTime(user.UpdatedAt),
	}, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NewWithFormat(apperrors.NotFound, "user with email %s not found", email)
		}
		return nil, apperrors.NewWithFormat(apperrors.InternalServer, "failed to get user: %v", err)
	}

	return &domain.User{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    TimestamptzToTime(user.CreatedAt),
		UpdatedAt:    TimestamptzToTime(user.UpdatedAt),
	}, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	user, err := r.queries.GetUserByUserName(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NewWithFormat(apperrors.NotFound, "user with username %s not found", username)
		}
		return nil, apperrors.NewWithFormat(apperrors.InternalServer, "failed to get user: %v", err)
	}

	return &domain.User{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    TimestamptzToTime(user.CreatedAt),
		UpdatedAt:    TimestamptzToTime(user.UpdatedAt),
	}, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	_, err := r.queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		UpdatedAt: TimeToTimestamptz(user.UpdatedAt),
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.NewWithFormat(apperrors.NotFound, "user with id %s not found", user.ID)
		}
		return apperrors.NewWithFormat(apperrors.InternalServer, "failed to update user: %v", err)
	}

	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteUser(ctx, id)
	if err != nil {
		return apperrors.NewWithFormat(apperrors.InternalServer, "failed to delete user: %v", err)
	}

	return nil
}

func (r *UserRepository) WithTx(ctx context.Context, txOptions pgx.TxOptions, fn func(repositories.UserRepository) error) error {
	tx, err := r.pool.BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txRepo := &UserRepository{
		queries: r.queries.WithTx(tx),
		pool:    r.pool,
	}

	if err := fn(txRepo); err != nil {
		return err
	}

	return tx.Commit(ctx)

}
