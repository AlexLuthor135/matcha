package discovery

import (
	"backend/models"
	"context"
	"database/sql"
)

type Repository interface {
	GetDiscoveryProfile(ctx context.Context, userID uint) (models.User, error)
	ListProfileCandidates(ctx context.Context, userID uint, preferredGender string, ownGender string, excludeDecided bool) ([]models.User, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}
