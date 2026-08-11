package notification

import "database/sql"

type Module struct {
	Handler *NotificationHandler
	Service *Service
}

func NewModule(db *sql.DB) *Module {
	repository := NewPostgresRepository(db)
	service := NewService(repository)
	handler := NewNotificationHandler(service)

	return &Module{
		Handler: handler,
		Service: service,
	}
}
