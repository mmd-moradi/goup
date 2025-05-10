package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"time"

	"github.com/google/uuid"
	"github.com/mmd-moradi/goup/configs"
	"github.com/mmd-moradi/goup/pkg/apperrors"
	"github.com/redis/go-redis/v9"
)

type TokenService struct {
	redis       *redis.Client
	config      *configs.AuthConfig
	tokenPrefix string
}

type TokenDetails struct {
	Token     string
	UserID    uuid.UUID
	ExpiresAt time.Time
}

func NewTokenService(redis *redis.Client, config *configs.AuthConfig) *TokenService {
	return &TokenService{
		redis:       redis,
		config:      config,
		tokenPrefix: "auth_token:",
	}
}

func (s *TokenService) GenerateToken(userID uuid.UUID) (*TokenDetails, error) {
	token, err := generateRandomString(32)
	if err != nil {
		return nil, apperrors.New(apperrors.InternalServer, "failed to generate token")
	}

	expiresAt := time.Now().Add(time.Duration(s.config.TokenExpirationMin) * time.Minute)

	tokenDetail := &TokenDetails{
		Token:     token,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}

	key := s.tokenPrefix + token
	err = s.redis.Set(
		context.Background(),
		key,
		userID.String(),
		time.Until(expiresAt),
	).Err()

	if err != nil {
		return nil, apperrors.NewWithFormat(apperrors.InternalServer, "failed to save token: %v", err)
	}

	return tokenDetail, nil
}

func (s *TokenService) ValidateToken(token string) (uuid.UUID, error) {
	key := s.tokenPrefix + token
	userIDStr, err := s.redis.Get(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			return uuid.Nil, apperrors.New(apperrors.Unauthorized, "Invalid or expired token")
		}
		return uuid.Nil, apperrors.NewWithFormat(apperrors.InternalServer, "Failed to validate token: %v", err)
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, apperrors.New(apperrors.InternalServer, "Invalid user ID format in token")
	}
	return userID, nil
}

func (s *TokenService) RevokeToken(token string) error {
	key := s.tokenPrefix + token
	err := s.redis.Del(context.Background(), key).Err()
	if err != nil {
		return apperrors.NewWithFormat(apperrors.InternalServer, "Failed to revoke token: %v", err)
	}
	return nil
}

func (s *TokenService) RefreshToken(token string) (*TokenDetails, error) {
	userIDStr, err := s.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	err = s.RevokeToken(token)
	if err != nil {
		return nil, err
	}

	return s.GenerateToken(userIDStr)

}

func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length/2)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	plainText := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes)

	hash := sha256.Sum256([]byte(plainText))
	hashedToken := hash[:]

	return string(hashedToken), nil
}
