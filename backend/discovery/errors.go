package discovery

import "errors"

var DiscoveryErrors = struct {
	UserNotFound             error
	InvalidProfileFeedLimit  error
	InvalidProfileFeedFilter error
	InvalidProfileFeedSort   error
	InvalidLocation          error
}{
	UserNotFound:             errors.New("user not found"),
	InvalidProfileFeedLimit:  errors.New("profile feed limit must be a positive integer"),
	InvalidProfileFeedFilter: errors.New("invalid profile feed filter"),
	InvalidProfileFeedSort:   errors.New("invalid profile feed sort"),
	InvalidLocation:          errors.New("user location is invalid"),
}
