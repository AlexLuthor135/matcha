package user

import "context"

type EmailSender interface {
	SendVerificationEmail(ctx context.Context, recipientEmail string, rawToken string) error
	SendPasswordResetEmail(ctx context.Context, recipientEmail string, rawToken string) error
}
