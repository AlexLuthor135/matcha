package account

import (
	"backend/middleware"
	"context"
	"net/http"
)

func authenticatedUserRequest(request *http.Request, userID uint) *http.Request {
	ctx := context.WithValue(request.Context(), middleware.UserIDKey, userID)
	return request.WithContext(ctx)
}
