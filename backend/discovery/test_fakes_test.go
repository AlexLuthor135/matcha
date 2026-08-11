package discovery

import (
	"backend/models"
	"context"
)

type fakeUserRepository struct {
	Repository
	getProfileFn            func(context.Context, uint) (models.User, error)
	listProfileCandidatesFn func(context.Context, uint, string, string, bool) ([]models.User, error)
}

func (r *fakeUserRepository) GetDiscoveryProfile(ctx context.Context, id uint) (models.User, error) {
	return r.getProfileFn(ctx, id)
}
func (r *fakeUserRepository) ListProfileCandidates(ctx context.Context, id uint, preferred string, own string, exclude bool) ([]models.User, error) {
	return r.listProfileCandidatesFn(ctx, id, preferred, own, exclude)
}
