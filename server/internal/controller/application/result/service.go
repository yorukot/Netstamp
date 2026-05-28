package result

type Service struct {
	pings         PingSeriesRepository
	tcps          TCPInsightRepository
	traceroutes   TracerouteRunsRepository
	measurements  MeasurementRepository
	projectAccess ProjectAccess
}

func NewService(pings PingSeriesRepository, tcps TCPInsightRepository, traceroutes TracerouteRunsRepository, measurements MeasurementRepository, projectAccess ProjectAccess) *Service {
	return &Service{
		pings:         pings,
		tcps:          tcps,
		traceroutes:   traceroutes,
		measurements:  measurements,
		projectAccess: projectAccess,
	}
}
