package user

import (
	"backend/models"
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"time"
)

type fakeUserRepository struct {
	Repository
	getUserByEmailFn                      func(context.Context, string) (models.User, error)
	getUserByUserNameFn                   func(context.Context, string) (models.User, error)
	createUserFn                          func(context.Context, models.User, models.AccountToken) (models.User, error)
	updateUserFn                          func(context.Context, uint, UserUpdateInput, *models.AccountToken) (UpdateUserResult, error)
	replaceAccountTokenFn                 func(context.Context, models.AccountToken) error
	verifyEmailFn                         func(context.Context, string) error
	resetPasswordWithTokenFn              func(context.Context, string, string) error
	getPasswordHashFn                     func(context.Context, uint) (string, error)
	updatePasswordHashAndRevokeSessionsFn func(context.Context, uint, string) error
	createAuthSessionFn                   func(context.Context, models.AuthSession) error
	useAuthSessionFn                      func(context.Context, string, uint) error
	revokeAuthSessionFn                   func(context.Context, string, uint) error
	revokeAllAuthSessionsFn               func(context.Context, uint) error
	getCompletionStatusFn                 func(context.Context, uint) (bool, error)
	getUserLocationFn                     func(context.Context, uint) (*float64, *float64, error)
	listProfileCandidatesFn               func(context.Context, uint, string, string, bool) ([]models.User, error)
	completeProfileFn                     func(context.Context, uint, CompleteProfileInput) (bool, error)
	updateProfileFn                       func(context.Context, uint, UpdateProfileInput) error
	getProfileFn                          func(context.Context, uint) (models.User, error)
	getProfileRelationshipFn              func(context.Context, uint, uint) (models.ProfileRelationship, error)
	getAvatarURLFn                        func(context.Context, uint) (string, error)
	updateAvatarURLFn                     func(context.Context, uint, string) (string, error)
	createPhotosFn                        func(context.Context, uint, []string, int) ([]models.Photo, error)
	deletePhotoFn                         func(context.Context, uint, uint) (string, error)
	saveDecisionFn                        func(context.Context, uint, uint, models.ProfileDecisionValue) (SaveProfileDecisionResult, error)
	listMatchesFn                         func(context.Context, uint) ([]models.Match, error)
	saveProfileViewFn                     func(context.Context, uint, uint) (models.ProfileView, error)
	listProfileViewersFn                  func(context.Context, uint) ([]models.ProfileViewer, error)
	listProfileLikersFn                   func(context.Context, uint) ([]models.ProfileLiker, error)
	blockUserFn                           func(context.Context, uint, uint) error
	unblockUserFn                         func(context.Context, uint, uint) error
	hasBlockBetweenUsersFn                func(context.Context, uint, uint) (bool, error)
	reportUserFn                          func(context.Context, uint, uint) error
	updateLastSeenFn                      func(context.Context, uint, time.Time) error
}

func (repo *fakeUserRepository) CreateUser(
	ctx context.Context,
	newUser models.User,
	token models.AccountToken,
) (models.User, error) {
	if repo.createUserFn == nil {
		panic("unexpected CreateUser call")
	}
	return repo.createUserFn(ctx, newUser, token)
}

func (repo *fakeUserRepository) UpdateUser(
	ctx context.Context,
	userID uint,
	input UserUpdateInput,
	verificationToken *models.AccountToken,
) (UpdateUserResult, error) {
	if repo.updateUserFn == nil {
		panic("unexpected UpdateUser call")
	}
	return repo.updateUserFn(ctx, userID, input, verificationToken)
}

func (repo *fakeUserRepository) GetUserByUserName(
	ctx context.Context,
	userName string,
) (models.User, error) {
	if repo.getUserByUserNameFn == nil {
		panic("unexpected GetUserByUserName call")
	}
	return repo.getUserByUserNameFn(ctx, userName)
}

func (repo *fakeUserRepository) GetUserByEmail(
	ctx context.Context,
	email string,
) (models.User, error) {
	if repo.getUserByEmailFn == nil {
		panic("unexpected GetUserByEmail call")
	}
	return repo.getUserByEmailFn(ctx, email)
}

func (repo *fakeUserRepository) ReplaceAccountToken(
	ctx context.Context,
	token models.AccountToken,
) error {
	if repo.replaceAccountTokenFn == nil {
		panic("unexpected ReplaceAccountToken call")
	}
	return repo.replaceAccountTokenFn(ctx, token)
}

func (repo *fakeUserRepository) VerifyEmail(
	ctx context.Context,
	tokenHash string,
) error {
	if repo.verifyEmailFn == nil {
		panic("unexpected VerifyEmail call")
	}
	return repo.verifyEmailFn(ctx, tokenHash)
}

func (repo *fakeUserRepository) ResetPasswordWithToken(
	ctx context.Context,
	tokenHash string,
	newPasswordHash string,
) error {
	if repo.resetPasswordWithTokenFn == nil {
		panic("unexpected ResetPasswordWithToken call")
	}
	return repo.resetPasswordWithTokenFn(ctx, tokenHash, newPasswordHash)
}

func (repo *fakeUserRepository) GetPasswordHash(
	ctx context.Context,
	userID uint,
) (string, error) {
	if repo.getPasswordHashFn == nil {
		panic("unexpected GetPasswordHash call")
	}
	return repo.getPasswordHashFn(ctx, userID)
}

func (repo *fakeUserRepository) UpdatePasswordHashAndRevokeSessions(
	ctx context.Context,
	userID uint,
	passwordHash string,
) error {
	if repo.updatePasswordHashAndRevokeSessionsFn == nil {
		panic("unexpected UpdatePasswordHashAndRevokeSessions call")
	}
	return repo.updatePasswordHashAndRevokeSessionsFn(ctx, userID, passwordHash)
}

func (repo *fakeUserRepository) CreateAuthSession(
	ctx context.Context,
	session models.AuthSession,
) error {
	if repo.createAuthSessionFn == nil {
		panic("unexpected CreateAuthSession call")
	}
	return repo.createAuthSessionFn(ctx, session)
}

func (repo *fakeUserRepository) UseAuthSession(
	ctx context.Context,
	sessionID string,
	userID uint,
) error {
	if repo.useAuthSessionFn == nil {
		panic("unexpected UseAuthSession call")
	}
	return repo.useAuthSessionFn(ctx, sessionID, userID)
}

func (repo *fakeUserRepository) RevokeAuthSession(
	ctx context.Context,
	sessionID string,
	userID uint,
) error {
	if repo.revokeAuthSessionFn == nil {
		panic("unexpected RevokeAuthSession call")
	}
	return repo.revokeAuthSessionFn(ctx, sessionID, userID)
}

func (repo *fakeUserRepository) RevokeAllAuthSessions(
	ctx context.Context,
	userID uint,
) error {
	if repo.revokeAllAuthSessionsFn == nil {
		panic("unexpected RevokeAllAuthSessions call")
	}
	return repo.revokeAllAuthSessionsFn(ctx, userID)
}

func (repo *fakeUserRepository) GetUserLocation(
	ctx context.Context,
	userID uint,
) (*float64, *float64, error) {
	if repo.getUserLocationFn == nil {
		panic("unexpected GetUserLocation call")
	}
	return repo.getUserLocationFn(ctx, userID)
}

func (repo *fakeUserRepository) ListProfileCandidates(
	ctx context.Context,
	userID uint,
	preferredGender string,
	ownGender string,
	excludeDecided bool,
) ([]models.User, error) {
	if repo.listProfileCandidatesFn == nil {
		panic("unexpected ListProfileCandidates call")
	}
	return repo.listProfileCandidatesFn(ctx, userID, preferredGender, ownGender, excludeDecided)
}

func (repo *fakeUserRepository) UpdateProfile(
	ctx context.Context,
	userID uint,
	input UpdateProfileInput,
) error {
	if repo.updateProfileFn == nil {
		panic("unexpected UpdateProfile call")
	}
	return repo.updateProfileFn(ctx, userID, input)
}

func (repo *fakeUserRepository) CompleteProfile(
	ctx context.Context,
	userID uint,
	input CompleteProfileInput,
) (bool, error) {
	if repo.completeProfileFn == nil {
		panic("unexpected CompleteProfile call")
	}
	return repo.completeProfileFn(ctx, userID, input)
}

func (repo *fakeUserRepository) GetCompletionStatus(ctx context.Context, userID uint) (bool, error) {
	if repo.getCompletionStatusFn == nil {
		panic("unexpected GetCompletionStatus call")
	}
	return repo.getCompletionStatusFn(ctx, userID)
}

func (repo *fakeUserRepository) GetProfile(ctx context.Context, userID uint) (models.User, error) {
	if repo.getProfileFn == nil {
		panic("unexpected GetProfile call")
	}
	return repo.getProfileFn(ctx, userID)
}

func (repo *fakeUserRepository) GetProfileRelationship(
	ctx context.Context,
	viewerID uint,
	targetUserID uint,
) (models.ProfileRelationship, error) {
	if repo.getProfileRelationshipFn == nil {
		panic("unexpected GetProfileRelationship call")
	}
	return repo.getProfileRelationshipFn(ctx, viewerID, targetUserID)
}

func (repo *fakeUserRepository) GetAvatarURL(ctx context.Context, userID uint) (string, error) {
	if repo.getAvatarURLFn == nil {
		panic("unexpected GetAvatarURL call")
	}
	return repo.getAvatarURLFn(ctx, userID)
}

func (repo *fakeUserRepository) UpdateAvatarURL(ctx context.Context, userID uint, avatarURL string) (string, error) {
	if repo.updateAvatarURLFn == nil {
		panic("unexpected UpdateAvatarURL call")
	}
	return repo.updateAvatarURLFn(ctx, userID, avatarURL)
}

func (repo *fakeUserRepository) CreatePhotos(ctx context.Context, userID uint, photoURLs []string, maxAllowed int) ([]models.Photo, error) {
	if repo.createPhotosFn == nil {
		panic("unexpected CreatePhotos call")
	}
	return repo.createPhotosFn(ctx, userID, photoURLs, maxAllowed)
}

func (repo *fakeUserRepository) DeletePhoto(ctx context.Context, userID uint, photoID uint) (string, error) {
	if repo.deletePhotoFn == nil {
		panic("unexpected DeletePhoto call")
	}
	return repo.deletePhotoFn(ctx, userID, photoID)
}

func (repo *fakeUserRepository) SaveProfileDecision(
	ctx context.Context,
	userID uint,
	targetUserID uint,
	decision models.ProfileDecisionValue,
) (SaveProfileDecisionResult, error) {
	if repo.saveDecisionFn == nil {
		panic("unexpected SaveProfileDecision call")
	}
	return repo.saveDecisionFn(ctx, userID, targetUserID, decision)
}

func (repo *fakeUserRepository) ListMatches(ctx context.Context, userID uint) ([]models.Match, error) {
	if repo.listMatchesFn == nil {
		panic("unexpected ListMatches call")
	}
	return repo.listMatchesFn(ctx, userID)
}

func (repo *fakeUserRepository) SaveProfileView(
	ctx context.Context,
	viewerID uint,
	viewedUserID uint,
) (models.ProfileView, error) {
	if repo.saveProfileViewFn == nil {
		panic("unexpected SaveProfileView call")
	}
	return repo.saveProfileViewFn(ctx, viewerID, viewedUserID)
}

func (repo *fakeUserRepository) ListProfileViewers(
	ctx context.Context,
	userID uint,
) ([]models.ProfileViewer, error) {
	if repo.listProfileViewersFn == nil {
		panic("unexpected ListProfileViewers call")
	}
	return repo.listProfileViewersFn(ctx, userID)
}

func (repo *fakeUserRepository) ListProfileLikers(
	ctx context.Context,
	userID uint,
) ([]models.ProfileLiker, error) {
	if repo.listProfileLikersFn == nil {
		panic("unexpected ListProfileLikers call")
	}
	return repo.listProfileLikersFn(ctx, userID)
}

func (repo *fakeUserRepository) BlockUser(
	ctx context.Context,
	blockerID uint,
	blockedUserID uint,
) error {
	if repo.blockUserFn == nil {
		panic("unexpected BlockUser call")
	}
	return repo.blockUserFn(ctx, blockerID, blockedUserID)
}

func (repo *fakeUserRepository) UnblockUser(
	ctx context.Context,
	blockerID uint,
	blockedUserID uint,
) error {
	if repo.unblockUserFn == nil {
		panic("unexpected UnblockUser call")
	}
	return repo.unblockUserFn(ctx, blockerID, blockedUserID)
}

func (repo *fakeUserRepository) HasBlockBetweenUsers(
	ctx context.Context,
	firstUserID uint,
	secondUserID uint,
) (bool, error) {
	if repo.hasBlockBetweenUsersFn == nil {
		panic("unexpected HasBlockBetweenUsers call")
	}
	return repo.hasBlockBetweenUsersFn(ctx, firstUserID, secondUserID)
}

func (repo *fakeUserRepository) ReportUser(
	ctx context.Context,
	reporterID uint,
	reportedUserID uint,
) error {
	if repo.reportUserFn == nil {
		panic("unexpected ReportUser call")
	}
	return repo.reportUserFn(ctx, reporterID, reportedUserID)
}

func (repo *fakeUserRepository) UpdateLastSeen(
	ctx context.Context,
	userID uint,
	lastSeenAt time.Time,
) error {
	if repo.updateLastSeenFn == nil {
		panic("unexpected UpdateLastSeen call")
	}
	return repo.updateLastSeenFn(ctx, userID, lastSeenAt)
}

type fakeImageStorage struct {
	saveAvatarFn   func([]byte, string) (string, error)
	deleteAvatarFn func(string) error
	savePhotoFn    func([]byte, string) (string, error)
	deletePhotoFn  func(string) error
}

type fakeEmailSender struct {
	sendVerificationEmailFn  func(context.Context, string, string) error
	sendPasswordResetEmailFn func(context.Context, string, string) error
}

func (sender *fakeEmailSender) SendPasswordResetEmail(
	ctx context.Context,
	recipientEmail string,
	rawToken string,
) error {
	if sender.sendPasswordResetEmailFn == nil {
		panic("unexpected SendPasswordResetEmail call")
	}
	return sender.sendPasswordResetEmailFn(ctx, recipientEmail, rawToken)
}

func (sender *fakeEmailSender) SendVerificationEmail(
	ctx context.Context,
	recipientEmail string,
	rawToken string,
) error {
	if sender.sendVerificationEmailFn == nil {
		panic("unexpected SendVerificationEmail call")
	}
	return sender.sendVerificationEmailFn(ctx, recipientEmail, rawToken)
}

func (storage *fakeImageStorage) SaveAvatar(data []byte, extension string) (string, error) {
	if storage.saveAvatarFn == nil {
		panic("unexpected SaveAvatar call")
	}
	return storage.saveAvatarFn(data, extension)
}

func (storage *fakeImageStorage) DeleteAvatar(avatarURL string) error {
	if storage.deleteAvatarFn == nil {
		panic("unexpected DeleteAvatar call")
	}
	return storage.deleteAvatarFn(avatarURL)
}

func (storage *fakeImageStorage) SavePhoto(data []byte, extension string) (string, error) {
	if storage.savePhotoFn == nil {
		panic("unexpected SavePhoto call")
	}
	return storage.savePhotoFn(data, extension)
}

func (storage *fakeImageStorage) DeletePhoto(photoURL string) error {
	if storage.deletePhotoFn == nil {
		panic("unexpected DeletePhoto call")
	}
	return storage.deletePhotoFn(photoURL)
}

type fakeUserNotifier struct {
	notifyMatchFn       func(context.Context, uint, uint) (models.Notification, error)
	notifyProfileViewFn func(context.Context, uint, uint) (models.Notification, error)
	notifyLikeFn        func(context.Context, uint, uint) (models.Notification, error)
	notifyUnlikeFn      func(context.Context, uint, uint) (models.Notification, error)
}

type fakeUserPresence struct {
	isUserOnlineFn func(context.Context, uint) bool
}

func (presence *fakeUserPresence) IsUserOnline(ctx context.Context, userID uint) bool {
	if presence.isUserOnlineFn == nil {
		panic("unexpected IsUserOnline call")
	}
	return presence.isUserOnlineFn(ctx, userID)
}

func (notifier *fakeUserNotifier) NotifyMatch(
	ctx context.Context,
	recipientID uint,
	matchedUserID uint,
) (models.Notification, error) {
	if notifier.notifyMatchFn == nil {
		panic("unexpected NotifyMatch call")
	}
	return notifier.notifyMatchFn(ctx, recipientID, matchedUserID)
}

func (notifier *fakeUserNotifier) NotifyProfileView(
	ctx context.Context,
	recipientID uint,
	viewerID uint,
) (models.Notification, error) {
	if notifier.notifyProfileViewFn == nil {
		panic("unexpected NotifyProfileView call")
	}
	return notifier.notifyProfileViewFn(ctx, recipientID, viewerID)
}

func (notifier *fakeUserNotifier) NotifyLike(
	ctx context.Context,
	recipientID uint,
	likerID uint,
) (models.Notification, error) {
	if notifier.notifyLikeFn == nil {
		panic("unexpected NotifyLike call")
	}
	return notifier.notifyLikeFn(ctx, recipientID, likerID)
}

func (notifier *fakeUserNotifier) NotifyUnlike(
	ctx context.Context,
	recipientID uint,
	unlikerID uint,
) (models.Notification, error) {
	if notifier.notifyUnlikeFn == nil {
		panic("unexpected NotifyUnlike call")
	}
	return notifier.notifyUnlikeFn(ctx, recipientID, unlikerID)
}

func validPNGData() []byte {
	imageData := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	imageData.SetNRGBA(0, 0, color.NRGBA{R: 32, G: 96, B: 160, A: 255})

	var buffer bytes.Buffer
	if err := png.Encode(&buffer, imageData); err != nil {
		panic(err)
	}
	return buffer.Bytes()
}
