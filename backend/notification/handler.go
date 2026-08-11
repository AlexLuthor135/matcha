package notification

type NotificationHandler struct {
	service *Service
}

func NewNotificationHandler(service *Service) *NotificationHandler {
	return &NotificationHandler{
		service: service,
	}
}
