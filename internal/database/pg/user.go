package pg

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmd-moradi/goup/internal/database/db"
	"github.com/mmd-moradi/goup/internal/database/repositories"
	"github.com/mmd-moradi/goup/internal/domain"
)

type PgUserRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewPgUserRepository(pool *pgxpool.Pool) *PgUserRepository {
	return &PgUserRepository{
		queries: db.New(pool),
		pool:    pool,
	}
}

func (r *PgUserRepository) CreateUser(ctx context.Context, email, passwordHash string) (*domain.User, error) {
	createdUser, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
	})

	if err != nil {
		return nil, err
	}

	return domain.NewUserFromDB(createdUser), nil

}

func (r *PgUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return domain.NewUserFromDB(user), nil
}

func (r *PgUserRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	user, err := r.queries.GetUserByID(ctx, userID)

	if err != nil {
		return nil, err
	}

	return domain.NewUserFromDB(user), nil
}

func (r *PgUserRepository) WithTx(ctx context.Context, txOptions pgx.TxOptions, fn func(repositories.UserRepository) error) error {
	tx, err := r.pool.BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txRepo := &PgUserRepository{
		queries: r.queries.WithTx(tx),
		pool:    r.pool,
	}

	if err := fn(txRepo); err != nil {
		return err
	}

	return tx.Commit(ctx)

}
