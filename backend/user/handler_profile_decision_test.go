package user

import (
	"backend/middleware"
	"backend/models"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func authenticatedUserRequest(request *http.Request, userID uint) *http.Request {
	ctx := context.WithValue(request.Context(), middleware.UserIDKey, userID)
	return request.WithContext(ctx)
}

func profileDecisionRequest(userID uint, targetUserID string, body string) *http.Request {
	request := httptest.NewRequest(
		http.MethodPut,
		"/profiles/"+targetUserID+"/decision",
		strings.NewReader(body),
	)
	request.SetPathValue("targetUserID", targetUserID)
	return authenticatedUserRequest(request, userID)
}

func TestSaveProfileDecisionHandlerRejectsInvalidRequests(t *testing.T) {
	tests := []struct {
		name         string
		request      func() *http.Request
		wantStatus   int
		wantResponse string
	}{
		{
			name: "unauthorized",
			request: func() *http.Request {
				request := httptest.NewRequest(
					http.MethodPut,
					"/profiles/20/decision",
					strings.NewReader(`{"decision":"like"}`),
				)
				request.SetPathValue("targetUserID", "20")
				return request
			},
			wantStatus:   http.StatusUnauthorized,
			wantResponse: "Unauthorized\n",
		},
		{
			name: "invalid target user ID",
			request: func() *http.Request {
				return profileDecisionRequest(10, "invalid", `{"decision":"like"}`)
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: "Invalid target user ID\n",
		},
		{
			name: "invalid JSON",
			request: func() *http.Request {
				return profileDecisionRequest(10, "20", `{"decision":`)
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: "Invalid request payload\n",
		},
		{
			name: "own profile",
			request: func() *http.Request {
				return profileDecisionRequest(10, "10", `{"decision":"like"}`)
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: UserErrors.CannotDecideOwnProfile.Error() + "\n",
		},
		{
			name: "invalid decision",
			request: func() *http.Request {
				return profileDecisionRequest(10, "20", `{"decision":"maybe"}`)
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: UserErrors.InvalidProfileDecision.Error() + "\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewUserHandler(NewService(&fakeUserRepository{}, &fakeImageStorage{}))
			response := httptest.NewRecorder()

			handler.SaveProfileDecision(response, test.request())

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if response.Body.String() != test.wantResponse {
				t.Fatalf("body = %q, want %q", response.Body.String(), test.wantResponse)
			}
		})
	}
}

func TestSaveProfileDecisionHandlerMapsRepositoryErrors(t *testing.T) {
	tests := []struct {
		name          string
		repositoryErr error
		wantStatus    int
		wantResponse  string
	}{
		{
			name:          "target user not found",
			repositoryErr: UserErrors.TargetUserNotFound,
			wantStatus:    http.StatusNotFound,
			wantResponse:  "Target user not found\n",
		},
		{
			name:          "current user not found",
			repositoryErr: UserErrors.UserNotFound,
			wantStatus:    http.StatusNotFound,
			wantResponse:  "User not found\n",
		},
		{
			name:          "unexpected database error",
			repositoryErr: errors.New("database unavailable"),
			wantStatus:    http.StatusInternalServerError,
			wantResponse:  "Failed to save profile decision\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeUserRepository{
				getAvatarURLFn: func(context.Context, uint) (string, error) {
					return "/uploads/avatars/user.jpg", nil
				},
				saveDecisionFn: func(
					context.Context,
					uint,
					uint,
					models.ProfileDecisionValue,
				) (SaveProfileDecisionResult, error) {
					return SaveProfileDecisionResult{}, test.repositoryErr
				},
			}
			handler := NewUserHandler(NewService(repository, &fakeImageStorage{}))
			response := httptest.NewRecorder()

			handler.SaveProfileDecision(
				response,
				profileDecisionRequest(10, "20", `{"decision":"like"}`),
			)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if response.Body.String() != test.wantResponse {
				t.Fatalf("body = %q, want %q", response.Body.String(), test.wantResponse)
			}
		})
	}
}

func TestSaveProfileDecisionHandlerDoesNotNotifyAfterBlockedError(t *testing.T) {
	repository := &fakeUserRepository{
		getAvatarURLFn: func(context.Context, uint) (string, error) {
			return "/uploads/avatars/user.jpg", nil
		},
		saveDecisionFn: func(
			context.Context,
			uint,
			uint,
			models.ProfileDecisionValue,
		) (SaveProfileDecisionResult, error) {
			return SaveProfileDecisionResult{}, UserErrors.TargetUserNotFound
		},
	}
	notificationCalls := 0
	notifier := &fakeUserNotifier{
		notifyLikeFn: func(context.Context, uint, uint) (models.Notification, error) {
			notificationCalls++
			return models.Notification{}, nil
		},
		notifyMatchFn: func(context.Context, uint, uint) (models.Notification, error) {
			notificationCalls++
			return models.Notification{}, nil
		},
		notifyUnlikeFn: func(context.Context, uint, uint) (models.Notification, error) {
			notificationCalls++
			return models.Notification{}, nil
		},
	}
	handler := NewUserHandler(NewService(repository, &fakeImageStorage{}))
	handler.SetUserNotifier(notifier)
	response := httptest.NewRecorder()

	handler.SaveProfileDecision(
		response,
		profileDecisionRequest(10, "20", `{"decision":"like"}`),
	)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
	if notificationCalls != 0 {
		t.Fatalf("notification calls = %d, want 0", notificationCalls)
	}
}

func TestSaveProfileDecisionHandlerReturnsSavedDecision(t *testing.T) {
	createdAt := time.Date(2026, time.August, 5, 12, 0, 0, 0, time.UTC)
	wantDecision := models.ProfileDecision{
		ID:           30,
		CreatedAt:    createdAt,
		UpdatedAt:    createdAt,
		UserID:       10,
		TargetUserID: 20,
		Decision:     models.ProfileDecisionLike,
	}
	repository := &fakeUserRepository{
		getAvatarURLFn: func(context.Context, uint) (string, error) {
			return "/uploads/avatars/user.jpg", nil
		},
		saveDecisionFn: func(
			_ context.Context,
			userID uint,
			targetUserID uint,
			decision models.ProfileDecisionValue,
		) (SaveProfileDecisionResult, error) {
			if userID != 10 || targetUserID != 20 || decision != models.ProfileDecisionLike {
				t.Fatalf(
					"SaveProfileDecision() arguments = (%d, %d, %q)",
					userID,
					targetUserID,
					decision,
				)
			}
			return SaveProfileDecisionResult{
				ProfileDecision: wantDecision,
				IsMatch:         true,
				DecisionChanged: true,
			}, nil
		},
	}
	handler := NewUserHandler(NewService(repository, &fakeImageStorage{}))
	response := httptest.NewRecorder()

	handler.SaveProfileDecision(
		response,
		profileDecisionRequest(10, "20", `{"decision":"like"}`),
	)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}

	var body SaveProfileDecisionResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.ProfileDecision != wantDecision {
		t.Fatalf("profile decision = %+v, want %+v", body.ProfileDecision, wantDecision)
	}
	if !body.IsMatch {
		t.Fatal("is_match = false, want true")
	}
}

func TestSaveProfileDecisionHandlerNotifiesLikedAndMatchedUser(t *testing.T) {
	tests := []struct {
		name           string
		likeNotifyErr  error
		matchNotifyErr error
	}{
		{
			name: "notification succeeds",
		},
		{
			name:          "like notification fails after saved decision",
			likeNotifyErr: errors.New("like notification unavailable"),
		},
		{
			name:           "match notification fails after saved decision",
			matchNotifyErr: errors.New("match notification unavailable"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeUserRepository{
				getAvatarURLFn: func(context.Context, uint) (string, error) {
					return "/uploads/avatars/user.jpg", nil
				},
				saveDecisionFn: func(
					_ context.Context,
					userID uint,
					targetUserID uint,
					decision models.ProfileDecisionValue,
				) (SaveProfileDecisionResult, error) {
					return SaveProfileDecisionResult{
						ProfileDecision: models.ProfileDecision{
							ID:           30,
							UserID:       userID,
							TargetUserID: targetUserID,
							Decision:     decision,
						},
						IsMatch:         true,
						DecisionChanged: true,
					}, nil
				},
			}
			likeNotifyCalls := 0
			matchNotifyCalls := 0
			notifier := &fakeUserNotifier{
				notifyLikeFn: func(_ context.Context, recipientID uint, likerID uint) (models.Notification, error) {
					likeNotifyCalls++
					if recipientID != 20 || likerID != 10 {
						t.Fatalf("NotifyLike() arguments = (%d, %d), want (20, 10)", recipientID, likerID)
					}
					return models.Notification{ID: 39}, test.likeNotifyErr
				},
				notifyMatchFn: func(_ context.Context, recipientID uint, matchedUserID uint) (models.Notification, error) {
					matchNotifyCalls++
					if recipientID != 20 || matchedUserID != 10 {
						t.Fatalf("NotifyMatch() arguments = (%d, %d), want (20, 10)", recipientID, matchedUserID)
					}
					return models.Notification{ID: 40}, test.matchNotifyErr
				},
			}
			handler := NewUserHandler(NewService(repository, &fakeImageStorage{}))
			handler.SetUserNotifier(notifier)
			response := httptest.NewRecorder()

			handler.SaveProfileDecision(
				response,
				profileDecisionRequest(10, "20", `{"decision":"like"}`),
			)

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
			}
			if likeNotifyCalls != 1 {
				t.Fatalf("NotifyLike() calls = %d, want 1", likeNotifyCalls)
			}
			if matchNotifyCalls != 1 {
				t.Fatalf("NotifyMatch() calls = %d, want 1", matchNotifyCalls)
			}
		})
	}
}

func TestSaveProfileDecisionHandlerNotifiesOnlyForChangedLike(t *testing.T) {
	tests := []struct {
		name            string
		decision        models.ProfileDecisionValue
		decisionChanged bool
		wantLikeCalls   int
	}{
		{
			name:            "new like",
			decision:        models.ProfileDecisionLike,
			decisionChanged: true,
			wantLikeCalls:   1,
		},
		{
			name:            "repeated like",
			decision:        models.ProfileDecisionLike,
			decisionChanged: false,
			wantLikeCalls:   0,
		},
		{
			name:            "changed dislike",
			decision:        models.ProfileDecisionDislike,
			decisionChanged: true,
			wantLikeCalls:   0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeUserRepository{
				getAvatarURLFn: func(context.Context, uint) (string, error) {
					return "/uploads/avatars/user.jpg", nil
				},
				saveDecisionFn: func(
					_ context.Context,
					userID uint,
					targetUserID uint,
					decision models.ProfileDecisionValue,
				) (SaveProfileDecisionResult, error) {
					return SaveProfileDecisionResult{
						ProfileDecision: models.ProfileDecision{
							ID:           30,
							UserID:       userID,
							TargetUserID: targetUserID,
							Decision:     decision,
						},
						DecisionChanged: test.decisionChanged,
					}, nil
				},
			}
			likeCalls := 0
			notifier := &fakeUserNotifier{
				notifyLikeFn: func(context.Context, uint, uint) (models.Notification, error) {
					likeCalls++
					return models.Notification{ID: 40}, nil
				},
			}
			handler := NewUserHandler(NewService(repository, &fakeImageStorage{}))
			handler.SetUserNotifier(notifier)
			response := httptest.NewRecorder()

			handler.SaveProfileDecision(
				response,
				profileDecisionRequest(
					10,
					"20",
					`{"decision":"`+string(test.decision)+`"}`,
				),
			)

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
			}
			if likeCalls != test.wantLikeCalls {
				t.Fatalf("NotifyLike() calls = %d, want %d", likeCalls, test.wantLikeCalls)
			}
		})
	}
}

func TestSaveProfileDecisionHandlerNotifiesOnlyWhenMatchEnded(t *testing.T) {
	tests := []struct {
		name            string
		matchEnded      bool
		notifyUnlikeErr error
		wantCalls       int
	}{
		{
			name:       "existing match ended",
			matchEnded: true,
			wantCalls:  1,
		},
		{
			name:       "ordinary dislike without previous match",
			matchEnded: false,
			wantCalls:  0,
		},
		{
			name:            "notification error does not revert decision",
			matchEnded:      true,
			notifyUnlikeErr: errors.New("unlike notification unavailable"),
			wantCalls:       1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeUserRepository{
				getAvatarURLFn: func(context.Context, uint) (string, error) {
					return "/uploads/avatars/user.jpg", nil
				},
				saveDecisionFn: func(
					_ context.Context,
					userID uint,
					targetUserID uint,
					decision models.ProfileDecisionValue,
				) (SaveProfileDecisionResult, error) {
					return SaveProfileDecisionResult{
						ProfileDecision: models.ProfileDecision{
							ID:           30,
							UserID:       userID,
							TargetUserID: targetUserID,
							Decision:     decision,
						},
						DecisionChanged: true,
						MatchEnded:      test.matchEnded,
					}, nil
				},
			}

			unlikeCalls := 0
			notifier := &fakeUserNotifier{
				notifyUnlikeFn: func(
					_ context.Context,
					recipientID uint,
					unlikerID uint,
				) (models.Notification, error) {
					unlikeCalls++
					if recipientID != 20 || unlikerID != 10 {
						t.Fatalf(
							"NotifyUnlike() arguments = (%d, %d), want (20, 10)",
							recipientID,
							unlikerID,
						)
					}
					return models.Notification{ID: 41}, test.notifyUnlikeErr
				},
			}

			handler := NewUserHandler(NewService(repository, &fakeImageStorage{}))
			handler.SetUserNotifier(notifier)
			response := httptest.NewRecorder()

			handler.SaveProfileDecision(
				response,
				profileDecisionRequest(10, "20", `{"decision":"dislike"}`),
			)

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
			}
			if unlikeCalls != test.wantCalls {
				t.Fatalf("NotifyUnlike() calls = %d, want %d", unlikeCalls, test.wantCalls)
			}
		})
	}
}

func TestSaveProfileDecisionHandlerRequiresProfilePicture(t *testing.T) {
	repository := &fakeUserRepository{
		getAvatarURLFn: func(context.Context, uint) (string, error) {
			return "", nil
		},
	}
	handler := NewUserHandler(NewService(repository, &fakeImageStorage{}))
	response := httptest.NewRecorder()

	handler.SaveProfileDecision(
		response,
		profileDecisionRequest(10, "20", `{"decision":"dislike"}`),
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if response.Body.String() != UserErrors.ProfilePictureRequired.Error()+"\n" {
		t.Fatalf(
			"body = %q, want %q",
			response.Body.String(),
			UserErrors.ProfilePictureRequired.Error()+"\n",
		)
	}
}

func TestListMatchesHandler(t *testing.T) {
	t.Run("unauthorized", func(t *testing.T) {
		handler := NewUserHandler(NewService(&fakeUserRepository{}, &fakeImageStorage{}))
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/matches", nil)

		handler.ListMatches(response, request)

		if response.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		repository := &fakeUserRepository{
			listMatchesFn: func(context.Context, uint) ([]models.Match, error) {
				return nil, errors.New("database unavailable")
			},
		}
		handler := NewUserHandler(NewService(repository, &fakeImageStorage{}))
		response := httptest.NewRecorder()
		request := authenticatedUserRequest(
			httptest.NewRequest(http.MethodGet, "/matches", nil),
			10,
		)

		handler.ListMatches(response, request)

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
		}
	})

	t.Run("empty matches", func(t *testing.T) {
		repository := &fakeUserRepository{
			listMatchesFn: func(_ context.Context, userID uint) ([]models.Match, error) {
				if userID != 10 {
					t.Fatalf("ListMatches() userID = %d, want 10", userID)
				}
				return make([]models.Match, 0), nil
			},
		}
		handler := NewUserHandler(NewService(repository, &fakeImageStorage{}))
		response := httptest.NewRecorder()
		request := authenticatedUserRequest(
			httptest.NewRequest(http.MethodGet, "/matches", nil),
			10,
		)

		handler.ListMatches(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
		}
		if response.Body.String() != "{\"matches\":[]}\n" {
			t.Fatalf("body = %q, want empty matches array", response.Body.String())
		}
	})

	t.Run("returns matches", func(t *testing.T) {
		wantMatches := []models.Match{
			{
				ID:        20,
				UserName:  "matched-user",
				FirstName: "Matched",
				LastName:  "User",
				Avatar:    "/uploads/avatars/matched-user.jpg",
			},
		}
		repository := &fakeUserRepository{
			listMatchesFn: func(context.Context, uint) ([]models.Match, error) {
				return wantMatches, nil
			},
		}
		handler := NewUserHandler(NewService(repository, &fakeImageStorage{}))
		response := httptest.NewRecorder()
		request := authenticatedUserRequest(
			httptest.NewRequest(http.MethodGet, "/matches", nil),
			10,
		)

		handler.ListMatches(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
		}
		var body ListMatchesResponse
		if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(body.Matches) != 1 || body.Matches[0] != wantMatches[0] {
			t.Fatalf("matches = %+v, want %+v", body.Matches, wantMatches)
		}
	})
}
