package relationship

import "context"

func (s *Service) ReportUser(ctx context.Context, reporterID uint, reportedUserID uint) error {
	if reportedUserID == 0 {
		return RelationshipErrors.InvalidTargetUserID
	}
	if reporterID == reportedUserID {
		return RelationshipErrors.CannotReportSelf
	}
	return s.repository.ReportUser(ctx, reporterID, reportedUserID)
}
