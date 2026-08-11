package relationship

import (
	"backend/models"
	"context"
	"database/sql"
)

type Repository interface {
	GetPublicProfile(ctx context.Context, userID uint) (models.User, error)
	GetUserLocation(ctx context.Context, userID uint) (*float64, *float64, error)
	GetProfileRelationship(ctx context.Context, viewerID uint, targetUserID uint) (models.ProfileRelationship, error)
	HasBlockBetweenUsers(ctx context.Context, firstUserID uint, secondUserID uint) (bool, error)
	GetCompletionStatus(ctx context.Context, userID uint) (bool, error)
	GetAvatarURL(ctx context.Context, userID uint) (string, error)
	SaveProfileDecision(ctx context.Context, userID uint, targetUserID uint, decision models.ProfileDecisionValue) (SaveProfileDecisionResult, error)
	SaveProfileView(ctx context.Context, viewerID uint, viewedUserID uint) (models.ProfileView, error)
	ListMatches(ctx context.Context, userID uint) ([]models.Match, error)
	ListProfileViewers(ctx context.Context, userID uint) ([]models.ProfileViewer, error)
	ListProfileLikers(ctx context.Context, userID uint) ([]models.ProfileLiker, error)
	BlockUser(ctx context.Context, blockerID uint, blockedUserID uint) error
	UnblockUser(ctx context.Context, blockerID uint, blockedUserID uint) error
	ReportUser(ctx context.Context, reporterID uint, reportedUserID uint) error
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}
