package relationship

import (
	"backend/models"
	"context"
)

type Handler struct {
	service  *Service
	notifier UserNotifier
	presence UserPresence
}

type UserPresence interface {
	IsUserOnline(ctx context.Context, userID uint) bool
}

type UserNotifier interface {
	NotifyMatch(ctx context.Context, recipientID uint, matchedUserID uint) (models.Notification, error)
	NotifyProfileView(ctx context.Context, recipientID uint, viewerID uint) (models.Notification, error)
	NotifyLike(ctx context.Context, recipientID uint, likerID uint) (models.Notification, error)
	NotifyUnlike(ctx context.Context, recipientID uint, unlikerID uint) (models.Notification, error)
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) SetUserNotifier(notifier UserNotifier) {
	h.notifier = notifier
}

func (h *Handler) SetUserPresence(presence UserPresence) {
	h.presence = presence
}
