package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/mmd-moradi/goup/internal/domain"
)

type UserRepository interface {
	CreateUser(ctx context.Context, email, passwordHash string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)

	WithTx(ctx context.Context, txOption pgx.TxOptions, fn func(UserRepository) error) error
}
