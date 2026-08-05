package user

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
			wantErr:      UserErrors.InvalidTargetUserID,
		},
		{
			name:         "own profile",
			userID:       10,
			targetUserID: 10,
			decision:     models.ProfileDecisionLike,
			wantErr:      UserErrors.CannotDecideOwnProfile,
		},
		{
			name:         "unknown decision",
			userID:       10,
			targetUserID: 20,
			decision:     models.ProfileDecisionValue("maybe"),
			wantErr:      UserErrors.InvalidProfileDecision,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(&fakeUserRepository{}, &fakeImageStorage{})

			_, _, err := service.SaveProfileDecision(
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
		saveDecisionFn: func(
			_ context.Context,
			gotUserID uint,
			gotTargetUserID uint,
			gotDecision models.ProfileDecisionValue,
		) (models.ProfileDecision, bool, error) {
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
			return wantDecision, true, nil
		},
	}

	decision, isMatch, err := NewService(repository, &fakeImageStorage{}).SaveProfileDecision(
		context.Background(),
		userID,
		targetUserID,
		models.ProfileDecisionLike,
	)
	if err != nil {
		t.Fatalf("SaveProfileDecision() unexpected error: %v", err)
	}
	if decision != wantDecision {
		t.Fatalf("SaveProfileDecision() decision = %+v, want %+v", decision, wantDecision)
	}
	if !isMatch {
		t.Fatal("SaveProfileDecision() isMatch = false, want true")
	}
}

func TestServiceSaveProfileDecisionPropagatesRepositoryError(t *testing.T) {
	repositoryError := errors.New("database unavailable")
	repository := &fakeUserRepository{
		saveDecisionFn: func(
			context.Context,
			uint,
			uint,
			models.ProfileDecisionValue,
		) (models.ProfileDecision, bool, error) {
			return models.ProfileDecision{}, false, repositoryError
		},
	}

	_, _, err := NewService(repository, &fakeImageStorage{}).SaveProfileDecision(
		context.Background(),
		10,
		20,
		models.ProfileDecisionLike,
	)
	if !errors.Is(err, repositoryError) {
		t.Fatalf("SaveProfileDecision() error = %v, want %v", err, repositoryError)
	}
}
