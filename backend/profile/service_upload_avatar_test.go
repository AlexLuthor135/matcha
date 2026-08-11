package profile

import (
	"context"
	"errors"
	"testing"
)

func TestServiceUploadAvatarRejectsInvalidFile(t *testing.T) {
	tests := []struct {
		name     string
		fileData []byte
		wantErr  error
	}{
		{
			name:     "empty avatar",
			fileData: nil,
			wantErr:  ProfileErrors.AvatarEmpty,
		},
		{
			name:     "avatar too large",
			fileData: make([]byte, maxAvatarFileSize+1),
			wantErr:  ProfileErrors.AvatarTooLarge,
		},
		{
			name:     "unsupported avatar type",
			fileData: []byte("plain text is not an image"),
			wantErr:  ProfileErrors.AvatarTypeUnsupported,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(&fakeUserRepository{}, &fakeImageStorage{})

			_, err := service.UploadAvatar(context.Background(), 10, test.fileData)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("UploadAvatar() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestServiceUploadAvatarDeletesNewFileWhenDatabaseUpdateFails(t *testing.T) {
	const (
		userID       uint = 10
		oldAvatarURL      = "/uploads/avatars/old.png"
		newAvatarURL      = "/uploads/avatars/new.png"
	)
	databaseError := errors.New("database unavailable")

	repository := &fakeUserRepository{
		getAvatarURLFn: func(_ context.Context, gotUserID uint) (string, error) {
			if gotUserID != userID {
				t.Fatalf("GetAvatarURL() userID = %d, want %d", gotUserID, userID)
			}
			return oldAvatarURL, nil
		},
		updateAvatarURLFn: func(_ context.Context, gotUserID uint, gotURL string) (string, error) {
			if gotUserID != userID || gotURL != newAvatarURL {
				t.Fatalf("UpdateAvatarURL() arguments = (%d, %q), want (%d, %q)", gotUserID, gotURL, userID, newAvatarURL)
			}
			return "", databaseError
		},
	}

	var deletedURLs []string
	storage := &fakeImageStorage{
		saveAvatarFn: func(_ []byte, extension string) (string, error) {
			if extension != ".png" {
				t.Fatalf("SaveAvatar() extension = %q, want .png", extension)
			}
			return newAvatarURL, nil
		},
		deleteAvatarFn: func(avatarURL string) error {
			deletedURLs = append(deletedURLs, avatarURL)
			return nil
		},
	}

	_, err := NewService(repository, storage).UploadAvatar(context.Background(), userID, validPNGData())
	if !errors.Is(err, databaseError) {
		t.Fatalf("UploadAvatar() error = %v, want %v", err, databaseError)
	}
	if len(deletedURLs) != 1 || deletedURLs[0] != newAvatarURL {
		t.Fatalf("deleted URLs = %v, want [%s]", deletedURLs, newAvatarURL)
	}
}

func TestServiceUploadAvatarReplacesOldFile(t *testing.T) {
	const (
		userID       uint = 10
		oldAvatarURL      = "/uploads/avatars/old.png"
		newAvatarURL      = "/uploads/avatars/new.png"
	)

	repository := &fakeUserRepository{
		getAvatarURLFn: func(_ context.Context, gotUserID uint) (string, error) {
			return oldAvatarURL, nil
		},
		updateAvatarURLFn: func(_ context.Context, gotUserID uint, gotURL string) (string, error) {
			if gotUserID != userID || gotURL != newAvatarURL {
				t.Fatalf("UpdateAvatarURL() arguments = (%d, %q), want (%d, %q)", gotUserID, gotURL, userID, newAvatarURL)
			}
			return newAvatarURL, nil
		},
	}

	var deletedURLs []string
	storage := &fakeImageStorage{
		saveAvatarFn: func(_ []byte, extension string) (string, error) {
			if extension != ".png" {
				t.Fatalf("SaveAvatar() extension = %q, want .png", extension)
			}
			return newAvatarURL, nil
		},
		deleteAvatarFn: func(avatarURL string) error {
			deletedURLs = append(deletedURLs, avatarURL)
			return nil
		},
	}

	avatarURL, err := NewService(repository, storage).UploadAvatar(context.Background(), userID, validPNGData())
	if err != nil {
		t.Fatalf("UploadAvatar() unexpected error: %v", err)
	}
	if avatarURL != newAvatarURL {
		t.Fatalf("UploadAvatar() URL = %q, want %q", avatarURL, newAvatarURL)
	}
	if len(deletedURLs) != 1 || deletedURLs[0] != oldAvatarURL {
		t.Fatalf("deleted URLs = %v, want [%s]", deletedURLs, oldAvatarURL)
	}
}
