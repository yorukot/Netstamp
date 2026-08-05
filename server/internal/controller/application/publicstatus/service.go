package publicstatus

import apptx "github.com/yorukot/netstamp/internal/controller/application/tx"

type Service struct {
	repo          Repository
	projectAccess ProjectAccess
	events        EventRecorder
	pings         PingSeriesRepository
	tcps          TCPSeriesRepository
	httpResults   HTTPSeriesRepository
	snapshots     *publicSnapshotCache
	transactor    apptx.Transactor
}

func (s *Service) ConfigureHTTP(repo HTTPSeriesRepository) { s.httpResults = repo }

func (s *Service) ConfigureTransactor(transactor apptx.Transactor) {
	if transactor != nil {
		s.transactor = transactor
	}
}

func NewService(repo Repository, projectAccess ProjectAccess, events EventRecorder, pings PingSeriesRepository, tcps TCPSeriesRepository) *Service {
	return &Service{
		repo:          repo,
		projectAccess: projectAccess,
		events:        events,
		pings:         pings,
		tcps:          tcps,
		snapshots:     newPublicSnapshotCache(publicSnapshotTTL),
		transactor:    apptx.NoopTransactor{},
	}
}
