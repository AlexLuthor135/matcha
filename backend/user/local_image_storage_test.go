package user

import (
	"bytes"
	"errors"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidatedImageExtension(t *testing.T) {
	validPNG := validPNGData()

	tests := []struct {
		name          string
		data          []byte
		wantExtension string
		wantAllowed   bool
	}{
		{
			name:          "valid PNG",
			data:          validPNG,
			wantExtension: ".png",
			wantAllowed:   true,
		},
		{
			name: "PNG signature followed by non-image content",
			data: append(
				[]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'},
				[]byte("<script>not an image</script>")...,
			),
		},
		{
			name: "truncated PNG",
			data: validPNG[:len(validPNG)/2],
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			extension, allowed := validatedImageExtension(test.data)
			if extension != test.wantExtension || allowed != test.wantAllowed {
				t.Fatalf(
					"validatedImageExtension() = (%q, %t), want (%q, %t)",
					extension,
					allowed,
					test.wantExtension,
					test.wantAllowed,
				)
			}
		})
	}
}

func TestValidatedImageExtensionRejectsOversizedDimensions(t *testing.T) {
	oversizedImage := image.NewNRGBA(image.Rect(0, 0, maxImageDimension+1, 1))
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, oversizedImage); err != nil {
		t.Fatalf("encode oversized test image: %v", err)
	}

	if extension, allowed := validatedImageExtension(encoded.Bytes()); allowed || extension != "" {
		t.Fatalf(
			"validatedImageExtension() = (%q, %t), want (\"\", false)",
			extension,
			allowed,
		)
	}
}

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
