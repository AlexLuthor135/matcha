package user

import (
	"backend/models"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"
)

const authSessionLifeTime = 7 * 24 * time.Hour

func generateAuthSessionID() (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

func (s *Service) CreateAuthSession(ctx context.Context, userID uint) (models.AuthSession, error) {
	sessionID, err := generateAuthSessionID()
	if err != nil {
		return models.AuthSession{}, err
	}
	session := models.AuthSession{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: time.Now().UTC().Add(authSessionLifeTime),
	}
	if err := s.repository.CreateAuthSession(ctx, session); err != nil {
		return models.AuthSession{}, err
	}
	return session, nil
}

func (s *Service) RevokeAuthSession(ctx context.Context, sessionID string, userID uint) error {
	return s.repository.RevokeAuthSession(ctx, sessionID, userID)
}

func (s *Service) ValidateAuthSession(ctx context.Context, sessionID string, userID uint) (bool, error) {
	err := s.repository.UseAuthSession(ctx, sessionID, userID)
	if errors.Is(err, UserErrors.InvalidAuthSession) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
