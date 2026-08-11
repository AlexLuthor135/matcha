package account

import "database/sql"

type Module struct {
	Handler *Handler
	Service *Service
}

func NewModule(db *sql.DB, emailSender EmailSender) *Module {
	repository := NewPostgresRepository(db)
	service := NewService(repository, emailSender)
	return &Module{Handler: NewHandler(service), Service: service}
}
