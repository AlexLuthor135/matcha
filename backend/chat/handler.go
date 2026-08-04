package chat

type ChatHandler struct {
	service *Service
}

func NewChatHandler(service *Service) *ChatHandler {
	return &ChatHandler{
		service: service,
	}
}
