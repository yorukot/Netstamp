package label

import (
	"context"

	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type Repository interface {
	ListLabels(ctx context.Context, projectID string) ([]domainlabel.Label, error)
	GetLabel(ctx context.Context, projectID string, labelID string) (domainlabel.Label, error)
	CreateLabel(ctx context.Context, input domainlabel.CreateLabelStorageInput) (domainlabel.Label, error)
	UpdateLabel(ctx context.Context, input domainlabel.UpdateLabelStorageInput) (domainlabel.Label, error)
	SoftDeleteLabel(ctx context.Context, projectID string, labelID string) error
	GetActiveLabelsByIDsForProject(ctx context.Context, projectID string, labelIDs []string) ([]domainlabel.Label, error)
}

type ProjectAccess interface {
	GetProjectForUser(ctx context.Context, projectRef string, userID string) (domainproject.Project, error)
	GetMemberRole(ctx context.Context, projectID string, userID string) (domainproject.Role, error)
}
