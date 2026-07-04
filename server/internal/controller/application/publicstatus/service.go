package publicstatus

type Service struct {
	repo          Repository
	projectAccess ProjectAccess
	events        EventRecorder
	pings         PingSeriesRepository
	tcps          TCPSeriesRepository
	snapshots     *publicSnapshotCache
}

func NewService(repo Repository, projectAccess ProjectAccess, events EventRecorder, pings PingSeriesRepository, tcps TCPSeriesRepository) *Service {
	return &Service{repo: repo, projectAccess: projectAccess, events: events, pings: pings, tcps: tcps, snapshots: newPublicSnapshotCache(publicSnapshotTTL)}
}
