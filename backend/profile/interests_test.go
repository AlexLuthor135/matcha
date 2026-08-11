package profile

import (
	"errors"
	"reflect"
	"testing"
)

func TestNormalizeInterests(t *testing.T) {
	tests := []struct {
		name      string
		interests []string
		want      []string
		wantErr   error
	}{
		{
			name:      "normalizes case whitespace and hashtag",
			interests: []string{"  #MUSIC  ", " travel ", "#Books"},
			want:      []string{"music", "travel", "books"},
		},
		{
			name:      "removes duplicates after normalization",
			interests: []string{"music", "#MUSIC", " Music ", "travel", "#travel"},
			want:      []string{"music", "travel"},
		},
		{
			name:      "removes blank values",
			interests: []string{"", "   ", "music"},
			want:      []string{"music"},
		},
		{
			name:      "returns an empty non-nil slice for blank input",
			interests: []string{"", "   "},
			want:      []string{},
		},
		{
			name:      "rejects unsupported tag",
			interests: []string{"music", "unknown-tag"},
			wantErr:   ProfileErrors.InvalidInterestTag,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := normalizeInterests(test.interests)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("normalizeInterests() error = %v, want %v", err, test.wantErr)
			}
			if test.wantErr != nil {
				return
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("normalizeInterests() = %#v, want %#v", got, test.want)
			}
			if got == nil {
				t.Fatal("normalizeInterests() returned nil slice, want non-nil slice")
			}
		})
	}
}

func TestServiceListInterestTagsReturnsIndependentCopy(t *testing.T) {
	service := NewService(&fakeUserRepository{}, &fakeImageStorage{})
	first := service.ListInterestTags()
	second := service.ListInterestTags()

	if len(first) == 0 {
		t.Fatal("ListInterestTags() returned empty catalog")
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("ListInterestTags() returned inconsistent catalogs: %v and %v", first, second)
	}

	originalFirstTag := second[0]
	first[0] = "modified"
	third := service.ListInterestTags()
	if third[0] != originalFirstTag {
		t.Fatalf("modifying returned catalog changed service catalog: first tag = %q, want %q", third[0], originalFirstTag)
	}
}
