package shared

import (
	"context"

	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type ProjectAccess interface {
	GetProjectForUser(ctx context.Context, projectRef, userID string) (domainproject.Project, error)
}
