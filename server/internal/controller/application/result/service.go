package result

type Service struct {
	pings         PingSeriesRepository
	traceroutes   TracerouteRunsRepository
	measurements  MeasurementRepository
	projectAccess ProjectAccess
}

func NewService(pings PingSeriesRepository, traceroutes TracerouteRunsRepository, measurements MeasurementRepository, projectAccess ProjectAccess) *Service {
	return &Service{
		pings:         pings,
		traceroutes:   traceroutes,
		measurements:  measurements,
		projectAccess: projectAccess,
	}
}
