package user

type Service struct {
	repository   Repository
	imageStorage ImageStorage
	emailSender  EmailSender
}

func NewService(repository Repository, imageStorage ImageStorage) *Service {
	return &Service{
		repository:   repository,
		imageStorage: imageStorage,
	}
}

func (s *Service) SetEmailSender(emailSender EmailSender) {
	s.emailSender = emailSender
}
