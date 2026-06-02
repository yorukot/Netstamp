package ping

import resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"

type Service struct {
	series        SeriesRepository
	projectAccess resultshared.ProjectAccess
}

func NewService(series SeriesRepository, projectAccess resultshared.ProjectAccess) *Service {
	return &Service{
		series:        series,
		projectAccess: projectAccess,
	}
}
