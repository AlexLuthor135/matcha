package user

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalImageStorageSavesAndDeletesAvatar(t *testing.T) {
	rootDirectory := t.TempDir()
	storage := NewLocalImageStorage(rootDirectory, "/uploads/")
	wantData := validPNGData()

	avatarURL, err := storage.SaveAvatar(wantData, ".png")
	if err != nil {
		t.Fatalf("SaveAvatar() unexpected error: %v", err)
	}
	if !strings.HasPrefix(avatarURL, "/uploads/avatars/") || !strings.HasSuffix(avatarURL, ".png") {
		t.Fatalf("SaveAvatar() URL = %q", avatarURL)
	}

	fileName := strings.TrimPrefix(avatarURL, "/uploads/avatars/")
	filePath := filepath.Join(rootDirectory, avatarsFolder, fileName)
	gotData, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read saved avatar: %v", err)
	}
	if string(gotData) != string(wantData) {
		t.Fatalf("saved avatar data = %v, want %v", gotData, wantData)
	}

	if err := storage.DeleteAvatar(avatarURL); err != nil {
		t.Fatalf("DeleteAvatar() unexpected error: %v", err)
	}
	if _, err := os.Stat(filePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("deleted avatar os.Stat() error = %v, want os.ErrNotExist", err)
	}
}

func TestLocalImageStorageRejectsUnsupportedExtension(t *testing.T) {
	storage := NewLocalImageStorage(t.TempDir(), "/uploads")

	if _, err := storage.SavePhoto(validPNGData(), ".gif"); err == nil {
		t.Fatal("SavePhoto() error = nil, want unsupported extension error")
	}
}

func TestLocalImageStorageIgnoresURLsOutsideItsPrefix(t *testing.T) {
	storage := NewLocalImageStorage(t.TempDir(), "/uploads")

	if err := storage.DeletePhoto("https://example.com/photo.png"); err != nil {
		t.Fatalf("DeletePhoto() unexpected error: %v", err)
	}
}
