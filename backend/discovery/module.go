package discovery

import "database/sql"

type Module struct {
	Handler *Handler
	Service *Service
}

func NewModule(db *sql.DB) *Module {
	repository := NewPostgresRepository(db)
	service := NewService(repository)
	return &Module{
		Handler: NewHandler(service),
		Service: service,
	}
}
