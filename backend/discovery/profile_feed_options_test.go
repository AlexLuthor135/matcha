package discovery

import (
	"errors"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAgeAtUsesBirthMonthAndDay(t *testing.T) {
	now := time.Now().UTC()
	birthDate := now.AddDate(-25, 0, 1)

	want := now.Year() - birthDate.Year()
	birthdayThisYear := time.Date(
		now.Year(),
		birthDate.Month(),
		birthDate.Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	)
	if now.Before(birthdayThisYear) {
		want--
	}

	if got := ageAt(birthDate); got != want {
		t.Fatalf("ageAt() = %d, want %d", got, want)
	}
}

func TestParseProfileFeedOptionsRejectsInvalidLimit(t *testing.T) {
	request := httptest.NewRequest("GET", "/profiles/feed?limit=wrong", nil)

	_, err := parseProfileFeedOptions(request)
	if !errors.Is(err, DiscoveryErrors.InvalidProfileFeedLimit) {
		t.Fatalf(
			"parseProfileFeedOptions() error = %v, want %v",
			err,
			DiscoveryErrors.InvalidProfileFeedLimit,
		)
	}
}

func TestProfileFeedOptionsRejectsUnknownSort(t *testing.T) {
	options := ProfileFeedOptions{
		Limit: 20,
		Sort:  ProfileFeedSort("unknown"),
	}
	options.Normalize()

	err := options.Validate()
	if !errors.Is(err, DiscoveryErrors.InvalidProfileFeedSort) {
		t.Fatalf(
			"ProfileFeedOptions.Validate() error = %v, want %v",
			err,
			DiscoveryErrors.InvalidProfileFeedSort,
		)
	}
}

func TestParseProfileFeedOptionsParsesFameRange(t *testing.T) {
	request := httptest.NewRequest(
		"GET",
		"/profiles/search?min_fame=2&max_fame=10",
		nil,
	)

	options, err := parseProfileFeedOptions(request)
	if err != nil {
		t.Fatalf("parseProfileFeedOptions() unexpected error: %v", err)
	}
	if options.MinFame == nil || *options.MinFame != 2 {
		t.Fatalf("MinFame = %v, want 2", options.MinFame)
	}
	if options.MaxFame == nil || *options.MaxFame != 10 {
		t.Fatalf("MaxFame = %v, want 10", options.MaxFame)
	}
}

func TestProfileFeedOptionsRejectsInvertedFameRange(t *testing.T) {
	minFame := int64(10)
	maxFame := int64(2)
	options := ProfileFeedOptions{
		Limit:   20,
		MinFame: &minFame,
		MaxFame: &maxFame,
	}
	options.Normalize()

	err := options.Validate()
	if !errors.Is(err, DiscoveryErrors.InvalidProfileFeedFilter) {
		t.Fatalf(
			"ProfileFeedOptions.Validate() error = %v, want %v",
			err,
			DiscoveryErrors.InvalidProfileFeedFilter,
		)
	}
}
