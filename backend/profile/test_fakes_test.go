package profile

import (
	"backend/models"
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"time"
)

type fakeUserRepository struct {
	Repository
	getProfileFn      func(context.Context, uint) (models.User, error)
	completeProfileFn func(context.Context, uint, CompleteProfileInput) (bool, error)
	updateProfileFn   func(context.Context, uint, UpdateProfileInput) error
	getAvatarURLFn    func(context.Context, uint) (string, error)
	updateAvatarURLFn func(context.Context, uint, string) (string, error)
	createPhotosFn    func(context.Context, uint, []string, int) ([]models.Photo, error)
	deletePhotoFn     func(context.Context, uint, uint) (string, error)
	updateLastSeenFn  func(context.Context, uint, time.Time) error
}

func (r *fakeUserRepository) GetProfile(ctx context.Context, id uint) (models.User, error) {
	return r.getProfileFn(ctx, id)
}
func (r *fakeUserRepository) CompleteProfile(ctx context.Context, id uint, input CompleteProfileInput) (bool, error) {
	return r.completeProfileFn(ctx, id, input)
}
func (r *fakeUserRepository) UpdateProfile(ctx context.Context, id uint, input UpdateProfileInput) error {
	return r.updateProfileFn(ctx, id, input)
}
func (r *fakeUserRepository) GetAvatarURL(ctx context.Context, id uint) (string, error) {
	return r.getAvatarURLFn(ctx, id)
}
func (r *fakeUserRepository) UpdateAvatarURL(ctx context.Context, id uint, url string) (string, error) {
	return r.updateAvatarURLFn(ctx, id, url)
}
func (r *fakeUserRepository) CreatePhotos(ctx context.Context, id uint, urls []string, max int) ([]models.Photo, error) {
	return r.createPhotosFn(ctx, id, urls, max)
}
func (r *fakeUserRepository) DeletePhoto(ctx context.Context, userID uint, photoID uint) (string, error) {
	return r.deletePhotoFn(ctx, userID, photoID)
}
func (r *fakeUserRepository) UpdateLastSeen(ctx context.Context, id uint, lastSeen time.Time) error {
	return r.updateLastSeenFn(ctx, id, lastSeen)
}

type fakeImageStorage struct {
	saveAvatarFn   func([]byte, string) (string, error)
	deleteAvatarFn func(string) error
	savePhotoFn    func([]byte, string) (string, error)
	deletePhotoFn  func(string) error
}

func (s *fakeImageStorage) SaveAvatar(data []byte, extension string) (string, error) {
	return s.saveAvatarFn(data, extension)
}
func (s *fakeImageStorage) DeleteAvatar(url string) error { return s.deleteAvatarFn(url) }
func (s *fakeImageStorage) SavePhoto(data []byte, extension string) (string, error) {
	return s.savePhotoFn(data, extension)
}
func (s *fakeImageStorage) DeletePhoto(url string) error { return s.deletePhotoFn(url) }

func validPNGData() []byte {
	imageData := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	imageData.SetNRGBA(0, 0, color.NRGBA{R: 32, G: 96, B: 160, A: 255})
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, imageData); err != nil {
		panic(err)
	}
	return buffer.Bytes()
}
