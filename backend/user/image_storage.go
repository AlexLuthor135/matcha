package user

import (
	"bytes"
	"image"

	_ "golang.org/x/image/webp"
	_ "image/jpeg"
	_ "image/png"
)

const (
	maxImageDimension = 4096
	maxImagePixels    = 4096 * 4096
)

type ImageStorage interface {
	SaveAvatar(data []byte, extension string) (string, error)
	DeleteAvatar(avatarURL string) error
	SavePhoto(data []byte, extension string) (string, error)
	DeletePhoto(photoURL string) error
}

func validatedImageExtension(data []byte) (string, bool) {
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return "", false
	}
	if config.Width < 1 || config.Height < 1 {
		return "", false
	}

	if config.Width > maxImageDimension || config.Height > maxImageDimension {
		return "", false
	}

	if config.Width > maxImagePixels/config.Height {
		return "", false
	}
	_, decodedFormat, err := image.Decode(bytes.NewReader(data))
	if err != nil || decodedFormat != format {
		return "", false
	}
	switch format {
	case "jpeg":
		return ".jpg", true
	case "png":
		return ".png", true
	case "webp":
		return ".webp", true
	default:
		return "", false
	}
}
