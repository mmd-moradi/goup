package storage

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	cfg "github.com/mmd-moradi/goup/configs"
	"github.com/mmd-moradi/goup/internal/domain"
	"github.com/mmd-moradi/goup/pkg/apperrors"
	"github.com/rs/zerolog"
)

type S3StorageService struct {
	s3Client *s3.Client
	bucket   string
	loger    zerolog.Logger
}

func NewS3StorageService(cfg *cfg.AWSConfig, logger zerolog.Logger) (*S3StorageService, error) {
	awsConfig, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsConfig)

	return &S3StorageService{
		s3Client: s3Client,
		bucket:   cfg.S3Bucket,
		loger:    logger,
	}, nil
}

func (s *S3StorageService) UploadPhoto(ctx context.Context, data []byte, userID uuid.UUID, photo *domain.Photo) error {
	timestamp := time.Now().Format("20060102-150405")
	storagePath := fmt.Sprintf(
		"users/%s/photos/%s-%s%s",
		userID.String(),
		timestamp,
		uuid.New().String()[:8],
		filepath.Ext(photo.FileName),
	)

	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(storagePath),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(photo.ContentType),
	}

	_, err := s.s3Client.PutObject(ctx, input)
	if err != nil {
		return apperrors.NewWithFormat(apperrors.InternalServer, "failed to upload photo: %v", err)
	}

	photo.StoragePath = storagePath
	photo.PublicURL = fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucket, storagePath)

	s.loger.Info().
		Str("userID", userID.String()).
		Str("photoID", photo.ID.String()).
		Str("path", storagePath).
		Msg("Photo uploaded to s3 successfully")

	return nil
}
