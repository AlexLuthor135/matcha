package user

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func userBlockRequest(method string, blockerID uint, targetUserID string) *http.Request {
	request := httptest.NewRequest(method, "/profiles/"+targetUserID+"/block", nil)
	request.SetPathValue("targetUserID", targetUserID)
	return authenticatedUserRequest(request, blockerID)
}

func TestUserBlockHandlersRejectInvalidRequests(t *testing.T) {
	tests := []struct {
		name       string
		request    *http.Request
		call       func(*UserHandler, http.ResponseWriter, *http.Request)
		wantStatus int
		wantBody   string
	}{
		{
			name:       "block unauthorized",
			request:    httptest.NewRequest(http.MethodPut, "/profiles/20/block", nil),
			call:       (*UserHandler).BlockUser,
			wantStatus: http.StatusUnauthorized,
			wantBody:   "Unauthorized\n",
		},
		{
			name:       "unblock unauthorized",
			request:    httptest.NewRequest(http.MethodDelete, "/profiles/20/block", nil),
			call:       (*UserHandler).UnblockUser,
			wantStatus: http.StatusUnauthorized,
			wantBody:   "Unauthorized\n",
		},
		{
			name:       "invalid target ID",
			request:    userBlockRequest(http.MethodPut, 10, "invalid"),
			call:       (*UserHandler).BlockUser,
			wantStatus: http.StatusBadRequest,
			wantBody:   UserErrors.InvalidTargetUserID.Error() + "\n",
		},
		{
			name:       "self block",
			request:    userBlockRequest(http.MethodPut, 10, "10"),
			call:       (*UserHandler).BlockUser,
			wantStatus: http.StatusBadRequest,
			wantBody:   UserErrors.CannotBlockSelf.Error() + "\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewUserHandler(NewService(&fakeUserRepository{}, &fakeImageStorage{}))
			response := httptest.NewRecorder()

			test.call(handler, response, test.request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if response.Body.String() != test.wantBody {
				t.Fatalf("body = %q, want %q", response.Body.String(), test.wantBody)
			}
		})
	}
}

func TestUserBlockHandlersReturnNoContent(t *testing.T) {
	blockCalls := 0
	unblockCalls := 0
	repository := &fakeUserRepository{
		blockUserFn: func(_ context.Context, blockerID uint, blockedUserID uint) error {
			blockCalls++
			if blockerID != 10 || blockedUserID != 20 {
				t.Fatalf("BlockUser() arguments = (%d, %d), want (10, 20)", blockerID, blockedUserID)
			}
			return nil
		},
		unblockUserFn: func(_ context.Context, blockerID uint, blockedUserID uint) error {
			unblockCalls++
			if blockerID != 10 || blockedUserID != 20 {
				t.Fatalf("UnblockUser() arguments = (%d, %d), want (10, 20)", blockerID, blockedUserID)
			}
			return nil
		},
	}
	handler := NewUserHandler(NewService(repository, &fakeImageStorage{}))

	blockResponse := httptest.NewRecorder()
	handler.BlockUser(blockResponse, userBlockRequest(http.MethodPut, 10, "20"))
	if blockResponse.Code != http.StatusNoContent {
		t.Fatalf("BlockUser status = %d, want %d", blockResponse.Code, http.StatusNoContent)
	}
	if blockCalls != 1 {
		t.Fatalf("BlockUser calls = %d, want 1", blockCalls)
	}

	unblockResponse := httptest.NewRecorder()
	handler.UnblockUser(unblockResponse, userBlockRequest(http.MethodDelete, 10, "20"))
	if unblockResponse.Code != http.StatusNoContent {
		t.Fatalf("UnblockUser status = %d, want %d", unblockResponse.Code, http.StatusNoContent)
	}
	if unblockCalls != 1 {
		t.Fatalf("UnblockUser calls = %d, want 1", unblockCalls)
	}
}

func TestUserBlockHandlerMapsServiceErrors(t *testing.T) {
	tests := []struct {
		name       string
		serviceErr error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "current user missing",
			serviceErr: UserErrors.UserNotFound,
			wantStatus: http.StatusNotFound,
			wantBody:   "User not found\n",
		},
		{
			name:       "target user missing",
			serviceErr: UserErrors.TargetUserNotFound,
			wantStatus: http.StatusNotFound,
			wantBody:   "Target user not found\n",
		},
		{
			name:       "unexpected error",
			serviceErr: errors.New("database unavailable"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Failed to change user block\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeUserRepository{
				blockUserFn: func(context.Context, uint, uint) error {
					return test.serviceErr
				},
			}
			handler := NewUserHandler(NewService(repository, &fakeImageStorage{}))
			response := httptest.NewRecorder()

			handler.BlockUser(response, userBlockRequest(http.MethodPut, 10, "20"))

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if response.Body.String() != test.wantBody {
				t.Fatalf("body = %q, want %q", response.Body.String(), test.wantBody)
			}
		})
	}
}
