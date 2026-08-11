package user

import (
	"backend/models"
	"context"
	"database/sql"
	"time"
)

type Repository interface {
	GetCompletionStatus(ctx context.Context, userID uint) (bool, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	GetUserByUserName(ctx context.Context, userName string) (models.User, error)
	CreateUser(ctx context.Context, newUser models.User, token models.AccountToken) (models.User, error)
	GetPasswordHash(ctx context.Context, userID uint) (string, error)
	UpdatePasswordHashAndRevokeSessions(ctx context.Context, userID uint, passwordHash string) error
	GetProfile(ctx context.Context, userID uint) (models.User, error)
	ListProfileCandidates(ctx context.Context, userID uint, preferredGender string, ownGender string, excludeDecided bool) ([]models.User, error)
	CompleteProfile(ctx context.Context, userID uint, input CompleteProfileInput) (bool, error)
	UpdateProfile(ctx context.Context, userID uint, input UpdateProfileInput) error
	UpdateUser(ctx context.Context, userID uint, input UserUpdateInput, verificationToken *models.AccountToken) (UpdateUserResult, error)
	GetAvatarURL(ctx context.Context, userID uint) (string, error)
	UpdateAvatarURL(ctx context.Context, userID uint, avatarURL string) (string, error)
	CreatePhotos(ctx context.Context, userID uint, photoURLs []string, maxAllowed int) ([]models.Photo, error)
	DeletePhoto(ctx context.Context, userID uint, photoID uint) (string, error)
	SaveProfileDecision(ctx context.Context, userID uint, targetUserID uint, decision models.ProfileDecisionValue) (SaveProfileDecisionResult, error)
	GetProfileRelationship(ctx context.Context, viewerID uint, targetUserID uint) (models.ProfileRelationship, error)
	ListMatches(ctx context.Context, userID uint) ([]models.Match, error)
	SaveProfileView(ctx context.Context, viewerID uint, viewedUserID uint) (models.ProfileView, error)
	ListProfileViewers(ctx context.Context, userID uint) ([]models.ProfileViewer, error)
	ListProfileLikers(ctx context.Context, userID uint) ([]models.ProfileLiker, error)
	BlockUser(ctx context.Context, blockerID uint, blockedUserID uint) error
	UnblockUser(ctx context.Context, blockerID uint, blockedUserID uint) error
	HasBlockBetweenUsers(ctx context.Context, firstUserID uint, secondUserID uint) (bool, error)
	ReportUser(ctx context.Context, reporterID uint, reportedUserID uint) error
	UpdateLastSeen(ctx context.Context, userID uint, lastSeenAt time.Time) error
	GetUserLocation(ctx context.Context, userID uint) (*float64, *float64, error)
	VerifyEmail(cxt context.Context, tokenHash string) error
	ReplaceAccountToken(ctx context.Context, token models.AccountToken) error
	ResetPasswordWithToken(ctx context.Context, tokenHash string, newPasswordHash string) error
	CreateAuthSession(ctx context.Context, session models.AuthSession) error
	UseAuthSession(ctx context.Context, sessionID string, userID uint) error
	RevokeAuthSession(ctx context.Context, sessionID string, userID uint) error
	RevokeAllAuthSessions(ctx context.Context, userID uint) error
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}
