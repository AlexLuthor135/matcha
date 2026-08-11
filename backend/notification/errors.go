package notification

import "errors"

var NotificationErrors = struct {
	NotificationNotFound  error
	InvalidUserID         error
	InvalidSenderID       error
	TypeBlank             error
	MessageBlank          error
	UserBlocked           error
	InvalidMessageID      error
	InvalidConversationID error
}{
	NotificationNotFound:  errors.New("notification not found"),
	InvalidUserID:         errors.New("invalid notification user id"),
	InvalidSenderID:       errors.New("invalid notification sender id"),
	TypeBlank:             errors.New("notification type cannot be blank"),
	MessageBlank:          errors.New("notification message cannot be blank"),
	UserBlocked:           errors.New("notification users are blocked"),
	InvalidMessageID:      errors.New("invalid message id"),
	InvalidConversationID: errors.New("invalid conversation id"),
}
