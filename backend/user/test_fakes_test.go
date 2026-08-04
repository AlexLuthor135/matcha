package user

import (
	"backend/models"
	"context"
)

type fakeUserRepository struct {
	Repository
	getAvatarURLFn    func(context.Context, uint) (string, error)
	updateAvatarURLFn func(context.Context, uint, string) (string, error)
	createPhotosFn    func(context.Context, uint, []string, int) ([]models.Photo, error)
	deletePhotoFn     func(context.Context, uint, uint) (string, error)
}

func (repo *fakeUserRepository) GetAvatarURL(ctx context.Context, userID uint) (string, error) {
	if repo.getAvatarURLFn == nil {
		panic("unexpected GetAvatarURL call")
	}
	return repo.getAvatarURLFn(ctx, userID)
}

func (repo *fakeUserRepository) UpdateAvatarURL(ctx context.Context, userID uint, avatarURL string) (string, error) {
	if repo.updateAvatarURLFn == nil {
		panic("unexpected UpdateAvatarURL call")
	}
	return repo.updateAvatarURLFn(ctx, userID, avatarURL)
}

func (repo *fakeUserRepository) CreatePhotos(ctx context.Context, userID uint, photoURLs []string, maxAllowed int) ([]models.Photo, error) {
	if repo.createPhotosFn == nil {
		panic("unexpected CreatePhotos call")
	}
	return repo.createPhotosFn(ctx, userID, photoURLs, maxAllowed)
}

func (repo *fakeUserRepository) DeletePhoto(ctx context.Context, userID uint, photoID uint) (string, error) {
	if repo.deletePhotoFn == nil {
		panic("unexpected DeletePhoto call")
	}
	return repo.deletePhotoFn(ctx, userID, photoID)
}

type fakeImageStorage struct {
	saveAvatarFn   func([]byte, string) (string, error)
	deleteAvatarFn func(string) error
	savePhotoFn    func([]byte, string) (string, error)
	deletePhotoFn  func(string) error
}

func (storage *fakeImageStorage) SaveAvatar(data []byte, extension string) (string, error) {
	if storage.saveAvatarFn == nil {
		panic("unexpected SaveAvatar call")
	}
	return storage.saveAvatarFn(data, extension)
}

func (storage *fakeImageStorage) DeleteAvatar(avatarURL string) error {
	if storage.deleteAvatarFn == nil {
		panic("unexpected DeleteAvatar call")
	}
	return storage.deleteAvatarFn(avatarURL)
}

func (storage *fakeImageStorage) SavePhoto(data []byte, extension string) (string, error) {
	if storage.savePhotoFn == nil {
		panic("unexpected SavePhoto call")
	}
	return storage.savePhotoFn(data, extension)
}

func (storage *fakeImageStorage) DeletePhoto(photoURL string) error {
	if storage.deletePhotoFn == nil {
		panic("unexpected DeletePhoto call")
	}
	return storage.deletePhotoFn(photoURL)
}

func validPNGData() []byte {
	return []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0x00, 0x00, 0x00, 0x0d}
}
