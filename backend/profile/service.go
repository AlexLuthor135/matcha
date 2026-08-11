package profile

type Service struct {
	repository   Repository
	imageStorage ImageStorage
}

func NewService(repository Repository, imageStorage ImageStorage) *Service {
	return &Service{repository: repository, imageStorage: imageStorage}
}
