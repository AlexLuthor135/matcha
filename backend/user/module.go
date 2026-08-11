package user

import "database/sql"

type Module struct {
	Handler *UserHandler
	Service *Service
}

func NewModule(db *sql.DB, emailSender EmailSender) *Module {
	imageStorage := NewLocalImageStorage("./uploads", "/uploads")

	repository := NewPostgresRepository(db)
	service := NewService(repository, imageStorage)
	service.SetEmailSender(emailSender)
	handler := NewUserHandler(service)

	return &Module{
		Handler: handler,
		Service: service,
	}
}
