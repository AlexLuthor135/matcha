package user

import (
	"context"
	"errors"
	"testing"
)

func TestIsCommonPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{name: "listed password", password: "password1!", want: true},
		{name: "case insensitive", password: "QWERTY1!", want: true},
		{name: "surrounding whitespace", password: "  Welcome1!  ", want: true},
		{name: "Unicode case insensitive", password: "ПАРОЛЬ1!", want: true},
		{name: "unlisted password", password: "River7!Orchid", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isCommonPassword(test.password); got != test.want {
				t.Fatalf("isCommonPassword(%q) = %v, want %v", test.password, got, test.want)
			}
		})
	}
}

func TestIsValidPasswordRejectsBlocklistedPassword(t *testing.T) {
	if isValidPassword("Password1!") {
		t.Fatal("isValidPassword() accepted a blocklisted password")
	}
	if !isValidPassword("River7!Orchid") {
		t.Fatal("isValidPassword() rejected an unlisted password that meets the composition rules")
	}
}

func TestPasswordBlocklistAppliesToAllPasswordCreationFlows(t *testing.T) {
	tests := []struct {
		name string
		call func() error
	}{
		{
			name: "registration",
			call: func() error {
				_, err := NewService(&fakeUserRepository{}, &fakeImageStorage{}).Register(
					context.Background(),
					RegisterInput{
						UserName:  "test-user",
						FirstName: "Test",
						LastName:  "User",
						Email:     "test@example.com",
						Password:  "Password1!",
					},
				)
				return err
			},
		},
		{
			name: "authenticated password update",
			call: func() error {
				return NewService(&fakeUserRepository{}, &fakeImageStorage{}).UpdatePassword(
					context.Background(),
					17,
					"Current1!Password",
					"Password1!",
				)
			},
		},
		{
			name: "password reset",
			call: func() error {
				return NewService(&fakeUserRepository{}, &fakeImageStorage{}).ResetPassword(
					context.Background(),
					ResetPasswordInput{
						Token:       "reset-token",
						NewPassword: "Password1!",
					},
				)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.call(); !errors.Is(err, UserErrors.InvalidPassword) {
				t.Fatalf("password flow error = %v, want %v", err, UserErrors.InvalidPassword)
			}
		})
	}
}
