package relationship

import (
	"backend/models"
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestServiceListMatchesDelegatesToRepository(t *testing.T) {
	const userID uint = 10
	wantMatches := []models.Match{
		{
			ID:        20,
			UserName:  "matched-user",
			FirstName: "Matched",
			LastName:  "User",
			Avatar:    "/uploads/avatars/matched-user.jpg",
		},
	}

	repository := &fakeUserRepository{
		listMatchesFn: func(_ context.Context, gotUserID uint) ([]models.Match, error) {
			if gotUserID != userID {
				t.Fatalf("ListMatches() userID = %d, want %d", gotUserID, userID)
			}
			return wantMatches, nil
		},
	}

	matches, err := NewService(repository).ListMatches(
		context.Background(),
		userID,
	)
	if err != nil {
		t.Fatalf("ListMatches() unexpected error: %v", err)
	}
	if !reflect.DeepEqual(matches, wantMatches) {
		t.Fatalf("ListMatches() = %+v, want %+v", matches, wantMatches)
	}
}

func TestServiceListMatchesPropagatesRepositoryError(t *testing.T) {
	repositoryError := errors.New("database unavailable")
	repository := &fakeUserRepository{
		listMatchesFn: func(context.Context, uint) ([]models.Match, error) {
			return nil, repositoryError
		},
	}

	_, err := NewService(repository).ListMatches(
		context.Background(),
		10,
	)
	if !errors.Is(err, repositoryError) {
		t.Fatalf("ListMatches() error = %v, want %v", err, repositoryError)
	}
}
