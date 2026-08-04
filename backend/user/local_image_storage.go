package user

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	avatarsFolder = "avatars"
	photosFolder  = "photos"
)

type LocalImageStorage struct {
	rootDirectory string
	urlPrefix     string
}

func NewLocalImageStorage(rootDirectory string, urlPrefix string) *LocalImageStorage {
	return &LocalImageStorage{rootDirectory: rootDirectory, urlPrefix: strings.TrimRight(urlPrefix, "/")}
}

func isAllowedImageExtension(extension string) bool {
	switch extension {
	case ".jpg", ".png", ".webp":
		return true
	default:
		return false
	}
}

func (storage *LocalImageStorage) SaveAvatar(data []byte, extension string) (string, error) {
	return storage.saveImage(avatarsFolder, data, extension)
}

func (storage *LocalImageStorage) DeleteAvatar(avatarURL string) error {
	return storage.deleteImage(avatarsFolder, avatarURL)
}

func (storage *LocalImageStorage) SavePhoto(data []byte, extension string) (string, error) {
	return storage.saveImage(photosFolder, data, extension)
}

func (storage *LocalImageStorage) DeletePhoto(photoURL string) error {
	return storage.deleteImage(photosFolder, photoURL)
}

func (storage *LocalImageStorage) saveImage(folder string, data []byte, extension string) (string, error) {
	if !isAllowedImageExtension(extension) {
		return "", fmt.Errorf("unsupported image extension: %q", extension)
	}
	fileName, err := generateImageName(extension)
	if err != nil {
		return "", err
	}
	directory := filepath.Join(storage.rootDirectory, folder)
	if err := os.MkdirAll(directory, 0755); err != nil {
		return "", err
	}
	filePath := filepath.Join(directory, fileName)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", err
	}
	imageURL := path.Join(storage.urlPrefix, folder, fileName)
	return imageURL, nil
}

func (storage *LocalImageStorage) deleteImage(folder string, imageURL string) error {
	expectedPrefix := path.Join(storage.urlPrefix, folder) + "/"
	if !strings.HasPrefix(imageURL, expectedPrefix) {
		return nil
	}
	fileName := strings.TrimPrefix(imageURL, expectedPrefix)
	if fileName == "" || path.Base(fileName) != fileName {
		return fmt.Errorf("invalid stored image URL: %q", imageURL)
	}
	filePath := filepath.Join(storage.rootDirectory, folder, fileName)
	err := os.Remove(filePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func generateImageName(extension string) (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes) + extension, nil
}
