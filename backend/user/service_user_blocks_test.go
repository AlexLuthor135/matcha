package user

import (
	"context"
	"errors"
	"testing"
)

func TestServiceUserBlocksRejectInvalidInput(t *testing.T) {
	tests := []struct {
		name          string
		blockerID     uint
		blockedUserID uint
		wantErr       error
	}{
		{
			name:          "zero blocked user ID",
			blockerID:     10,
			blockedUserID: 0,
			wantErr:       UserErrors.InvalidTargetUserID,
		},
		{
			name:          "self block",
			blockerID:     10,
			blockedUserID: 10,
			wantErr:       UserErrors.CannotBlockSelf,
		},
	}

	for _, test := range tests {
		for _, operation := range []struct {
			name string
			call func(*Service) error
		}{
			{
				name: "block",
				call: func(service *Service) error {
					return service.BlockUser(context.Background(), test.blockerID, test.blockedUserID)
				},
			},
			{
				name: "unblock",
				call: func(service *Service) error {
					return service.UnblockUser(context.Background(), test.blockerID, test.blockedUserID)
				},
			},
		} {
			t.Run(operation.name+"/"+test.name, func(t *testing.T) {
				service := NewService(&fakeUserRepository{}, &fakeImageStorage{})
				err := operation.call(service)
				if !errors.Is(err, test.wantErr) {
					t.Fatalf("error = %v, want %v", err, test.wantErr)
				}
			})
		}
	}
}

func TestServiceBlockUserDelegatesToRepository(t *testing.T) {
	repositoryError := errors.New("database unavailable")
	repository := &fakeUserRepository{
		blockUserFn: func(_ context.Context, blockerID uint, blockedUserID uint) error {
			if blockerID != 10 || blockedUserID != 20 {
				t.Fatalf("BlockUser() arguments = (%d, %d), want (10, 20)", blockerID, blockedUserID)
			}
			return repositoryError
		},
	}

	err := NewService(repository, &fakeImageStorage{}).BlockUser(context.Background(), 10, 20)
	if !errors.Is(err, repositoryError) {
		t.Fatalf("BlockUser() error = %v, want %v", err, repositoryError)
	}
}

func TestServiceUnblockUserDelegatesToRepository(t *testing.T) {
	repositoryError := errors.New("database unavailable")
	repository := &fakeUserRepository{
		unblockUserFn: func(_ context.Context, blockerID uint, blockedUserID uint) error {
			if blockerID != 10 || blockedUserID != 20 {
				t.Fatalf("UnblockUser() arguments = (%d, %d), want (10, 20)", blockerID, blockedUserID)
			}
			return repositoryError
		},
	}

	err := NewService(repository, &fakeImageStorage{}).UnblockUser(context.Background(), 10, 20)
	if !errors.Is(err, repositoryError) {
		t.Fatalf("UnblockUser() error = %v, want %v", err, repositoryError)
	}
}
