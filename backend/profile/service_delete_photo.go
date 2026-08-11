package profile

import (
	"context"
	"log"
)

func (s *Service) DeletePhoto(ctx context.Context, userID uint, photoID uint) error {
	if photoID == 0 {
		return ProfileErrors.InvalidPhotoID
	}
	photoURL, err := s.repository.DeletePhoto(ctx, userID, photoID)
	if err != nil {
		return err
	}
	if err := s.imageStorage.DeletePhoto(photoURL); err != nil {
		log.Printf("Delete photo file %q for user %d: %v", photoURL, userID, err)
	}
	return nil
}
