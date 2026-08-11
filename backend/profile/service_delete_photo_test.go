package profile

import (
	"context"
	"errors"
	"testing"
)

func TestServiceDeletePhotoRejectsZeroID(t *testing.T) {
	service := NewService(&fakeUserRepository{}, &fakeImageStorage{})

	err := service.DeletePhoto(context.Background(), 10, 0)
	if !errors.Is(err, ProfileErrors.InvalidPhotoID) {
		t.Fatalf("DeletePhoto() error = %v, want %v", err, ProfileErrors.InvalidPhotoID)
	}
}

func TestServiceDeletePhotoDoesNotDeleteFileWhenRepositoryFails(t *testing.T) {
	repositoryError := errors.New("database unavailable")
	repository := &fakeUserRepository{
		deletePhotoFn: func(_ context.Context, _ uint, _ uint) (string, error) {
			return "", repositoryError
		},
	}

	err := NewService(repository, &fakeImageStorage{}).DeletePhoto(context.Background(), 10, 20)
	if !errors.Is(err, repositoryError) {
		t.Fatalf("DeletePhoto() error = %v, want %v", err, repositoryError)
	}
}

func TestServiceDeletePhotoRemovesStoredFile(t *testing.T) {
	const photoURL = "/uploads/photos/photo.png"
	repository := &fakeUserRepository{
		deletePhotoFn: func(_ context.Context, gotUserID uint, gotPhotoID uint) (string, error) {
			if gotUserID != 10 || gotPhotoID != 20 {
				t.Fatalf("DeletePhoto() arguments = (%d, %d), want (10, 20)", gotUserID, gotPhotoID)
			}
			return photoURL, nil
		},
	}

	var deletedURL string
	storage := &fakeImageStorage{
		deletePhotoFn: func(photoURL string) error {
			deletedURL = photoURL
			return nil
		},
	}

	if err := NewService(repository, storage).DeletePhoto(context.Background(), 10, 20); err != nil {
		t.Fatalf("DeletePhoto() unexpected error: %v", err)
	}
	if deletedURL != photoURL {
		t.Fatalf("DeletePhoto() deleted URL = %q, want %q", deletedURL, photoURL)
	}
}
