package alert

type Service struct {
	repo               Repository
	projectAccess      ProjectAccess
	notificationTester NotificationTester
}

func NewService(repo Repository, projectAccess ProjectAccess, notificationTester NotificationTester) *Service {
	return &Service{repo: repo, projectAccess: projectAccess, notificationTester: notificationTester}
}
