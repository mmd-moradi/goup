package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mmd-moradi/goup/internal/auth"
	"github.com/mmd-moradi/goup/internal/domain"
	repositories "github.com/mmd-moradi/goup/internal/repository"
	"github.com/mmd-moradi/goup/pkg/apperrors"
	"github.com/mmd-moradi/goup/pkg/validator"
	"github.com/rs/zerolog"
)

type UserService struct {
	repo     repositories.UserRepository
	tokenSvc auth.TokenService
	logger   zerolog.Logger
}

type UserRegistrationInput struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type UserLoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type AuthResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

func NewUserService(repo repositories.UserRepository, tokenSvc auth.TokenService, logger zerolog.Logger) *UserService {
	return &UserService{
		repo:     repo,
		tokenSvc: tokenSvc,
		logger:   logger,
	}
}

func (s *UserService) Register(ctx context.Context, input UserRegistrationInput) (*AuthResponse, error) {
	if err := validator.Validate(input); err != nil {
		return nil, apperrors.Wrap(err, apperrors.BadRequest)
	}

	_, err := s.repo.GetByEmail(ctx, input.Email)
	if err == nil {
		return nil, apperrors.New(apperrors.Conflict, "email already exists")
	}

	_, err = s.repo.GetByUsername(ctx, input.Username)
	if err == nil {
		return nil, apperrors.New(apperrors.Conflict, "username already exists")
	}

	hashedPassword, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, apperrors.NewWithFormat(apperrors.InternalServer, "failed to hash password: %v", err)
	}

	user := domain.NewUser(input.Username, input.Email)
	user.PasswordHash = hashedPassword

	err = s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	tokenDetails, err := s.tokenSvc.GenerateToken(user.ID)
	if err != nil {
		return nil, err
	}

	s.logger.Info().
		Str("userID", user.ID.String()).
		Msg("user registered successfully")

	return &AuthResponse{
		User: UserResponse{
			ID:        user.ID.String(),
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		Token: tokenDetails.Token,
	}, nil
}

func (s *UserService) Login(ctx context.Context, input UserLoginInput) (*AuthResponse, error) {
	if err := validator.Validate(input); err != nil {
		return nil, apperrors.Wrap(err, apperrors.BadRequest)
	}

	user, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, apperrors.New(apperrors.Unauthorized, "invalid email or password")
	}

	if !auth.CheckPasswordHash(input.Password, user.PasswordHash) {
		return nil, apperrors.New(apperrors.Unauthorized, "invalid email or password")
	}

	tokenDetails, err := s.tokenSvc.GenerateToken(user.ID)
	if err != nil {
		return nil, err
	}

	s.logger.Info().
		Str("userID", user.ID.String()).
		Msg("user logged in successfully")

	return &AuthResponse{
		User: UserResponse{
			ID:        user.ID.String(),
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		Token: tokenDetails.Token,
	}, nil
}

func (s *UserService) GetUserByID(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &UserResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *UserService) Logout(ctx context.Context, token string) error {
	err := s.tokenSvc.RevokeToken(token)
	if err != nil {
		return err
	}
	s.logger.Info().Msg("user logged out successfully")
	return nil
}
