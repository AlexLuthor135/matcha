package user

import "strings"

var supportedInterestTags = []string{
	"art",
	"books",
	"coding",
	"cooking",
	"cycling",
	"fitness",
	"gaming",
	"hiking",
	"movies",
	"music",
	"photography",
	"travel",
}

var supportedInterestTagSet = func() map[string]struct{} {
	tags := make(map[string]struct{}, len(supportedInterestTags))
	for _, tag := range supportedInterestTags {
		tags[tag] = struct{}{}
	}
	return tags
}()

func normalizeInterests(interests []string) ([]string, error) {
	normalizedInterests := make([]string, 0, len(interests))
	seenInterests := make(map[string]struct{}, len(interests))
	for _, interest := range interests {
		interest = strings.ToLower(strings.TrimSpace(interest))
		interest = strings.TrimPrefix(interest, "#")
		if interest == "" {
			continue
		}
		if _, supported := supportedInterestTagSet[interest]; !supported {
			return nil, UserErrors.InvalidInterestTag
		}
		if _, duplicate := seenInterests[interest]; duplicate {
			continue
		}
		seenInterests[interest] = struct{}{}
		normalizedInterests = append(normalizedInterests, interest)
	}
	return normalizedInterests, nil
}

func (s *Service) ListInterestTags() []string {
	tags := make([]string, len(supportedInterestTags))
	copy(tags, supportedInterestTags)
	return tags
}
