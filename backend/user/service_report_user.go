package user

import "context"

func (s *Service) ReportUser(ctx context.Context, reporterID uint, reportedUserID uint) error {
	if reportedUserID == 0 {
		return UserErrors.InvalidTargetUserID
	}
	if reporterID == reportedUserID {
		return UserErrors.CannotReportSelf
	}
	return s.repository.ReportUser(ctx, reporterID, reportedUserID)
}
