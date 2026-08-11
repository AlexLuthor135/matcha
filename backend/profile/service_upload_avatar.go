package profile

import (
	"context"
	"log"
)

const maxAvatarFileSize = 5 << 20

func (s *Service) UploadAvatar(ctx context.Context, userID uint, fileData []byte) (string, error) {
	if len(fileData) == 0 {
		return "", ProfileErrors.AvatarEmpty
	}
	if len(fileData) > maxAvatarFileSize {
		return "", ProfileErrors.AvatarTooLarge
	}
	extension, allowed := validatedImageExtension(fileData)
	if !allowed {
		return "", ProfileErrors.AvatarTypeUnsupported
	}
	oldAvatarURL, err := s.repository.GetAvatarURL(ctx, userID)
	if err != nil {
		return "", err
	}
	newAvatarURL, err := s.imageStorage.SaveAvatar(fileData, extension)
	if err != nil {
		return "", err
	}
	savedAvatarURL, err := s.repository.UpdateAvatarURL(ctx, userID, newAvatarURL)
	if err != nil {
		if deleteErr := s.imageStorage.DeleteAvatar(newAvatarURL); deleteErr != nil {
			log.Printf("Delete unused avatar %q: %v", newAvatarURL, deleteErr)
		}
		return "", err
	}
	if oldAvatarURL != "" && oldAvatarURL != savedAvatarURL {
		if err := s.imageStorage.DeleteAvatar(oldAvatarURL); err != nil {
			log.Printf("Delete old avatar %q for user %d: %v", oldAvatarURL, userID, err)
		}
	}
	return savedAvatarURL, nil
}
