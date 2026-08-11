package alert

import (
	"context"

	apptx "github.com/yorukot/netstamp/internal/controller/application/tx"
)

type Service struct {
	repo               Repository
	projectAccess      ProjectAccess
	events             EventRecorder
	notificationTester NotificationTester
	incidents          IncidentReconciler
	tx                 apptx.Transactor
}

func NewService(repo Repository, projectAccess ProjectAccess, events EventRecorder, notificationTester NotificationTester) *Service {
	return &Service{repo: repo, projectAccess: projectAccess, events: events, notificationTester: notificationTester, tx: apptx.NoopTransactor{}}
}

func (s *Service) ConfigureIncidentReconciler(reconciler IncidentReconciler, transactor apptx.Transactor) {
	s.incidents = reconciler
	if transactor != nil {
		s.tx = transactor
	}
}

func (s *Service) reconcileStoppedEvaluations(ctx context.Context, projectID string) error {
	if s.incidents == nil {
		return nil
	}
	return s.incidents.ReconcileStoppedEvaluations(ctx, projectID)
}
