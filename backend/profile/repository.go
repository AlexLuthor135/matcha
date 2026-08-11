package profile

import (
	"backend/models"
	"context"
	"database/sql"
	"time"
)

type Repository interface {
	GetProfile(ctx context.Context, userID uint) (models.User, error)
	CompleteProfile(ctx context.Context, userID uint, input CompleteProfileInput) (bool, error)
	UpdateProfile(ctx context.Context, userID uint, input UpdateProfileInput) error
	GetAvatarURL(ctx context.Context, userID uint) (string, error)
	UpdateAvatarURL(ctx context.Context, userID uint, avatarURL string) (string, error)
	CreatePhotos(ctx context.Context, userID uint, photoURLs []string, maxAllowed int) ([]models.Photo, error)
	DeletePhoto(ctx context.Context, userID uint, photoID uint) (string, error)
	UpdateLastSeen(ctx context.Context, userID uint, lastSeenAt time.Time) error
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}
