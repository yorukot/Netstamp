package publicstatus

type Service struct {
	repo          Repository
	projectAccess ProjectAccess
	pings         PingSeriesRepository
	tcps          TCPSeriesRepository
	snapshots     *publicSnapshotCache
}

func NewService(repo Repository, projectAccess ProjectAccess, pings PingSeriesRepository, tcps TCPSeriesRepository) *Service {
	return &Service{repo: repo, projectAccess: projectAccess, pings: pings, tcps: tcps, snapshots: newPublicSnapshotCache(publicSnapshotTTL)}
}
