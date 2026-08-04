package user

type UserHandler struct {
	service *Service
}

func NewUserHandler(service *Service) *UserHandler {
	return &UserHandler{
		service: service,
	}
}
