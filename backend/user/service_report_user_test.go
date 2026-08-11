package user

import (
	"context"
	"errors"
	"testing"
)

func TestServiceReportUserRejectsInvalidTarget(t *testing.T) {
	tests := []struct {
		name           string
		reporterID     uint
		reportedUserID uint
		wantErr        error
	}{
		{
			name:           "zero target ID",
			reporterID:     10,
			reportedUserID: 0,
			wantErr:        UserErrors.InvalidTargetUserID,
		},
		{
			name:           "self report",
			reporterID:     10,
			reportedUserID: 10,
			wantErr:        UserErrors.CannotReportSelf,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(&fakeUserRepository{}, &fakeImageStorage{})

			err := service.ReportUser(
				context.Background(),
				test.reporterID,
				test.reportedUserID,
			)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("ReportUser() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestServiceReportUserDelegatesToRepository(t *testing.T) {
	const reporterID uint = 10
	const reportedUserID uint = 20
	repositoryCalls := 0
	repository := &fakeUserRepository{
		reportUserFn: func(_ context.Context, gotReporterID uint, gotReportedUserID uint) error {
			repositoryCalls++
			if gotReporterID != reporterID || gotReportedUserID != reportedUserID {
				t.Fatalf(
					"ReportUser() arguments = (%d, %d), want (%d, %d)",
					gotReporterID,
					gotReportedUserID,
					reporterID,
					reportedUserID,
				)
			}
			return nil
		},
	}

	err := NewService(repository, &fakeImageStorage{}).ReportUser(
		context.Background(),
		reporterID,
		reportedUserID,
	)
	if err != nil {
		t.Fatalf("ReportUser() unexpected error: %v", err)
	}
	if repositoryCalls != 1 {
		t.Fatalf("repository ReportUser() calls = %d, want 1", repositoryCalls)
	}
}

func TestServiceReportUserReturnsRepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	repository := &fakeUserRepository{
		reportUserFn: func(context.Context, uint, uint) error {
			return wantErr
		},
	}

	err := NewService(repository, &fakeImageStorage{}).ReportUser(
		context.Background(),
		10,
		20,
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("ReportUser() error = %v, want %v", err, wantErr)
	}
}
