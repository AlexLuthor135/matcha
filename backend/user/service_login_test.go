package user

import (
	"backend/models"
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func loginTestPasswordHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash login test password: %v", err)
	}
	return string(hash)
}

func TestServiceLoginWithUserName(t *testing.T) {
	const password = "Valid1!Password"
	repository := &fakeUserRepository{
		getUserByUserNameFn: func(_ context.Context, userName string) (models.User, error) {
			if userName != "test-user" {
				t.Fatalf("GetUserByUserName() username = %q, want %q", userName, "test-user")
			}
			return models.User{
				ID:         17,
				Password:   loginTestPasswordHash(t, password),
				IsVerified: true,
			}, nil
		},
	}

	user, err := NewService(repository, &fakeImageStorage{}).Login(
		context.Background(),
		"  test-user  ",
		password,
	)
	if err != nil {
		t.Fatalf("Login() with username unexpected error: %v", err)
	}
	if user.ID != 17 {
		t.Fatalf("Login() user ID = %d, want 17", user.ID)
	}
}

func TestServiceLoginWithEmail(t *testing.T) {
	const password = "Valid1!Password"
	repository := &fakeUserRepository{
		getUserByEmailFn: func(_ context.Context, email string) (models.User, error) {
			if email != "user@example.com" {
				t.Fatalf("GetUserByEmail() email = %q, want %q", email, "user@example.com")
			}
			return models.User{
				ID:         18,
				Password:   loginTestPasswordHash(t, password),
				IsVerified: true,
			}, nil
		},
	}

	user, err := NewService(repository, &fakeImageStorage{}).Login(
		context.Background(),
		"  USER@EXAMPLE.COM  ",
		password,
	)
	if err != nil {
		t.Fatalf("Login() with email unexpected error: %v", err)
	}
	if user.ID != 18 {
		t.Fatalf("Login() user ID = %d, want 18", user.ID)
	}
}

func TestServiceLoginRejectsInvalidInputAndCredentials(t *testing.T) {
	tests := []struct {
		name       string
		login      string
		password   string
		repository *fakeUserRepository
		want       error
	}{
		{
			name:       "blank login",
			password:   "Valid1!Password",
			repository: &fakeUserRepository{},
			want:       UserErrors.LoginFieldsMissing,
		},
		{
			name:       "blank password",
			login:      "test-user",
			repository: &fakeUserRepository{},
			want:       UserErrors.LoginFieldsMissing,
		},
		{
			name:     "unknown username",
			login:    "missing-user",
			password: "Valid1!Password",
			repository: &fakeUserRepository{
				getUserByUserNameFn: func(context.Context, string) (models.User, error) {
					return models.User{}, UserErrors.UserNotFound
				},
			},
			want: UserErrors.InvalidCredentials,
		},
		{
			name:     "wrong password",
			login:    "test-user",
			password: "Wrong1!Password",
			repository: &fakeUserRepository{
				getUserByUserNameFn: func(context.Context, string) (models.User, error) {
					return models.User{
						Password:   loginTestPasswordHash(t, "Valid1!Password"),
						IsVerified: true,
					}, nil
				},
			},
			want: UserErrors.InvalidCredentials,
		},
		{
			name:     "unverified email",
			login:    "test-user",
			password: "Valid1!Password",
			repository: &fakeUserRepository{
				getUserByUserNameFn: func(context.Context, string) (models.User, error) {
					return models.User{
						Password:   loginTestPasswordHash(t, "Valid1!Password"),
						IsVerified: false,
					}, nil
				},
			},
			want: UserErrors.EmailNotVerified,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewService(test.repository, &fakeImageStorage{}).Login(
				context.Background(),
				test.login,
				test.password,
			)
			if !errors.Is(err, test.want) {
				t.Fatalf("Login() error = %v, want %v", err, test.want)
			}
		})
	}
}
