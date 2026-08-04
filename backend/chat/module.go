package chat

import "database/sql"

type Module struct {
	Handler *ChatHandler
	Service *Service
}

func NewModule(db *sql.DB) *Module {
	repository := NewPostgresRepository(db)
	service := NewService(repository)
	handler := NewChatHandler(service)

	return &Module{
		Handler: handler,
		Service: service,
	}
}
