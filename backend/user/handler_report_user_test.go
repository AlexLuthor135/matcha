package user

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func reportUserRequest(reporterID uint, targetUserID string) *http.Request {
	request := httptest.NewRequest(
		http.MethodPut,
		"/profiles/"+targetUserID+"/report",
		nil,
	)
	request.SetPathValue("targetUserID", targetUserID)
	return authenticatedUserRequest(request, reporterID)
}

func TestReportUserHandlerRejectsInvalidRequest(t *testing.T) {
	tests := []struct {
		name       string
		request    *http.Request
		wantStatus int
		wantBody   string
	}{
		{
			name:       "unauthorized",
			request:    httptest.NewRequest(http.MethodPut, "/profiles/20/report", nil),
			wantStatus: http.StatusUnauthorized,
			wantBody:   "Unauthorized\n",
		},
		{
			name:       "invalid target ID",
			request:    reportUserRequest(10, "invalid"),
			wantStatus: http.StatusBadRequest,
			wantBody:   UserErrors.InvalidTargetUserID.Error() + "\n",
		},
		{
			name:       "self report",
			request:    reportUserRequest(10, "10"),
			wantStatus: http.StatusBadRequest,
			wantBody:   UserErrors.CannotReportSelf.Error() + "\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewUserHandler(NewService(&fakeUserRepository{}, &fakeImageStorage{}))
			response := httptest.NewRecorder()

			handler.ReportUser(response, test.request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if response.Body.String() != test.wantBody {
				t.Fatalf("body = %q, want %q", response.Body.String(), test.wantBody)
			}
		})
	}
}

func TestReportUserHandlerReturnsNoContent(t *testing.T) {
	repositoryCalls := 0
	repository := &fakeUserRepository{
		reportUserFn: func(_ context.Context, reporterID uint, reportedUserID uint) error {
			repositoryCalls++
			if reporterID != 10 || reportedUserID != 20 {
				t.Fatalf(
					"ReportUser() arguments = (%d, %d), want (10, 20)",
					reporterID,
					reportedUserID,
				)
			}
			return nil
		},
	}
	handler := NewUserHandler(NewService(repository, &fakeImageStorage{}))
	response := httptest.NewRecorder()

	handler.ReportUser(response, reportUserRequest(10, "20"))

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if response.Body.Len() != 0 {
		t.Fatalf("body = %q, want empty", response.Body.String())
	}
	if repositoryCalls != 1 {
		t.Fatalf("repository ReportUser() calls = %d, want 1", repositoryCalls)
	}
}

func TestReportUserHandlerMapsServiceErrors(t *testing.T) {
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
			wantBody:   "Failed to report user\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeUserRepository{
				reportUserFn: func(context.Context, uint, uint) error {
					return test.serviceErr
				},
			}
			handler := NewUserHandler(NewService(repository, &fakeImageStorage{}))
			response := httptest.NewRecorder()

			handler.ReportUser(response, reportUserRequest(10, "20"))

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if response.Body.String() != test.wantBody {
				t.Fatalf("body = %q, want %q", response.Body.String(), test.wantBody)
			}
		})
	}
}
