package relationship

import (
	"backend/models"
	"context"
	"errors"
	"testing"
)

func TestServiceSaveProfileDecisionRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name         string
		userID       uint
		targetUserID uint
		decision     models.ProfileDecisionValue
		wantErr      error
	}{
		{
			name:         "zero target user ID",
			userID:       10,
			targetUserID: 0,
			decision:     models.ProfileDecisionLike,
			wantErr:      RelationshipErrors.InvalidTargetUserID,
		},
		{
			name:         "own profile",
			userID:       10,
			targetUserID: 10,
			decision:     models.ProfileDecisionLike,
			wantErr:      RelationshipErrors.CannotDecideOwnProfile,
		},
		{
			name:         "unknown decision",
			userID:       10,
			targetUserID: 20,
			decision:     models.ProfileDecisionValue("maybe"),
			wantErr:      RelationshipErrors.InvalidProfileDecision,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(&fakeUserRepository{})

			_, err := service.SaveProfileDecision(
				context.Background(),
				test.userID,
				test.targetUserID,
				test.decision,
			)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("SaveProfileDecision() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestServiceSaveProfileDecisionDelegatesToRepository(t *testing.T) {
	const (
		userID       uint = 10
		targetUserID uint = 20
	)
	wantDecision := models.ProfileDecision{
		ID:           30,
		UserID:       userID,
		TargetUserID: targetUserID,
		Decision:     models.ProfileDecisionLike,
	}

	repository := &fakeUserRepository{
		getAvatarURLFn: func(_ context.Context, gotUserID uint) (string, error) {
			if gotUserID != userID {
				t.Fatalf("GetAvatarURL() userID = %d, want %d", gotUserID, userID)
			}
			return "/uploads/avatars/user.jpg", nil
		},
		saveDecisionFn: func(
			_ context.Context,
			gotUserID uint,
			gotTargetUserID uint,
			gotDecision models.ProfileDecisionValue,
		) (SaveProfileDecisionResult, error) {
			if gotUserID != userID || gotTargetUserID != targetUserID || gotDecision != models.ProfileDecisionLike {
				t.Fatalf(
					"SaveProfileDecision() arguments = (%d, %d, %q), want (%d, %d, %q)",
					gotUserID,
					gotTargetUserID,
					gotDecision,
					userID,
					targetUserID,
					models.ProfileDecisionLike,
				)
			}
			return SaveProfileDecisionResult{
				ProfileDecision: wantDecision,
				IsMatch:         true,
				DecisionChanged: true,
				MatchEnded:      true,
			}, nil
		},
	}

	result, err := NewService(repository).SaveProfileDecision(
		context.Background(),
		userID,
		targetUserID,
		models.ProfileDecisionLike,
	)
	if err != nil {
		t.Fatalf("SaveProfileDecision() unexpected error: %v", err)
	}
	if result.ProfileDecision != wantDecision {
		t.Fatalf("SaveProfileDecision() decision = %+v, want %+v", result.ProfileDecision, wantDecision)
	}
	if !result.IsMatch {
		t.Fatal("SaveProfileDecision() isMatch = false, want true")
	}
	if !result.DecisionChanged {
		t.Fatal("SaveProfileDecision() decisionChanged = false, want true")
	}
	if !result.MatchEnded {
		t.Fatal("SaveProfileDecision() matchEnded = false, want true")
	}
}

func TestServiceSaveProfileDecisionPropagatesRepositoryError(t *testing.T) {
	repositoryError := errors.New("database unavailable")
	repository := &fakeUserRepository{
		getAvatarURLFn: func(context.Context, uint) (string, error) {
			return "/uploads/avatars/user.jpg", nil
		},
		saveDecisionFn: func(
			context.Context,
			uint,
			uint,
			models.ProfileDecisionValue,
		) (SaveProfileDecisionResult, error) {
			return SaveProfileDecisionResult{}, repositoryError
		},
	}

	_, err := NewService(repository).SaveProfileDecision(
		context.Background(),
		10,
		20,
		models.ProfileDecisionLike,
	)
	if !errors.Is(err, repositoryError) {
		t.Fatalf("SaveProfileDecision() error = %v, want %v", err, repositoryError)
	}
}

func TestServiceSaveProfileDecisionRequiresPictureForEveryDecision(t *testing.T) {
	tests := []struct {
		name     string
		decision models.ProfileDecisionValue
	}{
		{
			name:     "like",
			decision: models.ProfileDecisionLike,
		},
		{
			name:     "dislike",
			decision: models.ProfileDecisionDislike,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeUserRepository{
				getAvatarURLFn: func(context.Context, uint) (string, error) {
					return "   ", nil
				},
			}

			_, err := NewService(repository).SaveProfileDecision(
				context.Background(),
				10,
				20,
				test.decision,
			)
			if !errors.Is(err, RelationshipErrors.ProfilePictureRequired) {
				t.Fatalf(
					"SaveProfileDecision() error = %v, want %v",
					err,
					RelationshipErrors.ProfilePictureRequired,
				)
			}
		})
	}
}

func TestServiceSaveProfileDecisionPropagatesAvatarLookupError(t *testing.T) {
	repositoryError := errors.New("avatar lookup failed")
	repository := &fakeUserRepository{
		getAvatarURLFn: func(context.Context, uint) (string, error) {
			return "", repositoryError
		},
	}

	_, err := NewService(repository).SaveProfileDecision(
		context.Background(),
		10,
		20,
		models.ProfileDecisionDislike,
	)
	if !errors.Is(err, repositoryError) {
		t.Fatalf("SaveProfileDecision() error = %v, want %v", err, repositoryError)
	}
}
