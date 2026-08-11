package user

import (
	"backend/models"
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestServiceUploadPhotosRejectsInvalidInput(t *testing.T) {
	validFiles := make([][]byte, 5)
	for index := range validFiles {
		validFiles[index] = validPNGData()
	}

	tests := []struct {
		name      string
		filesData [][]byte
		wantErr   error
	}{
		{
			name:      "no photos",
			filesData: nil,
			wantErr:   UserErrors.PhotosMissing,
		},
		{
			name:      "too many photos",
			filesData: validFiles,
			wantErr:   UserErrors.PhotoLimitExceeded,
		},
		{
			name:      "empty photo",
			filesData: [][]byte{nil},
			wantErr:   UserErrors.PhotoEmpty,
		},
		{
			name:      "photo too large",
			filesData: [][]byte{make([]byte, maxPhotoFileSize+1)},
			wantErr:   UserErrors.PhotoTooLarge,
		},
		{
			name:      "unsupported photo type",
			filesData: [][]byte{[]byte("plain text is not an image")},
			wantErr:   UserErrors.PhotoTypeUnsupported,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(&fakeUserRepository{}, &fakeImageStorage{})

			_, err := service.UploadPhotos(context.Background(), 10, test.filesData)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("UploadPhotos() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestServiceUploadPhotosValidatesEveryFileBeforeSaving(t *testing.T) {
	saveCalls := 0
	storage := &fakeImageStorage{
		savePhotoFn: func(_ []byte, _ string) (string, error) {
			saveCalls++
			return "/uploads/photos/unexpected.png", nil
		},
	}

	_, err := NewService(&fakeUserRepository{}, storage).UploadPhotos(
		context.Background(),
		10,
		[][]byte{validPNGData(), []byte("not an image")},
	)
	if !errors.Is(err, UserErrors.PhotoTypeUnsupported) {
		t.Fatalf("UploadPhotos() error = %v, want %v", err, UserErrors.PhotoTypeUnsupported)
	}
	if saveCalls != 0 {
		t.Fatalf("SavePhoto() calls = %d, want 0", saveCalls)
	}
}

func TestServiceUploadPhotosCleansUpWhenFileSaveFails(t *testing.T) {
	storageError := errors.New("storage unavailable")
	firstURL := "/uploads/photos/first.png"
	saveCalls := 0
	var deletedURLs []string

	storage := &fakeImageStorage{
		savePhotoFn: func(_ []byte, extension string) (string, error) {
			if extension != ".png" {
				t.Fatalf("SavePhoto() extension = %q, want .png", extension)
			}
			saveCalls++
			if saveCalls == 1 {
				return firstURL, nil
			}
			return "", storageError
		},
		deletePhotoFn: func(photoURL string) error {
			deletedURLs = append(deletedURLs, photoURL)
			return nil
		},
	}

	_, err := NewService(&fakeUserRepository{}, storage).UploadPhotos(
		context.Background(),
		10,
		[][]byte{validPNGData(), validPNGData()},
	)
	if !errors.Is(err, storageError) {
		t.Fatalf("UploadPhotos() error = %v, want %v", err, storageError)
	}
	if !reflect.DeepEqual(deletedURLs, []string{firstURL}) {
		t.Fatalf("deleted URLs = %v, want [%s]", deletedURLs, firstURL)
	}
}

func TestServiceUploadPhotosCleansUpWhenRepositoryFails(t *testing.T) {
	databaseError := errors.New("database unavailable")
	wantURLs := []string{
		"/uploads/photos/1.png",
		"/uploads/photos/2.png",
	}
	saveIndex := 0
	var deletedURLs []string

	storage := &fakeImageStorage{
		savePhotoFn: func(_ []byte, extension string) (string, error) {
			if extension != ".png" {
				t.Fatalf("SavePhoto() extension = %q, want .png", extension)
			}
			photoURL := wantURLs[saveIndex]
			saveIndex++
			return photoURL, nil
		},
		deletePhotoFn: func(photoURL string) error {
			deletedURLs = append(deletedURLs, photoURL)
			return nil
		},
	}

	repository := &fakeUserRepository{
		createPhotosFn: func(_ context.Context, gotUserID uint, gotURLs []string, gotMax int) ([]models.Photo, error) {
			if gotUserID != 10 || gotMax != maxProfilePhotos || !reflect.DeepEqual(gotURLs, wantURLs) {
				t.Fatalf("CreatePhotos() arguments = (%d, %v, %d)", gotUserID, gotURLs, gotMax)
			}
			return nil, databaseError
		},
	}

	_, err := NewService(repository, storage).UploadPhotos(
		context.Background(),
		10,
		[][]byte{validPNGData(), validPNGData()},
	)
	if !errors.Is(err, databaseError) {
		t.Fatalf("UploadPhotos() error = %v, want %v", err, databaseError)
	}
	if !reflect.DeepEqual(deletedURLs, wantURLs) {
		t.Fatalf("deleted URLs = %v, want %v", deletedURLs, wantURLs)
	}
}

func TestServiceUploadPhotosReturnsCreatedPhotos(t *testing.T) {
	wantURLs := []string{
		"/uploads/photos/1.png",
		"/uploads/photos/2.png",
	}
	wantPhotos := []models.Photo{
		{ID: 1, UserID: 10, URL: wantURLs[0]},
		{ID: 2, UserID: 10, URL: wantURLs[1]},
	}
	saveIndex := 0

	storage := &fakeImageStorage{
		savePhotoFn: func(_ []byte, _ string) (string, error) {
			if saveIndex >= len(wantURLs) {
				return "", fmt.Errorf("unexpected SavePhoto call %d", saveIndex+1)
			}
			photoURL := wantURLs[saveIndex]
			saveIndex++
			return photoURL, nil
		},
	}

	repository := &fakeUserRepository{
		createPhotosFn: func(_ context.Context, _ uint, gotURLs []string, _ int) ([]models.Photo, error) {
			if !reflect.DeepEqual(gotURLs, wantURLs) {
				t.Fatalf("CreatePhotos() URLs = %v, want %v", gotURLs, wantURLs)
			}
			return wantPhotos, nil
		},
	}

	photos, err := NewService(repository, storage).UploadPhotos(
		context.Background(),
		10,
		[][]byte{validPNGData(), validPNGData()},
	)
	if err != nil {
		t.Fatalf("UploadPhotos() unexpected error: %v", err)
	}
	if !reflect.DeepEqual(photos, wantPhotos) {
		t.Fatalf("UploadPhotos() = %+v, want %+v", photos, wantPhotos)
	}
}
