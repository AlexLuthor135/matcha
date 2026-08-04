package user

import "database/sql"

type Module struct {
	Handler *UserHandler
}

func NewModule(db *sql.DB) *Module {
	imageStorage := NewLocalImageStorage("./uploads", "/uploads")

	repository := NewPostgresRepository(db)
	service := NewService(repository, imageStorage)
	handler := NewUserHandler(service)

	return &Module{
		Handler: handler,
	}
}
