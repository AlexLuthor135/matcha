package account

import (
	"backend/models"
	"context"
)

type fakeUserRepository struct {
	Repository
	getCompletionStatusFn                 func(context.Context, uint) (bool, error)
	getUserByEmailFn                      func(context.Context, string) (models.User, error)
	getUserByUserNameFn                   func(context.Context, string) (models.User, error)
	createUserFn                          func(context.Context, models.User, models.AccountToken) (models.User, error)
	getPasswordHashFn                     func(context.Context, uint) (string, error)
	updatePasswordHashAndRevokeSessionsFn func(context.Context, uint, string) error
	updateUserFn                          func(context.Context, uint, UserUpdateInput, *models.AccountToken) (UpdateUserResult, error)
	verifyEmailFn                         func(context.Context, string) error
	replaceAccountTokenFn                 func(context.Context, models.AccountToken) error
	resetPasswordWithTokenFn              func(context.Context, string, string) error
	createAuthSessionFn                   func(context.Context, models.AuthSession) error
	useAuthSessionFn                      func(context.Context, string, uint) error
	revokeAuthSessionFn                   func(context.Context, string, uint) error
	revokeAllAuthSessionsFn               func(context.Context, uint) error
}

func (r *fakeUserRepository) GetCompletionStatus(ctx context.Context, id uint) (bool, error) {
	return r.getCompletionStatusFn(ctx, id)
}
func (r *fakeUserRepository) GetUserByEmail(ctx context.Context, value string) (models.User, error) {
	return r.getUserByEmailFn(ctx, value)
}
func (r *fakeUserRepository) GetUserByUserName(ctx context.Context, value string) (models.User, error) {
	return r.getUserByUserNameFn(ctx, value)
}
func (r *fakeUserRepository) CreateUser(ctx context.Context, user models.User, token models.AccountToken) (models.User, error) {
	return r.createUserFn(ctx, user, token)
}
func (r *fakeUserRepository) GetPasswordHash(ctx context.Context, id uint) (string, error) {
	return r.getPasswordHashFn(ctx, id)
}
func (r *fakeUserRepository) UpdatePasswordHashAndRevokeSessions(ctx context.Context, id uint, hash string) error {
	return r.updatePasswordHashAndRevokeSessionsFn(ctx, id, hash)
}
func (r *fakeUserRepository) UpdateUser(ctx context.Context, id uint, input UserUpdateInput, token *models.AccountToken) (UpdateUserResult, error) {
	return r.updateUserFn(ctx, id, input, token)
}
func (r *fakeUserRepository) VerifyEmail(ctx context.Context, hash string) error {
	return r.verifyEmailFn(ctx, hash)
}
func (r *fakeUserRepository) ReplaceAccountToken(ctx context.Context, token models.AccountToken) error {
	return r.replaceAccountTokenFn(ctx, token)
}
func (r *fakeUserRepository) ResetPasswordWithToken(ctx context.Context, tokenHash string, passwordHash string) error {
	return r.resetPasswordWithTokenFn(ctx, tokenHash, passwordHash)
}
func (r *fakeUserRepository) CreateAuthSession(ctx context.Context, session models.AuthSession) error {
	return r.createAuthSessionFn(ctx, session)
}
func (r *fakeUserRepository) UseAuthSession(ctx context.Context, sessionID string, userID uint) error {
	return r.useAuthSessionFn(ctx, sessionID, userID)
}
func (r *fakeUserRepository) RevokeAuthSession(ctx context.Context, sessionID string, userID uint) error {
	return r.revokeAuthSessionFn(ctx, sessionID, userID)
}
func (r *fakeUserRepository) RevokeAllAuthSessions(ctx context.Context, userID uint) error {
	return r.revokeAllAuthSessionsFn(ctx, userID)
}

type fakeEmailSender struct {
	sendVerificationEmailFn  func(context.Context, string, string) error
	sendPasswordResetEmailFn func(context.Context, string, string) error
}

func (s *fakeEmailSender) SendVerificationEmail(ctx context.Context, email string, token string) error {
	return s.sendVerificationEmailFn(ctx, email, token)
}
func (s *fakeEmailSender) SendPasswordResetEmail(ctx context.Context, email string, token string) error {
	return s.sendPasswordResetEmailFn(ctx, email, token)
}
