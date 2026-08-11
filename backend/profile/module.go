package profile

import "database/sql"

type Module struct {
	Handler *Handler
	Service *Service
}

func NewModule(db *sql.DB) *Module {
	storage := NewLocalImageStorage("./uploads", "/uploads")
	repository := NewPostgresRepository(db)
	service := NewService(repository, storage)
	return &Module{Handler: NewHandler(service), Service: service}
}
