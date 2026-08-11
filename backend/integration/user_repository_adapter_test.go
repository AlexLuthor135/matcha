package integration

import (
	"backend/account"
	"backend/discovery"
	"backend/models"
	"backend/profile"
	"backend/relationship"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"
)

func hashAccountToken(rawToken string) string {
	hash := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(hash[:])
}

func stringPointer(value string) *string {
	return &value
}

type CompleteProfileInput = profile.CompleteProfileInput
type LocationInput = profile.LocationInput
type UpdateProfileInput = profile.UpdateProfileInput
type UserUpdateInput = account.UserUpdateInput
type UpdateUserResult = account.UpdateUserResult
type SaveProfileDecisionResult = relationship.SaveProfileDecisionResult

var UserErrors = struct {
	UserNotFound              error
	TargetUserNotFound        error
	InvalidLocation           error
	InvalidPasswordResetToken error
	PhotoLimitExceeded        error
	PhotoNotFound             error
	UserAlreadyExists         error
}{
	UserNotFound:              errors.New("user not found"),
	TargetUserNotFound:        errors.New("target user not found"),
	InvalidLocation:           errors.New("user location is invalid"),
	InvalidPasswordResetToken: errors.New("password reset link is invalid or expired"),
	PhotoLimitExceeded:        errors.New("profile photo limit exceeded"),
	PhotoNotFound:             errors.New("photo not found"),
	UserAlreadyExists:         errors.New("username or email already exists"),
}

type PostgresRepository struct {
	account      *account.PostgresRepository
	profile      *profile.PostgresRepository
	discovery    *discovery.PostgresRepository
	relationship *relationship.PostgresRepository
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		account:      account.NewPostgresRepository(db),
		profile:      profile.NewPostgresRepository(db),
		discovery:    discovery.NewPostgresRepository(db),
		relationship: relationship.NewPostgresRepository(db),
	}
}

func normalizeRepositoryError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, account.AccountErrors.UserNotFound),
		errors.Is(err, profile.ProfileErrors.UserNotFound),
		errors.Is(err, relationship.RelationshipErrors.UserNotFound),
		errors.Is(err, discovery.DiscoveryErrors.UserNotFound):
		return UserErrors.UserNotFound
	case errors.Is(err, relationship.RelationshipErrors.TargetUserNotFound):
		return UserErrors.TargetUserNotFound
	case errors.Is(err, relationship.RelationshipErrors.InvalidLocation):
		return UserErrors.InvalidLocation
	case errors.Is(err, account.AccountErrors.InvalidPasswordResetToken):
		return UserErrors.InvalidPasswordResetToken
	case errors.Is(err, profile.ProfileErrors.PhotoLimitExceeded):
		return UserErrors.PhotoLimitExceeded
	case errors.Is(err, profile.ProfileErrors.PhotoNotFound):
		return UserErrors.PhotoNotFound
	case errors.Is(err, account.AccountErrors.UserAlreadyExists):
		return UserErrors.UserAlreadyExists
	default:
		return err
	}
}

func (r *PostgresRepository) CreateUser(ctx context.Context, user models.User, token models.AccountToken) (models.User, error) {
	result, err := r.account.CreateUser(ctx, user, token)
	return result, normalizeRepositoryError(err)
}
func (r *PostgresRepository) UpdateUser(ctx context.Context, id uint, input UserUpdateInput, token *models.AccountToken) (UpdateUserResult, error) {
	result, err := r.account.UpdateUser(ctx, id, input, token)
	return result, normalizeRepositoryError(err)
}
func (r *PostgresRepository) ReplaceAccountToken(ctx context.Context, token models.AccountToken) error {
	return normalizeRepositoryError(r.account.ReplaceAccountToken(ctx, token))
}
func (r *PostgresRepository) VerifyEmail(ctx context.Context, hash string) error {
	return normalizeRepositoryError(r.account.VerifyEmail(ctx, hash))
}
func (r *PostgresRepository) ResetPasswordWithToken(ctx context.Context, hash string, password string) error {
	return normalizeRepositoryError(r.account.ResetPasswordWithToken(ctx, hash, password))
}
func (r *PostgresRepository) CompleteProfile(ctx context.Context, id uint, input CompleteProfileInput) (bool, error) {
	result, err := r.profile.CompleteProfile(ctx, id, input)
	return result, normalizeRepositoryError(err)
}
func (r *PostgresRepository) UpdateProfile(ctx context.Context, id uint, input UpdateProfileInput) error {
	return normalizeRepositoryError(r.profile.UpdateProfile(ctx, id, input))
}
func (r *PostgresRepository) GetProfile(ctx context.Context, id uint) (models.User, error) {
	result, err := r.profile.GetProfile(ctx, id)
	return result, normalizeRepositoryError(err)
}
func (r *PostgresRepository) CreatePhotos(ctx context.Context, id uint, urls []string, max int) ([]models.Photo, error) {
	result, err := r.profile.CreatePhotos(ctx, id, urls, max)
	return result, normalizeRepositoryError(err)
}
func (r *PostgresRepository) DeletePhoto(ctx context.Context, userID uint, photoID uint) (string, error) {
	result, err := r.profile.DeletePhoto(ctx, userID, photoID)
	return result, normalizeRepositoryError(err)
}
func (r *PostgresRepository) UpdateLastSeen(ctx context.Context, id uint, seenAt time.Time) error {
	return normalizeRepositoryError(r.profile.UpdateLastSeen(ctx, id, seenAt))
}
func (r *PostgresRepository) ListProfileCandidates(ctx context.Context, id uint, preferred string, own string, exclude bool) ([]models.User, error) {
	result, err := r.discovery.ListProfileCandidates(ctx, id, preferred, own, exclude)
	return result, normalizeRepositoryError(err)
}
func (r *PostgresRepository) GetUserLocation(ctx context.Context, id uint) (*float64, *float64, error) {
	latitude, longitude, err := r.relationship.GetUserLocation(ctx, id)
	return latitude, longitude, normalizeRepositoryError(err)
}
func (r *PostgresRepository) GetProfileRelationship(ctx context.Context, a uint, b uint) (models.ProfileRelationship, error) {
	result, err := r.relationship.GetProfileRelationship(ctx, a, b)
	return result, normalizeRepositoryError(err)
}
func (r *PostgresRepository) HasBlockBetweenUsers(ctx context.Context, a uint, b uint) (bool, error) {
	result, err := r.relationship.HasBlockBetweenUsers(ctx, a, b)
	return result, normalizeRepositoryError(err)
}
func (r *PostgresRepository) SaveProfileDecision(ctx context.Context, a uint, b uint, decision models.ProfileDecisionValue) (SaveProfileDecisionResult, error) {
	result, err := r.relationship.SaveProfileDecision(ctx, a, b, decision)
	return result, normalizeRepositoryError(err)
}
func (r *PostgresRepository) SaveProfileView(ctx context.Context, a uint, b uint) (models.ProfileView, error) {
	result, err := r.relationship.SaveProfileView(ctx, a, b)
	return result, normalizeRepositoryError(err)
}
func (r *PostgresRepository) ListMatches(ctx context.Context, id uint) ([]models.Match, error) {
	result, err := r.relationship.ListMatches(ctx, id)
	return result, normalizeRepositoryError(err)
}
func (r *PostgresRepository) ListProfileViewers(ctx context.Context, id uint) ([]models.ProfileViewer, error) {
	result, err := r.relationship.ListProfileViewers(ctx, id)
	return result, normalizeRepositoryError(err)
}
func (r *PostgresRepository) ListProfileLikers(ctx context.Context, id uint) ([]models.ProfileLiker, error) {
	result, err := r.relationship.ListProfileLikers(ctx, id)
	return result, normalizeRepositoryError(err)
}
func (r *PostgresRepository) BlockUser(ctx context.Context, a uint, b uint) error {
	return normalizeRepositoryError(r.relationship.BlockUser(ctx, a, b))
}
func (r *PostgresRepository) UnblockUser(ctx context.Context, a uint, b uint) error {
	return normalizeRepositoryError(r.relationship.UnblockUser(ctx, a, b))
}
func (r *PostgresRepository) ReportUser(ctx context.Context, a uint, b uint) error {
	return normalizeRepositoryError(r.relationship.ReportUser(ctx, a, b))
}
