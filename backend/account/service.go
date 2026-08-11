package account

type Service struct {
	repository  Repository
	emailSender EmailSender
}

func NewService(repository Repository, emailSender EmailSender) *Service {
	return &Service{repository: repository, emailSender: emailSender}
}

func (s *Service) SetEmailSender(emailSender EmailSender) {
	s.emailSender = emailSender
}
