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

func (s *PhotoService) GetPhotoByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*PhotoResponse, error) {
	photo, err := s.photoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if photo.UserID != userID {
		return nil, apperrors.New(apperrors.Forbidden, "You don't have access to this photo")
	}

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

func (s *PhotoService) GetPhotosByID(ctx context.Context, userID uuid.UUID, page, pageSize int) (*PhotosResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	photos, total, err := s.photoRepo.GetByUserID(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	totalPages := (total + pageSize - 1) / pageSize

	photoResponses := make([]PhotoResponse, len(photos))
	for i, photo := range photos {
		photoResponses[i] = PhotoResponse{
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
		}
	}

	return &PhotosResponse{
		Photos:     photoResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil

}

func (s *PhotoService) UpdatePhoto(ctx context.Context, id uuid.UUID, input PhotoUpdateInput, userID uuid.UUID) (*PhotoResponse, error) {
	if err := validator.Validate(input); err != nil {
		return nil, apperrors.Wrap(err, apperrors.BadRequest)
	}

	photo, err := s.photoRepo.GetByID(ctx, id)

	if err != nil {
		return nil, err
	}

	if photo.UserID != userID {
		return nil, apperrors.New(apperrors.Forbidden, "You don't have access to this photo")
	}

	photo.Title = input.Title
	photo.Description = input.Description
	photo.UpdatedAt = time.Now()

	err = s.photoRepo.Update(ctx, photo)

	if err != nil {
		return nil, err
	}

	s.logger.Info().
		Str("userID", userID.String()).
		Str("photoID", photo.ID.String()).
		Msg("photo updated successfully")

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

func (s *PhotoService) DeletePhoto(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	photo, err := s.photoRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if photo.UserID != userID {
		return apperrors.New(apperrors.Forbidden, "You don't have access to this photo")
	}

	err = s.storage.DeletePhoto(ctx, photo.StoragePath)
	if err != nil {
		return err
	}

	err = s.photoRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	s.logger.Info().
		Str("userID", userID.String()).
		Str("photoID", photo.ID.String()).
		Msg("photo deleted successfully")

	return nil

}
