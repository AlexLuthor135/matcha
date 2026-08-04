package user

import (
	"backend/models"
	"context"
	"database/sql"
)

type Repository interface {
	GetCompletionStatus(ctx context.Context, userID uint) (bool, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	CreateUser(ctx context.Context, newUser models.User) (models.User, error)
	GetPasswordHash(ctx context.Context, userID uint) (string, error)
	UpdatePasswordHash(ctx context.Context, userID uint, passwordHash string) error
	GetProfile(ctx context.Context, userID uint) (models.User, error)
	GetProfileFeed(ctx context.Context, userID uint, limit int) ([]models.User, error)
	CompleteProfile(ctx context.Context, userID uint, gender string, preferences string, bio string, interests []string) (bool, error)
	UpdateProfile(ctx context.Context, userID uint, gender *string, preferences *string, bio *string, interests *[]string) error
	UpdateUser(ctx context.Context, userID uint, userName *string, firstName *string, lastName *string, email *string) error
	GetAvatarURL(ctx context.Context, userID uint) (string, error)
	UpdateAvatarURL(ctx context.Context, userID uint, avatarURL string) (string, error)
	CreatePhotos(ctx context.Context, userID uint, photoURLs []string, maxAllowed int) ([]models.Photo, error)
	DeletePhoto(ctx context.Context, userID uint, photoID uint) (string, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}
