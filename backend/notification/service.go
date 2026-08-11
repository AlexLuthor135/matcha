package notification

const defaultNotificationLimit = 50

type Publisher interface {
	SendToUser(userID uint, message []byte)
}

type Service struct {
	repository Repository
	publisher  Publisher
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) SetPublisher(publisher Publisher) {
	s.publisher = publisher
}
