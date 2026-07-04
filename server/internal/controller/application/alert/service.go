package alert

type Service struct {
	repo               Repository
	projectAccess      ProjectAccess
	events             EventRecorder
	notificationTester NotificationTester
}

func NewService(repo Repository, projectAccess ProjectAccess, events EventRecorder, notificationTester NotificationTester) *Service {
	return &Service{repo: repo, projectAccess: projectAccess, events: events, notificationTester: notificationTester}
}
