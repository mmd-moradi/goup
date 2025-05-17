package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mmd-moradi/goup/internal/domain"
	repositories "github.com/mmd-moradi/goup/internal/repository"
	"github.com/mmd-moradi/goup/internal/storage"
	"github.com/mmd-moradi/goup/pkg/apperrors"
	"github.com/mmd-moradi/goup/pkg/validator"
	"github.com/rs/zerolog"
)

type PhotoService struct {
	photoRepo repositories.PhotoRepository
	userRepo  repositories.UserRepository
	storage   storage.StorageService
	logger    zerolog.Logger
}

type PhotoUploadInput struct {
	Title       string `json:"title" validate:"required,max=255"`
	Description string `json:"description" validate:"max=1000"`
	FileName    string `json:"file_name" validate:"required"`
	FileSize    int64  `json:"file_size" validate:"required"`
	ContentType string `json:"content_type" validate:"required"`
}

type PhotoUpdateInput struct {
	Title       string `json:"title" validate:"max=255"`
	Description string `json:"description" validate:"max=1000"`
}

type PhotoResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	ContentType string    `json:"content_type"`
	PublicURL   string    `json:"public_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PhotosResponse struct {
	Photos     []PhotoResponse `json:"photos"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

func NewPhotoService(
	photoRepo repositories.PhotoRepository,
	userRepo repositories.UserRepository,
	storage storage.StorageService,
	logger zerolog.Logger,
) *PhotoService {

	return &PhotoService{
		photoRepo: photoRepo,
		userRepo:  userRepo,
		storage:   storage,
		logger:    logger,
	}
}

func (s *PhotoService) UploadPhoto(ctx context.Context, input PhotoUploadInput, data []byte, userID uuid.UUID) (*PhotoResponse, error) {
	if err := validator.Validate(input); err != nil {
		return nil, apperrors.Wrap(err, apperrors.BadRequest)
	}
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	photo := domain.NewPhoto(
		userID,
		input.FileSize,
		input.Title,
		input.Description,
		input.FileName,
		input.ContentType,
		"",
		"",
	)

	err = s.storage.UploadPhoto(ctx, data, userID, photo)
	if err != nil {
		return nil, err
	}

	err = s.photoRepo.Create(ctx, photo)
	if err != nil {
		cleanUpErr := s.storage.DeletePhoto(ctx, photo.StoragePath)
		if cleanUpErr != nil {
			s.logger.Error().Err(cleanUpErr).Msg("failed to clean up photo after databse error")
		}
		return nil, err
	}

	s.logger.Info().
		Str("userID", userID.String()).
		Str("photoID", photo.ID.String()).
		Msg("photo uploaded successfully")

	return &PhotoResponse{
		ID:          photo.ID.String(),
		UserID:      photo.UserID.String(),
		Title:       photo.Title,
		Description: photo.Description,
		FileName:    photo.FileName,
		FileSize:    photo.FileSize,
		ContentType: photo.ContentType,
		PublicURL:   photo.PublicURL,
		CreatedAt:   photo.CreatedAt,
		UpdatedAt:   photo.UpdatedAt,
	}, nil
}
