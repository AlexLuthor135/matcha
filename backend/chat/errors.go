package chat

import "errors"

var ChatErrors = struct {
	SenderNotFound       error
	RecipientNotFound    error
	CannotMessageSelf    error
	MessageBlank         error
	MessageTooLong       error
	MessageNotFound      error
	ConversationNotFound error
}{
	SenderNotFound:       errors.New("sender not found"),
	RecipientNotFound:    errors.New("recipient not found"),
	CannotMessageSelf:    errors.New("cannot send message to yourself"),
	MessageBlank:         errors.New("message cannot be blank"),
	MessageTooLong:       errors.New("message is too long"),
	MessageNotFound:      errors.New("message not found"),
	ConversationNotFound: errors.New("conversation not found"),
}
