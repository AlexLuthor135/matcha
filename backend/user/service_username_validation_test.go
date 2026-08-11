package user

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestIsValidUserName(t *testing.T) {
	tests := []struct {
		name     string
		userName string
		want     bool
	}{
		{name: "minimum length", userName: "abc", want: true},
		{name: "maximum ASCII length", userName: strings.Repeat("a", 30), want: true},
		{name: "maximum Unicode length", userName: strings.Repeat("я", 30), want: true},
		{name: "allowed punctuation", userName: "alex.wolf-42_test", want: true},
		{name: "Unicode letters", userName: "алекс42", want: true},
		{name: "too short", userName: "ab", want: false},
		{name: "too long", userName: strings.Repeat("a", 31), want: false},
		{name: "too many Unicode characters", userName: strings.Repeat("я", 31), want: false},
		{name: "email separator", userName: "alex@example.com", want: false},
		{name: "space", userName: "alex wolf", want: false},
		{name: "slash", userName: "alex/wolf", want: false},
		{name: "emoji", userName: "alex🙂", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isValidUserName(test.userName); got != test.want {
				t.Fatalf("isValidUserName(%q) = %v, want %v", test.userName, got, test.want)
			}
		})
	}
}

func TestServiceRegisterRejectsInvalidUserName(t *testing.T) {
	_, err := NewService(&fakeUserRepository{}, &fakeImageStorage{}).Register(
		context.Background(),
		RegisterInput{
			UserName:  "user@example.com",
			FirstName: "Test",
			LastName:  "User",
			Email:     "user@example.com",
			Password:  "Valid1!Password",
		},
	)
	if !errors.Is(err, UserErrors.InvalidUserName) {
		t.Fatalf("Register() error = %v, want %v", err, UserErrors.InvalidUserName)
	}
}

func TestServiceUpdateUserRejectsInvalidUserName(t *testing.T) {
	invalidUserName := "user@example.com"
	_, err := NewService(&fakeUserRepository{}, &fakeImageStorage{}).UpdateUser(context.Background(), 17, UserUpdateInput{UserName: &invalidUserName})
	if !errors.Is(err, UserErrors.InvalidUserName) {
		t.Fatalf("UpdateUser() error = %v, want %v", err, UserErrors.InvalidUserName)
	}
}
