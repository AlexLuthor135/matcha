package notification

import (
	"backend/models"
	"context"
)

type fakeNotificationRepository struct {
	Repository
	listNotificationsFn    func(context.Context, uint, uint) ([]models.Notification, error)
	markNotificationReadFn func(context.Context, uint, uint) (models.NotificationReadReceipt, error)
	createNotificationFn   func(context.Context, CreateNotificationInput) (models.Notification, error)
}

func (repo *fakeNotificationRepository) ListNotifications(
	ctx context.Context,
	userID uint,
	limit uint,
) ([]models.Notification, error) {
	if repo.listNotificationsFn == nil {
		panic("unexpected ListNotifications call")
	}
	return repo.listNotificationsFn(ctx, userID, limit)
}

func (repo *fakeNotificationRepository) MarkNotificationRead(
	ctx context.Context,
	userID uint,
	notificationID uint,
) (models.NotificationReadReceipt, error) {
	if repo.markNotificationReadFn == nil {
		panic("unexpected MarkNotificationRead call")
	}
	return repo.markNotificationReadFn(ctx, userID, notificationID)
}

func (repo *fakeNotificationRepository) CreateNotification(
	ctx context.Context,
	input CreateNotificationInput,
) (models.Notification, error) {
	if repo.createNotificationFn == nil {
		panic("unexpected CreateNotification call")
	}
	return repo.createNotificationFn(ctx, input)
}

type fakePublisher struct {
	sendToUserFn func(uint, []byte)
}

func (publisher *fakePublisher) SendToUser(userID uint, message []byte) {
	if publisher.sendToUserFn == nil {
		panic("unexpected SendToUser call")
	}
	publisher.sendToUserFn(userID, message)
}
