package user

import (
	"context"
	"net"
	"net/mail"
	"net/smtp"
	"net/url"
	"strings"
)

type SMTPConfig struct {
	Host              string
	Port              string
	UserName          string
	Password          string
	From              string
	PublicBackendURL  string
	PublicFrontendURL string
}

type SMTPEmailSender struct {
	config SMTPConfig
}

func (config *SMTPConfig) Normalize() {
	config.Host = strings.TrimSpace(config.Host)
	config.Port = strings.TrimSpace(config.Port)
	config.From = strings.TrimSpace(config.From)
	config.PublicBackendURL = strings.TrimRight(strings.TrimSpace(config.PublicBackendURL), "/")
	config.PublicFrontendURL = strings.TrimRight(strings.TrimSpace(config.PublicFrontendURL), "/")
}

func (config SMTPConfig) HasMissingFields() bool {
	return config.Host == "" || config.Port == "" || config.From == "" || config.PublicBackendURL == "" || config.PublicFrontendURL == ""
}

func NewSMTPEmailSender(config SMTPConfig) (*SMTPEmailSender, error) {
	config.Normalize()
	if config.HasMissingFields() {
		return nil, UserErrors.SMTPConfigInvalid
	}
	if _, err := mail.ParseAddress(config.From); err != nil {
		return nil, UserErrors.SMTPAddressInvalid
	}
	return &SMTPEmailSender{config: config}, nil
}

func (sender *SMTPEmailSender) sendTextEmail(ctx context.Context, recipientEmail string, subject string, body string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	recipient, err := mail.ParseAddress(recipientEmail)
	if err != nil {
		return UserErrors.SMTPAddressInvalid
	}
	message := []byte(
		"From: " + sender.config.From + "\r\n" +
			"To: " + recipient.Address + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n" +
			"\r\n" +
			body + "\r\n",
	)
	var auth smtp.Auth
	if sender.config.UserName != "" {
		auth = smtp.PlainAuth("", sender.config.UserName, sender.config.Password, sender.config.Host)
	}
	smtpAddress := net.JoinHostPort(sender.config.Host, sender.config.Port)
	return smtp.SendMail(smtpAddress, auth, sender.config.From, []string{recipient.Address}, message)
}

func (sender *SMTPEmailSender) SendVerificationEmail(ctx context.Context, recipientEmail string, rawToken string) error {
	verificationURL := sender.config.PublicBackendURL + "/api/verify-email?token=" + url.QueryEscape(rawToken)
	body := "Welcome to Matcha!\r\n\r\n" + "Open this link to verify your email:\r\n" + verificationURL
	return sender.sendTextEmail(ctx, recipientEmail, "Verify your Matcha email", body)
}

func (sender *SMTPEmailSender) SendPasswordResetEmail(ctx context.Context, recipientEmail string, rawToken string) error {
	resetURL := sender.config.PublicFrontendURL + "/reset-password?token=" + url.QueryEscape(rawToken)
	body := "A password reset was requested for your Matcha account.\r\n\r\n" + "Open this link to choose a new password:\r\n" + resetURL + "\r\n\r\n" + "If you did not request this, ignore this email."
	return sender.sendTextEmail(ctx, recipientEmail, "Reset your Matcha password", body)
}
