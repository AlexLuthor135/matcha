package relationship

import (
	"backend/models"
	"context"
)

type fakeUserRepository struct {
	Repository
	getProfileFn             func(context.Context, uint) (models.User, error)
	getUserLocationFn        func(context.Context, uint) (*float64, *float64, error)
	getProfileRelationshipFn func(context.Context, uint, uint) (models.ProfileRelationship, error)
	hasBlockBetweenUsersFn   func(context.Context, uint, uint) (bool, error)
	getCompletionStatusFn    func(context.Context, uint) (bool, error)
	getAvatarURLFn           func(context.Context, uint) (string, error)
	saveDecisionFn           func(context.Context, uint, uint, models.ProfileDecisionValue) (SaveProfileDecisionResult, error)
	saveProfileViewFn        func(context.Context, uint, uint) (models.ProfileView, error)
	listMatchesFn            func(context.Context, uint) ([]models.Match, error)
	listProfileViewersFn     func(context.Context, uint) ([]models.ProfileViewer, error)
	listProfileLikersFn      func(context.Context, uint) ([]models.ProfileLiker, error)
	blockUserFn              func(context.Context, uint, uint) error
	unblockUserFn            func(context.Context, uint, uint) error
	reportUserFn             func(context.Context, uint, uint) error
}

func (r *fakeUserRepository) GetPublicProfile(ctx context.Context, id uint) (models.User, error) {
	return r.getProfileFn(ctx, id)
}
func (r *fakeUserRepository) GetUserLocation(ctx context.Context, id uint) (*float64, *float64, error) {
	return r.getUserLocationFn(ctx, id)
}
func (r *fakeUserRepository) GetProfileRelationship(ctx context.Context, a uint, b uint) (models.ProfileRelationship, error) {
	return r.getProfileRelationshipFn(ctx, a, b)
}
func (r *fakeUserRepository) HasBlockBetweenUsers(ctx context.Context, a uint, b uint) (bool, error) {
	return r.hasBlockBetweenUsersFn(ctx, a, b)
}
func (r *fakeUserRepository) GetCompletionStatus(ctx context.Context, id uint) (bool, error) {
	return r.getCompletionStatusFn(ctx, id)
}
func (r *fakeUserRepository) GetAvatarURL(ctx context.Context, id uint) (string, error) {
	return r.getAvatarURLFn(ctx, id)
}
func (r *fakeUserRepository) SaveProfileDecision(ctx context.Context, a uint, b uint, decision models.ProfileDecisionValue) (SaveProfileDecisionResult, error) {
	return r.saveDecisionFn(ctx, a, b, decision)
}
func (r *fakeUserRepository) SaveProfileView(ctx context.Context, a uint, b uint) (models.ProfileView, error) {
	return r.saveProfileViewFn(ctx, a, b)
}
func (r *fakeUserRepository) ListMatches(ctx context.Context, id uint) ([]models.Match, error) {
	return r.listMatchesFn(ctx, id)
}
func (r *fakeUserRepository) ListProfileViewers(ctx context.Context, id uint) ([]models.ProfileViewer, error) {
	return r.listProfileViewersFn(ctx, id)
}
func (r *fakeUserRepository) ListProfileLikers(ctx context.Context, id uint) ([]models.ProfileLiker, error) {
	return r.listProfileLikersFn(ctx, id)
}
func (r *fakeUserRepository) BlockUser(ctx context.Context, a uint, b uint) error {
	return r.blockUserFn(ctx, a, b)
}
func (r *fakeUserRepository) UnblockUser(ctx context.Context, a uint, b uint) error {
	return r.unblockUserFn(ctx, a, b)
}
func (r *fakeUserRepository) ReportUser(ctx context.Context, a uint, b uint) error {
	return r.reportUserFn(ctx, a, b)
}

type fakeUserNotifier struct {
	notifyMatchFn       func(context.Context, uint, uint) (models.Notification, error)
	notifyProfileViewFn func(context.Context, uint, uint) (models.Notification, error)
	notifyLikeFn        func(context.Context, uint, uint) (models.Notification, error)
	notifyUnlikeFn      func(context.Context, uint, uint) (models.Notification, error)
}

func (n *fakeUserNotifier) NotifyMatch(ctx context.Context, a uint, b uint) (models.Notification, error) {
	return n.notifyMatchFn(ctx, a, b)
}
func (n *fakeUserNotifier) NotifyProfileView(ctx context.Context, a uint, b uint) (models.Notification, error) {
	return n.notifyProfileViewFn(ctx, a, b)
}
func (n *fakeUserNotifier) NotifyLike(ctx context.Context, a uint, b uint) (models.Notification, error) {
	return n.notifyLikeFn(ctx, a, b)
}
func (n *fakeUserNotifier) NotifyUnlike(ctx context.Context, a uint, b uint) (models.Notification, error) {
	return n.notifyUnlikeFn(ctx, a, b)
}

type fakeUserPresence struct {
	isUserOnlineFn func(context.Context, uint) bool
}

func (p *fakeUserPresence) IsUserOnline(ctx context.Context, id uint) bool {
	return p.isUserOnlineFn(ctx, id)
}
