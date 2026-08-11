package user

import (
	"backend/models"
	"context"
	"errors"
	"testing"
	"time"
)

func TestServiceRecordProfileViewRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name         string
		viewerID     uint
		viewedUserID uint
		wantErr      error
	}{
		{
			name:         "missing viewed user ID",
			viewerID:     10,
			viewedUserID: 0,
			wantErr:      UserErrors.InvalidTargetUserID,
		},
		{
			name:         "own profile",
			viewerID:     10,
			viewedUserID: 10,
			wantErr:      UserErrors.CannotViewOwnProfile,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(&fakeUserRepository{}, &fakeImageStorage{})

			_, err := service.RecordProfileView(
				context.Background(),
				test.viewerID,
				test.viewedUserID,
			)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("RecordProfileView() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestServiceRecordProfileViewMapsMissingTargetUser(t *testing.T) {
	repository := &fakeUserRepository{
		getCompletionStatusFn: func(context.Context, uint) (bool, error) {
			return false, UserErrors.UserNotFound
		},
	}

	_, err := NewService(repository, &fakeImageStorage{}).RecordProfileView(
		context.Background(),
		10,
		20,
	)
	if !errors.Is(err, UserErrors.TargetUserNotFound) {
		t.Fatalf("RecordProfileView() error = %v, want %v", err, UserErrors.TargetUserNotFound)
	}
}

func TestServiceRecordProfileViewRejectsIncompleteTarget(t *testing.T) {
	repository := &fakeUserRepository{
		getCompletionStatusFn: func(_ context.Context, userID uint) (bool, error) {
			if userID != 20 {
				t.Fatalf("GetCompletionStatus() userID = %d, want 20", userID)
			}
			return false, nil
		},
	}

	_, err := NewService(repository, &fakeImageStorage{}).RecordProfileView(
		context.Background(),
		10,
		20,
	)
	if !errors.Is(err, UserErrors.TargetUserNotFound) {
		t.Fatalf("RecordProfileView() error = %v, want %v", err, UserErrors.TargetUserNotFound)
	}
}

func TestServiceRecordProfileViewPropagatesCompletionCheckError(t *testing.T) {
	repositoryError := errors.New("database unavailable")
	repository := &fakeUserRepository{
		getCompletionStatusFn: func(context.Context, uint) (bool, error) {
			return false, repositoryError
		},
	}

	_, err := NewService(repository, &fakeImageStorage{}).RecordProfileView(
		context.Background(),
		10,
		20,
	)
	if !errors.Is(err, repositoryError) {
		t.Fatalf("RecordProfileView() error = %v, want %v", err, repositoryError)
	}
}

func TestServiceRecordProfileViewPropagatesSaveError(t *testing.T) {
	repositoryError := errors.New("database unavailable")
	repository := &fakeUserRepository{
		getCompletionStatusFn: func(context.Context, uint) (bool, error) {
			return true, nil
		},
		saveProfileViewFn: func(context.Context, uint, uint) (models.ProfileView, error) {
			return models.ProfileView{}, repositoryError
		},
	}

	_, err := NewService(repository, &fakeImageStorage{}).RecordProfileView(
		context.Background(),
		10,
		20,
	)
	if !errors.Is(err, repositoryError) {
		t.Fatalf("RecordProfileView() error = %v, want %v", err, repositoryError)
	}
}

func TestServiceRecordProfileViewReturnsSavedView(t *testing.T) {
	createdAt := time.Date(2026, time.August, 6, 12, 0, 0, 0, time.UTC)
	wantView := models.ProfileView{
		ID:           30,
		CreatedAt:    createdAt,
		UpdatedAt:    createdAt,
		ViewerID:     10,
		ViewedUserID: 20,
	}
	repository := &fakeUserRepository{
		getCompletionStatusFn: func(_ context.Context, userID uint) (bool, error) {
			if userID != wantView.ViewedUserID {
				t.Fatalf("GetCompletionStatus() userID = %d, want %d", userID, wantView.ViewedUserID)
			}
			return true, nil
		},
		saveProfileViewFn: func(_ context.Context, viewerID uint, viewedUserID uint) (models.ProfileView, error) {
			if viewerID != wantView.ViewerID || viewedUserID != wantView.ViewedUserID {
				t.Fatalf(
					"SaveProfileView() arguments = (%d, %d), want (%d, %d)",
					viewerID,
					viewedUserID,
					wantView.ViewerID,
					wantView.ViewedUserID,
				)
			}
			return wantView, nil
		},
	}

	profileView, err := NewService(repository, &fakeImageStorage{}).RecordProfileView(
		context.Background(),
		wantView.ViewerID,
		wantView.ViewedUserID,
	)
	if err != nil {
		t.Fatalf("RecordProfileView() unexpected error: %v", err)
	}
	if profileView != wantView {
		t.Fatalf("RecordProfileView() = %+v, want %+v", profileView, wantView)
	}
}
