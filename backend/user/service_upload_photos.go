package user

import (
	"backend/models"
	"context"
	"log"
)

const (
	maxPhotoFileSize = 5 << 20
	maxProfilePhotos = 4
)

func (s *Service) UploadPhotos(ctx context.Context, userID uint, filesData [][]byte) ([]models.Photo, error) {
	if len(filesData) == 0 {
		return nil, UserErrors.PhotosMissing
	}
	if len(filesData) > maxProfilePhotos {
		return nil, UserErrors.PhotoLimitExceeded
	}
	extensions := make([]string, len(filesData))
	for index, fileData := range filesData {
		if len(fileData) == 0 {
			return nil, UserErrors.PhotoEmpty
		}
		if len(fileData) > maxPhotoFileSize {
			return nil, UserErrors.PhotoTooLarge
		}
		extension, allowed := validatedImageExtension(fileData)
		if !allowed {
			return nil, UserErrors.PhotoTypeUnsupported
		}
		extensions[index] = extension
	}
	photoURLs := make([]string, 0, len(filesData))
	for index, fileData := range filesData {
		photoURL, err := s.imageStorage.SavePhoto(fileData, extensions[index])
		if err != nil {
			s.deleteStoredPhotos(photoURLs)
			return nil, err
		}
		photoURLs = append(photoURLs, photoURL)
	}
	photos, err := s.repository.CreatePhotos(ctx, userID, photoURLs, maxProfilePhotos)
	if err != nil {
		s.deleteStoredPhotos(photoURLs)
		return nil, err
	}
	return photos, nil
}

func (s *Service) deleteStoredPhotos(photoURLs []string) {
	for _, photoURL := range photoURLs {
		if err := s.imageStorage.DeletePhoto(photoURL); err != nil {
			log.Printf("Delete unused photo %q: %v", photoURL, err)
		}
	}
}
