package user

type ImageStorage interface {
	SaveAvatar(data []byte, extension string) (string, error)
	DeleteAvatar(avatarURL string) error
	SavePhoto(data []byte, extension string) (string, error)
	DeletePhoto(photoURL string) error
}

func imageExtension(contentType string) (string, bool) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/webp":
		return ".webp", true
	default:
		return "", false
	}
}
