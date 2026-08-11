package account

import (
	"backend/models"
	"context"
	"database/sql"
)

type Repository interface {
	GetCompletionStatus(ctx context.Context, userID uint) (bool, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	GetUserByUserName(ctx context.Context, userName string) (models.User, error)
	CreateUser(ctx context.Context, newUser models.User, token models.AccountToken) (models.User, error)
	GetPasswordHash(ctx context.Context, userID uint) (string, error)
	UpdatePasswordHashAndRevokeSessions(ctx context.Context, userID uint, passwordHash string) error
	UpdateUser(ctx context.Context, userID uint, input UserUpdateInput, verificationToken *models.AccountToken) (UpdateUserResult, error)
	VerifyEmail(ctx context.Context, tokenHash string) error
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
	return &PostgresRepository{db: db}
}
